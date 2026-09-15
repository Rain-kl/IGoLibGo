// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	"Wavelet/plugins/domain/msg_gateway/bot"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
	"Wavelet/plugins/domain/msg_gateway/model/entity"
	"Wavelet/plugins/domain/msg_gateway/service"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestParseBotCommand(t *testing.T) {
	tests := []struct {
		in    string
		isCmd bool
		name  string
		args  []string
	}{
		{"/help", true, "help", nil},
		{"/HELP", true, "help", nil},
		{"/run@MyBot id1", true, "run", []string{"id1"}},
		{"/run  myseat01", true, "run", []string{"myseat01"}},
		{"hello", false, "", nil},
		{"", false, "", nil},
		{"  ", false, "", nil},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.in), func(t *testing.T) {
			name, args, isCmd := service.ParseBotCommand(tt.in)
			assert.Equal(t, tt.isCmd, isCmd)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.args, args)
		})
	}
}

type replyCmd struct {
	name, desc, usage string
	aliases           []string
	reply             string
	handled           int
	lastUserID        uint64
	lastArgs          []string
	lastText          string
	err               error
	panic             any
}

func (c *replyCmd) Name() string        { return c.name }
func (c *replyCmd) Aliases() []string   { return c.aliases }
func (c *replyCmd) Description() string { return c.desc }
func (c *replyCmd) Usage() string       { return c.usage }
func (c *replyCmd) Handle(_ context.Context, req contracts.BotCommandRequest) error {
	c.handled++
	c.lastUserID = req.UserID()
	c.lastArgs = req.Args()
	c.lastText = req.Text()
	if c.panic != nil {
		panic(c.panic)
	}
	if c.err != nil {
		return c.err
	}
	if c.reply != "" {
		return req.Reply(c.reply)
	}
	return nil
}

func helpReplyCmd() *replyCmd {
	return &replyCmd{
		name:    consts.CommandHelp,
		desc:    "show help",
		usage:   "/help",
		aliases: []string{consts.CommandStart},
		reply:   "HELP_OK",
	}
}

type sendCapture struct {
	channelID uint64
	to        do.Recipient
	text      string
	calls     int
}

func (s *sendCapture) fn(_ context.Context, channelID uint64, to do.Recipient, text string) error {
	s.channelID = channelID
	s.to = to
	s.text = text
	s.calls++
	return nil
}

func setupInboundTestDB(t *testing.T) {
	t.Helper()
	_ = idgen.Init(1)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "inbound_test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}, &entity.MessagePairingCode{}))
	dao.SetDBServiceForTest(&testDBService{db: db})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })
}

func inboundMsg(platformUserID, text string) do.InboundMessage {
	return do.InboundMessage{
		ChannelID:      10,
		ChannelType:    consts.ChannelTypeTelegram,
		PlatformUserID: platformUserID,
		ChatID:         "999",
		MessageID:      "m1",
		Text:           text,
	}
}

func bindTestUser(t *testing.T, ctx context.Context, platformUserID string, userID uint64) {
	t.Helper()
	require.NoError(t, dao.CreateMessageBinding(ctx, &entity.MessageBinding{
		UserID:         userID,
		ChannelID:      10,
		PlatformUserID: platformUserID,
	}))
}

