// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package msg_gateway_test

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway"
	"Wavelet/plugins/domain/msg_gateway/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushRegistry(t *testing.T) {
	ctx := core.NewContext(context.Background())
	require.NoError(t, msg_gateway.New().Apply(ctx))

	registry, err := ctx.Inject[contracts.PushRegistry]()
	require.NoError(t, err)
	require.NotNil(t, registry)

	botReg, err := ctx.Inject[contracts.BotCommandRegistry]()
	require.NoError(t, err)
	require.NotNil(t, botReg)

	const key = "test.push_registry.probe"
	registry.RegisterBuiltInEvent(contracts.PushEventMeta{
		Key:         key,
		Name:        "Push Registry Probe",
		Description: "observability probe for contracts.PushRegistry",
		DefaultTemplate: contracts.PushNotificationTemplate{
			Title:   "Probe Title",
			Content: "Probe Content",
			Level:   "INFO",
			Ext:     map[string]any{"source": "test"},
		},
	})

	found := false
	for _, ev := range service.GetBuiltInEvents() {
		if ev.Key != key {
			continue
		}
		found = true
		assert.Equal(t, "Push Registry Probe", ev.Name)
		assert.Equal(t, "observability probe for contracts.PushRegistry", ev.Description)
		assert.Equal(t, "Probe Title", ev.DefaultTemplate.Title)
		assert.Equal(t, "Probe Content", ev.DefaultTemplate.Content)
		assert.Equal(t, "INFO", ev.DefaultTemplate.Level)
		assert.Equal(t, map[string]any{"source": "test"}, ev.DefaultTemplate.Ext)
		break
	}
	require.True(t, found, "registered key %q should be visible via GetBuiltInEvents", key)
}

func TestNotificationPushUsesContractEventKey(t *testing.T) {
	assert.Equal(t, "notification:push", contracts.EventTopicNotificationPush)

	ctx := core.NewContext(context.Background())
	require.NoError(t, msg_gateway.New().Apply(ctx))

	var received contracts.PushNotificationEvent
	var fired bool
	ctx.Events().On(contracts.EventTopicNotificationPush, func(_ context.Context, e contracts.PushNotificationEvent) error {
		fired = true
		received = e
		return nil
	})

	err := ctx.Events().Emit(context.Background(), contracts.EventTopicNotificationPush, contracts.PushNotificationEvent{
		EventKey: "igo.grab_succeeded",
		UserID:   42,
		Title:    "抢座成功",
		Content:  "A1 预约成功",
		Metadata: map[string]any{"library_name": "二楼"},
	})
	require.NoError(t, err)
	require.True(t, fired)
	assert.Equal(t, "igo.grab_succeeded", received.EventKey)
	assert.Equal(t, uint64(42), received.UserID)

	var aliased msg_gateway.PushNotificationEvent = received
	assert.Equal(t, "igo.grab_succeeded", aliased.EventKey)
}
