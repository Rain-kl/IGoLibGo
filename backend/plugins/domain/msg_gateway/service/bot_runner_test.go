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
	"time"

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

type ctxCaptureChannel struct {
	connectCtx  context.Context
	connects    int
	disconnects int
	sentText    string
}

func (m *ctxCaptureChannel) Type() string { return "ctx_capture" }
func (m *ctxCaptureChannel) Connect(ctx context.Context) error {
	m.connects++
	m.connectCtx = ctx
	return nil
}
func (m *ctxCaptureChannel) Disconnect(context.Context) error {
	m.disconnects++
	return nil
}
func (m *ctxCaptureChannel) Send(_ context.Context, _ do.Recipient, msg do.OutboundMessage) error {
	m.sentText = msg.Text
	return nil
}
func (m *ctxCaptureChannel) Capabilities() do.Capability { return do.Capability{Text: true} }

func setupRunnerDB(t *testing.T) context.Context {
	t.Helper()
	_ = idgen.Init(1)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "runner_test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}, &entity.MessagePairingCode{}))
	dao.SetDBServiceForTest(&testDBService{db: db})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })
	return context.Background()
}

func insertEnabledChannel(t *testing.T, ctx context.Context, typ, name string) entity.MessageChannel {
	t.Helper()
	cipher, err := service.EncryptCredentials(map[string]string{"bot_token": "mock_token"})
	require.NoError(t, err)
	row := entity.MessageChannel{
		Name:        name,
		Type:        typ,
		OwnerScope:  "system",
		Credentials: cipher,
		Enabled:     true,
	}
	require.NoError(t, dao.CreateMessageChannel(ctx, &row))
	return row
}

func TestRunner_ReloadDoesNotStopChannelWhenRequestContextCancels(t *testing.T) {
	ctx := setupRunnerDB(t)
	row := insertEnabledChannel(t, ctx, "ctx_capture", "tg")

	ch := &ctxCaptureChannel{}
	service.Register("ctx_capture", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return ch, nil
	})

	life, stop := context.WithCancel(context.Background())
	t.Cleanup(stop)

	runner := &service.Runner{}
	require.NoError(t, runner.Start(life))
	require.Equal(t, 1, ch.connects)

	reqCtx, cancelReq := context.WithCancel(context.Background())
	require.NoError(t, runner.Reload(reqCtx))
	cancelReq()

	require.NotNil(t, ch.connectCtx)
	select {
	case <-ch.connectCtx.Done():
		t.Fatal("channel lifetime must follow the runner, not the HTTP request context")
	default:
	}

	require.NoError(t, runner.SendText(ctx, row.ID, do.Recipient{PlatformUserID: "1"}, "still-alive"))
	assert.Equal(t, "still-alive", ch.sentText)

	runner.Stop()
	select {
	case <-ch.connectCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("runner.Stop must cancel the channel lifetime context")
	}
}

type gatedConnectChannel struct {
	entered chan struct{}
	release chan struct{}
}

func (g *gatedConnectChannel) Type() string { return "gated" }
func (g *gatedConnectChannel) Connect(context.Context) error {
	close(g.entered)
	<-g.release
	return nil
}
func (g *gatedConnectChannel) Disconnect(context.Context) error { return nil }
func (g *gatedConnectChannel) Send(context.Context, do.Recipient, do.OutboundMessage) error {
	return nil
}
func (g *gatedConnectChannel) Capabilities() do.Capability { return do.Capability{Text: true} }

func TestRunner_SendTextDoesNotDeadlockWhileConnectRuns(t *testing.T) {
	ctx := setupRunnerDB(t)
	row := insertEnabledChannel(t, ctx, "gated", "slow")

	gated := &gatedConnectChannel{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	service.Register("gated", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return gated, nil
	})

	runner := &service.Runner{}
	startErr := make(chan error, 1)
	go func() { startErr <- runner.Start(ctx) }()

	select {
	case <-gated.entered:
	case <-time.After(time.Second):
		t.Fatal("Connect did not start")
	}

	done := make(chan struct{})
	go func() {
		_ = runner.SendText(ctx, row.ID, do.Recipient{PlatformUserID: "1"}, "ping")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		close(gated.release)
		t.Fatal("SendText deadlocked because Connect held the runner mutex")
	}
	close(gated.release)
	require.NoError(t, <-startErr)
	runner.Stop()
}

func TestCreateChannel_ReturnsBeforeConnectFinishes(t *testing.T) {
	ctx := setupRunnerDB(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	service.Register("telegram", func(do.ChannelConfig, service.Handler) (service.Channel, error) {
		return &gatedConnectChannel{entered: entered, release: release}, nil
	})

	orig := service.GlobalRunner
	runner := &service.Runner{}
	service.GlobalRunner = runner
	t.Cleanup(func() {
		close(release)
		runner.Stop()
		service.GlobalRunner = orig
	})
	require.NoError(t, runner.Start(ctx))

	done := make(chan error, 1)
	go func() {
		enabled := true
		_, err := service.CreateChannel(ctx, do.CreateChannelRequest{
			Name:        "T",
			Type:        "telegram",
			Enabled:     &enabled,
			Credentials: map[string]string{"token": "mock_token"},
		})
		done <- err
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("CreateChannel must not wait for Telegram Connect")
	}
}