func TestBotRegistry_Dispatch(t *testing.T) {
	ctx := context.Background()
	const (
		unboundUser = "tg_unbound"
		boundUser   = "tg_bound"
		boundUID    = uint64(42)
	)

	t.Run("unbound /help allowUnbound reaches Handle without pairing", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		help := helpReplyCmd()
		require.NoError(t, reg.RegisterBuiltin(help, true))

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(unboundUser, "/help"), sent.fn)
		require.NoError(t, err)
		assert.Equal(t, "HELP_OK", sent.text)
		assert.NotContains(t, sent.text, "配对码")
		assert.Equal(t, 1, help.handled)
		assert.Equal(t, uint64(0), help.lastUserID)
		assert.Equal(t, uint64(10), sent.channelID)
		assert.Equal(t, "999", sent.to.ChatID)
		assert.Equal(t, unboundUser, sent.to.PlatformUserID)
	})

	t.Run("unbound hello replies pairing code", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		require.NoError(t, reg.RegisterBuiltin(helpReplyCmd(), true))

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(unboundUser, "hello"), sent.fn)
		require.NoError(t, err)
		assert.Contains(t, sent.text, "配对码")
		assert.NotContains(t, sent.text, "未知命令")
	})

	t.Run("unbound unknown /foo replies pairing code", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		require.NoError(t, reg.RegisterBuiltin(helpReplyCmd(), true))

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(unboundUser, "/foo"), sent.fn)
		require.NoError(t, err)
		assert.Contains(t, sent.text, "配对码")
		assert.NotContains(t, sent.text, "未知命令")
	})

	t.Run("bound /help reaches Handle", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		help := helpReplyCmd()
		require.NoError(t, reg.RegisterBuiltin(help, true))
		bindTestUser(t, ctx, boundUser, boundUID)

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(boundUser, "/help"), sent.fn)
		require.NoError(t, err)
		assert.Equal(t, "HELP_OK", sent.text)
		assert.Equal(t, 1, help.handled)
		assert.Equal(t, boundUID, help.lastUserID)
	})

	t.Run("bound unknown /foo replies with /help", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		require.NoError(t, reg.RegisterBuiltin(helpReplyCmd(), true))
		bindTestUser(t, ctx, boundUser, boundUID)

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(boundUser, "/foo"), sent.fn)
		require.NoError(t, err)
		assert.Contains(t, sent.text, "/help")
		assert.NotContains(t, sent.text, "配对码")
	})

	t.Run("bound hello non-command does not send", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		require.NoError(t, reg.RegisterBuiltin(helpReplyCmd(), true))
		bindTestUser(t, ctx, boundUser, boundUID)

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(boundUser, "hello"), sent.fn)
		require.NoError(t, err)
		assert.Equal(t, "", sent.text)
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("bound /start alias uses help Handle", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		help := helpReplyCmd()
		require.NoError(t, reg.RegisterBuiltin(help, true))
		bindTestUser(t, ctx, boundUser, boundUID)

		sentHelp := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(boundUser, "/help"), sentHelp.fn))
		sentStart := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(boundUser, "/start"), sentStart.fn))
		assert.Equal(t, sentHelp.text, sentStart.text)
		assert.Equal(t, "HELP_OK", sentStart.text)
		assert.Equal(t, 2, help.handled)
	})

	t.Run("handle panic is recovered and not replied", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		cmd := &replyCmd{name: "boom", desc: "d", usage: "/boom", panic: "boom"}
		require.NoError(t, reg.Register(cmd))
		bindTestUser(t, ctx, boundUser, boundUID)

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(boundUser, "/boom"), sent.fn)
		require.NoError(t, err)
		assert.Equal(t, "", sent.text)
		assert.Equal(t, 1, cmd.handled)
	})

	t.Run("handle error is logged and not replied", func(t *testing.T) {
		setupInboundTestDB(t)
		reg := service.NewBotRegistry()
		cmd := &replyCmd{name: "fail", desc: "d", usage: "/fail", err: errors.New("secret internal")}
		require.NoError(t, reg.Register(cmd))
		bindTestUser(t, ctx, boundUser, boundUID)

		sent := &sendCapture{}
		err := reg.Dispatch(ctx, inboundMsg(boundUser, "/fail"), sent.fn)
		require.NoError(t, err)
		assert.Equal(t, "", sent.text)
		assert.NotContains(t, sent.text, "secret internal")
	})
}

func TestRegisterBuiltins_DispatchMeAndHelp(t *testing.T) {
	setupInboundTestDB(t)
	ctx := context.Background()
	reg := service.NewBotRegistry()
	require.NoError(t, bot.RegisterBuiltins(reg))

	unbound := &sendCapture{}
	require.NoError(t, reg.Dispatch(ctx, inboundMsg("tg_unbound", "/me"), unbound.fn))
	assert.Contains(t, unbound.text, "配对码")

	bindTestUser(t, ctx, "tg_bound", 42)
	bound := &sendCapture{}
	require.NoError(t, reg.Dispatch(ctx, inboundMsg("tg_bound", "/me"), bound.fn))
	assert.Contains(t, bound.text, "42")
	assert.NotContains(t, bound.text, "配对码")

	sentHelp := &sendCapture{}
	require.NoError(t, reg.Dispatch(ctx, inboundMsg("tg_bound", "/help"), sentHelp.fn))
	sentStart := &sendCapture{}
	require.NoError(t, reg.Dispatch(ctx, inboundMsg("tg_bound", "/start"), sentStart.fn))
	assert.Equal(t, sentHelp.text, sentStart.text)
	assert.Contains(t, sentHelp.text, "/me")
	assert.Contains(t, sentHelp.text, "/start, /help")
}

const convBoundUser = "tg_conv"

