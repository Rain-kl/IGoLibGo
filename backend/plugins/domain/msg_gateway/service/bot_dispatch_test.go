// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"Wavelet/pkg/idgen"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
	"Wavelet/plugins/domain/msg_gateway/model/entity"
	"Wavelet/plugins/domain/msg_gateway/service"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dispatchTestDB struct{ db *gorm.DB }

func (m *dispatchTestDB) GORM() *gorm.DB                  { return m.db }
func (m *dispatchTestDB) DB(ctx context.Context) *gorm.DB { return m.db.WithContext(ctx) }
func (m *dispatchTestDB) Named(_ string) *gorm.DB         { return m.db }

func TestBotDispatchValidatePayload(t *testing.T) {
	h := &service.BotDispatchHandler{}
	_, err := h.ValidatePayload([]byte(`{}`))
	require.Error(t, err)
	_, err = h.ValidatePayload([]byte(`{"text":"hello"}`))
	require.NoError(t, err)
}

type countingDispatchChannel struct {
	connects int
	sends    int
}

func (c *countingDispatchChannel) Type() string { return "telegram" }
func (c *countingDispatchChannel) Connect(context.Context) error {
	c.connects++
	return nil
}
func (c *countingDispatchChannel) Disconnect(context.Context) error { return nil }
func (c *countingDispatchChannel) Send(_ context.Context, _ do.Recipient, _ do.OutboundMessage) error {
	c.sends++
	return nil
}
func (c *countingDispatchChannel) Capabilities() do.Capability { return do.Capability{Text: true} }

func TestBotDispatchReusesConnectedRunnerChannel(t *testing.T) {
	_ = idgen.Init(1)
	testDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "dispatch_reuse.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}))
	dao.SetDBServiceForTest(&dispatchTestDB{db: testDB})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })

	ch := &countingDispatchChannel{}
	service.Register("telegram", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return ch, nil
	})

	cipher, err := service.EncryptCredentials(map[string]string{"token": "t"})
	require.NoError(t, err)
	row := entity.MessageChannel{
		Name: "TG", Type: "telegram", OwnerScope: "system", Credentials: cipher, Enabled: true,
	}
	require.NoError(t, dao.CreateMessageChannel(context.Background(), &row))
	require.NoError(t, dao.CreateMessageBinding(context.Background(), &entity.MessageBinding{
		ChannelID:      row.ID,
		UserID:         1,
		PlatformUserID: "42",
	}))

	orig := service.GlobalRunner
	runner := &service.Runner{}
	service.GlobalRunner = runner
	t.Cleanup(func() {
		runner.Stop()
		service.GlobalRunner = orig
	})
	require.NoError(t, runner.Start(context.Background()))
	require.Equal(t, 1, ch.connects)

	h := &service.BotDispatchHandler{}
	res, err := h.Execute(context.Background(), []byte(`{"text":"hello","channel_id":"`+fmt.Sprintf("%d", row.ID)+`"}`))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, ch.connects, "dispatch must not open a second getUpdates session")
	assert.Equal(t, 1, ch.sends)
}

func TestBotDispatchNoChannels(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "dispatch.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}))
	dao.SetDBServiceForTest(&dispatchTestDB{db: testDB})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })

	h := &service.BotDispatchHandler{}
	res, err := h.Execute(context.Background(), []byte(`{"text":"hello"}`))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.Message, "成功 0")
}
