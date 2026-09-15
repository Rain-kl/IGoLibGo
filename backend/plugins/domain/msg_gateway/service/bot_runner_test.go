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
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubChannel struct{}

func (stubChannel) Type() string                                                 { return "stub" }
func (stubChannel) Connect(context.Context) error                                { return nil }
func (stubChannel) Disconnect(context.Context) error                             { return nil }
func (stubChannel) Send(context.Context, do.Recipient, do.OutboundMessage) error { return nil }
func (stubChannel) Capabilities() do.Capability                                  { return do.Capability{Text: true} }

func TestRegisterLookup(t *testing.T) {
	service.Register("stub", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return stubChannel{}, nil
	})
	fn, ok := service.Lookup("stub")
	require.True(t, ok)

	ch, err := fn(do.ChannelConfig{}, nil)
	require.NoError(t, err)
	assert.Equal(t, "stub", ch.Type())
}

type mockChannel struct {
	sentText string
}

func (m *mockChannel) Type() string                     { return "stub_mock" }
func (m *mockChannel) Connect(context.Context) error    { return nil }
func (m *mockChannel) Disconnect(context.Context) error { return nil }
func (m *mockChannel) Send(_ context.Context, _ do.Recipient, msg do.OutboundMessage) error {
	m.sentText = msg.Text
	return nil
}
func (m *mockChannel) Capabilities() do.Capability { return do.Capability{Text: true} }

func TestRunner_Lifecycle(t *testing.T) {
	_ = idgen.Init(1)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "runner_test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}, &entity.MessagePairingCode{}))
	dao.SetDBServiceForTest(&testDBService{db: db})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })

	ctx := context.Background()

	cipher, err := service.EncryptCredentials(map[string]string{"bot_token": "mock_token"})
	require.NoError(t, err)

	chRow := entity.MessageChannel{
		Name:        "stub_bot",
		Type:        "stub_mock",
		OwnerScope:  "system",
		Credentials: cipher,
		Enabled:     true,
	}
	require.NoError(t, dao.CreateMessageChannel(ctx, &chRow))

	mockCh := &mockChannel{}
	service.Register("stub_mock", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return mockCh, nil
	})

	runner := &service.Runner{}
	require.NoError(t, runner.Start(ctx))

	err = runner.SendText(ctx, chRow.ID, do.Recipient{PlatformUserID: "123"}, "hello from runner")
	require.NoError(t, err)
	assert.Equal(t, "hello from runner", mockCh.sentText)

	require.NoError(t, runner.Reload(ctx))
	runner.Stop()
}