func TestBotRegistry_Conversation(t *testing.T) {
	const boundUID = uint64(7)

	t.Run("Begin then cookie-text goes to OnMessage not other commands", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		begin, show, _ := registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.NoError(t, begin.beginErr)
		assert.Equal(t, 1, begin.handled)

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), sent.fn))
		assert.Equal(t, []string{"cookie-text"}, login.messages)
		assert.Empty(t, login.cancels)
		assert.Equal(t, 0, show.handled)
		assert.Equal(t, 0, sent.calls)
		assert.NotContains(t, sent.text, "未知命令")
	})

	t.Run("cancel calls OnCancel then clears occupancy", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		_, _, cancel := registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/cancel"), sent.fn))
		assert.Equal(t, "已取消", sent.text)
		assert.Equal(t, 1, cancel.handled)
		assert.Equal(t, []string{"/cancel"}, login.cancels)
		assert.Empty(t, login.messages)

		sent2 := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "hello"), sent2.fn))
		assert.Empty(t, login.messages)
		assert.Equal(t, 0, sent2.calls)

		sent3 := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/cancel"), sent3.fn))
		assert.Equal(t, "当前没有进行中的操作", sent3.text)
		assert.Equal(t, []string{"/cancel"}, login.cancels)
	})

	t.Run("registered /show silent-clears occupancy without OnCancel", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		_, show, _ := registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/show"), sent.fn))
		assert.Equal(t, "SHOW_OK", sent.text)
		assert.Equal(t, 1, show.handled)
		assert.Empty(t, login.cancels)
		assert.Empty(t, login.messages)

		sent2 := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), sent2.fn))
		assert.Empty(t, login.messages)
		assert.Equal(t, 0, sent2.calls)
	})

	t.Run("unregistered /nope during conversation goes to OnMessage", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		_, show, _ := registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/nope"), sent.fn))
		assert.Equal(t, []string{"/nope"}, login.messages)
		assert.Equal(t, 0, show.handled)
		assert.Empty(t, login.cancels)
		assert.NotContains(t, sent.text, "未知命令")
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("End clears occupancy so later text is dropped", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{
			name: "igo.login_auth",
			hook: func(req contracts.BotConversationRequest) error {
				return req.End()
			},
		}
		registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), (&sendCapture{}).fn))
		assert.Equal(t, []string{"cookie-text"}, login.messages)

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "later-text"), sent.fn))
		assert.Equal(t, []string{"cookie-text"}, login.messages)
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("Transition routes the next message to the new conversation", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{
			name: "igo.login_auth",
			hook: func(req contracts.BotConversationRequest) error {
				return req.Transition("igo.checkin_auth", map[string]string{"step": "checkin"})
			},
		}
		checkin := &recordingConv{name: "igo.checkin_auth"}
		registerConversationFixture(t, reg, login, checkin)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), (&sendCapture{}).fn))
		assert.Equal(t, []string{"cookie-text"}, login.messages)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "next-text"), (&sendCapture{}).fn))
		assert.Equal(t, []string{"next-text"}, checkin.messages)
		assert.Equal(t, []string{"cookie-text"}, login.messages)
	})

	t.Run("Begin unknown conversation name returns error", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		begin, _, _ := registerConversationFixture(t, reg, login)
		begin.conv = "igo.missing"
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.Error(t, begin.beginErr)
		assert.Equal(t, 1, begin.handled)

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), sent.fn))
		assert.Empty(t, login.messages)
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("cache miss is treated as no conversation", func(t *testing.T) {
		ctx, reg, cache := setupConversationTestWithCache(t)
		login := &recordingConv{name: "igo.login_auth"}
		registerConversationFixture(t, reg, login)
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.NoError(t, cache.Delete(ctx, fmt.Sprintf("msg_gateway:bot_conv:%d:%s", uint64(10), convBoundUser)))

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), sent.fn))
		assert.Empty(t, login.messages)
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("expired occupancy is treated as no conversation", func(t *testing.T) {
		ctx, reg := setupConversationTest(t)
		login := &recordingConv{name: "igo.login_auth"}
		begin, _, _ := registerConversationFixture(t, reg, login)
		begin.ttl = time.Nanosecond
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.NoError(t, begin.beginErr)
		time.Sleep(2 * time.Millisecond)

		sent := &sendCapture{}
		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "cookie-text"), sent.fn))
		assert.Empty(t, login.messages)
		assert.Equal(t, 0, sent.calls)
	})

	t.Run("Begin without cache returns error", func(t *testing.T) {
		setupInboundTestDB(t)
		service.SetCacheService(nil)
		t.Cleanup(func() { service.SetCacheService(nil) })

		reg := service.NewBotRegistry()
		login := &recordingConv{name: "igo.login_auth"}
		begin, _, _ := registerConversationFixture(t, reg, login)
		ctx := context.Background()
		bindTestUser(t, ctx, convBoundUser, boundUID)

		require.NoError(t, reg.Dispatch(ctx, inboundMsg(convBoundUser, "/auth"), (&sendCapture{}).fn))
		require.Error(t, begin.beginErr)
		assert.Contains(t, begin.beginErr.Error(), "bot conversation cache unavailable")
	})
}

