// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"testing"

	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memRegistry struct {
	events []contracts.PushEventMeta
}

func (m *memRegistry) RegisterBuiltInEvent(meta contracts.PushEventMeta) {
	m.events = append(m.events, meta)
}

func (m *memRegistry) SyncEvents(context.Context) error { return nil }

type captureEmitter struct {
	topic   string
	payload any
}

func (c *captureEmitter) Emit(_ context.Context, topic string, payload any) error {
	c.topic = topic
	c.payload = payload
	return nil
}

func TestRegisterPushEvents(t *testing.T) {
	reg := &memRegistry{}
	RegisterPushEvents(reg)
	keys := map[string]struct{}{}
	for _, ev := range reg.events {
		keys[ev.Key] = struct{}{}
		assert.Contains(t, ev.Description, "通知中心")
	}
	assert.Equal(t, 7, len(reg.events))
	assert.Contains(t, keys, consts.PushGrabSucceeded)
	assert.Contains(t, keys, consts.PushTaskFailed)
	assert.Contains(t, keys, consts.PushCookieExpiring)
}

func TestNotifyEmitsConfiguredEventKey(t *testing.T) {
	em := &captureEmitter{}
	s := New()
	s.SetEmitter(em)
	s.notify(context.Background(), 9, consts.PushGrabSucceeded, "抢座成功", "A1 预约成功", map[string]any{
		"library_name": "二楼",
		"seat_name":    "A1",
	})
	assert.Equal(t, contracts.EventTopicNotificationPush, em.topic)
	ev, ok := em.payload.(contracts.PushNotificationEvent)
	require.True(t, ok)
	assert.Equal(t, consts.PushGrabSucceeded, ev.EventKey)
	assert.Equal(t, uint64(9), ev.UserID)
	assert.Equal(t, "二楼", ev.Metadata["library_name"])
}

func TestNotifyFailureUsesSessionInvalidForAuthErrors(t *testing.T) {
	em := &captureEmitter{}
	s := New()
	s.SetEmitter(em)
	s.notifyFailure(context.Background(), 3, consts.TaskKindGrab, "未登录，Cookie 已失效")
	ev := em.payload.(contracts.PushNotificationEvent)
	assert.Equal(t, consts.PushSessionInvalid, ev.EventKey)
}
