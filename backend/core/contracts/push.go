// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package contracts defines unified service interfaces and DTOs for cross-plugin communication.
package contracts

import "context"

// EventTopicNotificationPush is the EventBus topic for triggering a configured push event.
const EventTopicNotificationPush = "notification:push"

// PushNotificationEvent is the cross-plugin payload for EventTopicNotificationPush.
// EventKey must match a built-in event registered via PushRegistry so the
// notification center can bind channels. Channel is kept for older emitters.
type PushNotificationEvent struct {
	EventKey string         `json:"event_key"`
	UserID   uint64         `json:"user_id,string"`
	Channel  string         `json:"channel,omitempty"`
	Title    string         `json:"title"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PushNotificationTemplate defines notification message template payload.
type PushNotificationTemplate struct {
	Title   string
	Content string
	Level   string
	Ext     map[string]any
}

// PushEventMeta defines metadata for a system push event.
type PushEventMeta struct {
	Key             string
	Name            string
	Description     string
	DefaultTemplate PushNotificationTemplate
}

// PushRegistry defines the interface for registering built-in events.
type PushRegistry interface {
	RegisterBuiltInEvent(meta PushEventMeta)
	SyncEvents(ctx context.Context) error
}