func setupConversationTest(t *testing.T) (context.Context, *service.BotRegistry) {
	t.Helper()
	ctx, reg, _ := setupConversationTestWithCache(t)
	return ctx, reg
}

func setupConversationTestWithCache(t *testing.T) (context.Context, *service.BotRegistry, *memoryCache) {
	t.Helper()
	setupInboundTestDB(t)
	cache := newMemoryCache()
	service.SetCacheService(cache)
	t.Cleanup(func() { service.SetCacheService(nil) })
	return context.Background(), service.NewBotRegistry(), cache
}

func registerConversationFixture(t *testing.T, reg *service.BotRegistry, convs ...contracts.BotConversation) (*beginCmd, *replyCmd, *cancelStub) {
	t.Helper()
	begin := &beginCmd{name: "auth", conv: "igo.login_auth", state: map[string]string{"k": "v"}, ttl: 5 * time.Minute}
	show := &replyCmd{name: "show", desc: "show info", usage: "/show", reply: "SHOW_OK"}
	cancel := &cancelStub{}
	require.NoError(t, reg.Register(begin))
	require.NoError(t, reg.Register(show))
	require.NoError(t, reg.RegisterBuiltin(cancel, true))
	for _, conv := range convs {
		require.NoError(t, reg.RegisterConversation(conv))
	}
	return begin, show, cancel
}

type beginCmd struct {
	name     string
	conv     string
	state    any
	ttl      time.Duration
	handled  int
	beginErr error
}

func (c *beginCmd) Name() string        { return c.name }
func (c *beginCmd) Aliases() []string   { return nil }
func (c *beginCmd) Description() string { return "begin conversation" }
func (c *beginCmd) Usage() string       { return "/" + c.name }
func (c *beginCmd) Handle(_ context.Context, req contracts.BotCommandRequest) error {
	c.handled++
	c.beginErr = req.Begin(c.conv, c.state, c.ttl)
	return c.beginErr
}

type cancelStub struct {
	handled int
}

func (c *cancelStub) Name() string        { return consts.CommandCancel }
func (c *cancelStub) Aliases() []string   { return nil }
func (c *cancelStub) Description() string { return "cancel active conversation" }
func (c *cancelStub) Usage() string       { return "/cancel" }
func (c *cancelStub) Handle(_ context.Context, req contracts.BotCommandRequest) error {
	c.handled++
	if req.HasConversation() {
		if err := req.CancelActive(); err != nil {
			return err
		}
		return req.Reply("已取消")
	}
	return req.Reply("当前没有进行中的操作")
}

type recordingConv struct {
	name     string
	messages []string
	cancels  []string
	hook     func(contracts.BotConversationRequest) error
}

func (c *recordingConv) Name() string { return c.name }
func (c *recordingConv) OnMessage(_ context.Context, req contracts.BotConversationRequest) error {
	c.messages = append(c.messages, req.Text())
	if c.hook != nil {
		return c.hook(req)
	}
	return nil
}
func (c *recordingConv) OnCancel(_ context.Context, req contracts.BotConversationRequest) error {
	c.cancels = append(c.cancels, req.Text())
	return nil
}

type memoryEntry struct {
	raw      []byte
	expireAt time.Time
}

type memoryCache struct {
	mu   sync.Mutex
	data map[string]memoryEntry
}

func newMemoryCache() *memoryCache {
	return &memoryCache{data: make(map[string]memoryEntry)}
}

var _ contracts.CacheService = (*memoryCache)(nil)

func (m *memoryCache) Get(_ context.Context, key string, target any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.data[key]
	if !ok {
		return contracts.ErrCacheMiss
	}
	if !e.expireAt.IsZero() && !time.Now().Before(e.expireAt) {
		delete(m.data, key)
		return contracts.ErrCacheMiss
	}
	return json.Unmarshal(e.raw, target)
}

func (m *memoryCache) Set(_ context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = memoryEntry{raw: b, expireAt: expireAt}
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
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func (m *memoryCache) Invalidate(ctx context.Context, key string) error {
	return m.Delete(ctx, key)
}
