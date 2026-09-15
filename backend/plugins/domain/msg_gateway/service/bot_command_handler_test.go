// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"Wavelet/core/contracts"
	igodao "Wavelet/igo-lib/plugins/igo/dao"
	igoentity "Wavelet/igo-lib/plugins/igo/model/entity"
	igosvc "Wavelet/igo-lib/plugins/igo/service"
	"Wavelet/pkg/idgen"
	"Wavelet/pkg/testhelper"
	"Wavelet/plugins/domain/msg_gateway/dao"
	mgdo "Wavelet/plugins/domain/msg_gateway/model/do"
	mgentity "Wavelet/plugins/domain/msg_gateway/model/entity"
	mgservice "Wavelet/plugins/domain/msg_gateway/service"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type memoryCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMemoryCache() *memoryCache {
	return &memoryCache{data: make(map[string]string)}
}

func (m *memoryCache) Get(_ context.Context, key string, target any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, ok := m.data[key]
	if !ok {
		return contracts.ErrCacheMiss
	}
	return json.Unmarshal([]byte(raw), target)
}

func (m *memoryCache) Set(_ context.Context, key string, val any, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	m.data[key] = string(b)
	return nil
}

func (m *memoryCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *memoryCache) GetOrSet(ctx context.Context, key string, target any, ttl time.Duration, loader func() (any, error)) error {
	if err := m.Get(ctx, key, target); err == nil {
		return nil
	}
	val, err := loader()
	if err != nil {
		return err
	}
	if err := m.Set(ctx, key, val, ttl); err != nil {
		return err
	}
	return m.Get(ctx, key, target)
}

func (m *memoryCache) Invalidate(ctx context.Context, key string) error {
	return m.Delete(ctx, key)
}

type stubDBService struct {
	db *gorm.DB
}

func (s stubDBService) DB(context.Context) *gorm.DB {
	return s.db
}

func (s stubDBService) GORM() *gorm.DB {
	return s.db
}

func (s stubDBService) Named(string) *gorm.DB {
	return s.db
}

type dummyChannel struct {
	sent []string
	mu   sync.Mutex
}

func (d *dummyChannel) Type() string { return "telegram" }
func (d *dummyChannel) Connect(context.Context) error {
	return nil
}
func (d *dummyChannel) Disconnect(context.Context) error {
	return nil
}
func (d *dummyChannel) Capabilities() mgdo.Capability {
	return mgdo.Capability{Text: true}
}
func (d *dummyChannel) Send(_ context.Context, _ mgdo.Recipient, msg mgdo.OutboundMessage) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sent = append(d.sent, msg.Text)
	return nil
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = r.target.Scheme
	req2.URL.Host = r.target.Host
	if r.base == nil {
		return http.DefaultTransport.RoundTrip(req2)
	}
	return r.base.RoundTrip(req2)
}

func rewriteHost(targetURL string) http.RoundTripper {
	u, _ := url.Parse(targetURL)
	return &rewriteTransport{target: u}
}

func TestBotCommandHandler(t *testing.T) {
	_ = idgen.Init(1)
	db, _, cleanup := testhelper.SetupTestEnvironment(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(
		&mgentity.MessageChannel{},
		&mgentity.MessageBinding{},
		&mgentity.MessagePairingCode{},
		&igoentity.PipelineConfig{},
		&igoentity.ProtocolOverride{},
		&igoentity.Settings{},
	))

	dao.SetDBServiceForTest(stubDBService{db: db})
	igodao.SetDBService(stubDBService{db: db})
	t.Cleanup(func() {
		dao.SetDBServiceForTest(nil)
		igodao.SetDBService(nil)
	})

	cache := newMemoryCache()
	dao.SetCacheService(cache)

	dummy := &dummyChannel{}
	mgservice.Register("telegram", func(cfg mgdo.ChannelConfig, onInbound mgservice.Handler) (mgservice.Channel, error) {
		return dummy, nil
	})

	ctx := context.Background()

	mgservice.SetCredentialSecret("test-secret-key-1234567890123456")
	encCreds, err := mgservice.EncryptCredentials(map[string]string{"bot_token": "test_token"})
	require.NoError(t, err)

	// 1. Create a channel
	ch := &mgentity.MessageChannel{
		ID:          1,
		Type:        "telegram",
		Name:        "测试Bot",
		Enabled:     true,
		Credentials: encCreds,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, dao.CreateMessageChannel(ctx, ch))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "graphql") {
			_, _ = w.Write([]byte(`{"data":{"userAuth":{"reserve":{"libs":[
				{"lib_id":1,"lib_name":"一楼","lib_floor":"1","is_open":true,"lib_layout":{"seats":[{"key":"S-01","name":"01号","type":1,"status":false}]}},
				{"lib_id":2,"lib_name":"二楼","lib_floor":"2","is_open":true,"lib_layout":{"seats":[{"key":"S-02","name":"02号","type":1,"status":false}]}}
			],"reserueSeat":true}}}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	igoSvc := igosvc.New()
	igoSvc.SetHTTPClient(&http.Client{Timeout: 5 * time.Second, Transport: rewriteHost(srv.URL)})
	handler := mgservice.NewBotCommandHandler(igoSvc)

	// Test 1: Unbound user -> reply pairing code
	unboundMsg := mgdo.InboundMessage{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		ChatID:         "tg_user_999",
		Text:           "/start",
	}
	require.NoError(t, handler.HandleInbound(ctx, unboundMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "专属配对码")

	// Bind user 100 to tg_user_999
	require.NoError(t, dao.CreateMessageBinding(ctx, &mgentity.MessageBinding{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		UserID:         100,
		CreatedAt:      time.Now(),
	}))

	// Test 2: /help command
	dummy.sent = nil
	helpMsg := mgdo.InboundMessage{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		ChatID:         "tg_user_999",
		Text:           "/help",
	}
	require.NoError(t, handler.HandleInbound(ctx, helpMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "/show")
	assert.Contains(t, dummy.sent[0], "/run")

	// Test 3: /show command with no configs
	dummy.sent = nil
	showMsg := mgdo.InboundMessage{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		ChatID:         "tg_user_999",
		Text:           "/show",
	}
	require.NoError(t, handler.HandleInbound(ctx, showMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "暂无配置任何一条龙自动化卡片")

	// Create pipeline config for user 100
	cfg := &igoentity.PipelineConfig{
		ID:          "myseat01",
		UserID:      100,
		Name:        "专座",
		Cookie:      "Authorization=valid-cookie",
		LibraryID:   1,
		LibraryName: "一楼",
		Floor:       "1",
		SeatKey:     "S-01",
		SeatName:    "01号",
		AutoCheckin: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, igodao.CreatePipelineConfig(ctx, cfg))

	// Test 4: /show command with 1 config
	dummy.sent = nil
	require.NoError(t, handler.HandleInbound(ctx, showMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "[myseat01] 专座")

	// Test 5: /run myseat01 -> executes directly
	dummy.sent = nil
	runMsg := mgdo.InboundMessage{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		ChatID:         "tg_user_999",
		Text:           "/run myseat01",
	}
	require.NoError(t, handler.HandleInbound(ctx, runMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "一条龙自动化执行成功")

	// Test 6: /run with expired cookie -> returns OAuth link and enters waiting state
	_ = igodao.UpdatePipelineCookie(ctx, "myseat01", "", nil) // empty cookie
	dummy.sent = nil
	require.NoError(t, handler.HandleInbound(ctx, runMsg))
	require.Len(t, dummy.sent, 1)
	assert.Contains(t, dummy.sent[0], "open.weixin.qq.com")
	assert.Contains(t, dummy.sent[0], "TraceInt 登录凭据已过期")

	// User replies with new Cookie
	replyMsg := mgdo.InboundMessage{
		ChannelID:      1,
		PlatformUserID: "tg_user_999",
		ChatID:         "tg_user_999",
		Text:           "Authorization=new-cookie; SERVERID=123",
	}
	dummy.sent = nil
	require.NoError(t, handler.HandleInbound(ctx, replyMsg))
	// Should have replied with "登录凭据已更新" and execution result
	assert.True(t, len(dummy.sent) >= 1)
	combined := strings.Join(dummy.sent, "\n")
	assert.Contains(t, combined, "一条龙自动化执行成功")
}
