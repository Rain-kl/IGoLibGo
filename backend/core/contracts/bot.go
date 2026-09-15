// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"context"
	"time"
)

// Note: Bot 命令接入契约由 msg_gateway Provide 注册表，业务插件只面向本文件 — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// BotInbound is the normalized inbound message payload for bot command handling.
type BotInbound struct {
	ChannelID      uint64
	ChannelType    string
	PlatformUserID string
	ChatID         string
	MessageID      string
	Text           string
	UserID         uint64
}

// BotCommand is a registered slash/text command handled by the bot registry.
type BotCommand interface {
	Name() string
	Aliases() []string
	Description() string
	Usage() string
	Handle(ctx context.Context, req BotCommandRequest) error
}

// BotCommandRequest is the per-invocation context passed to BotCommand.Handle.
type BotCommandRequest interface {
	UserID() uint64
	Args() []string
	Text() string
	Inbound() BotInbound
	Reply(text string) error
	Begin(conversation string, state any, ttl time.Duration) error
	HasConversation() bool
	CancelActive() error
}

// BotConversation is a multi-turn conversation handler registered with the bot.
type BotConversation interface {
	Name() string
	OnMessage(ctx context.Context, req BotConversationRequest) error
	OnCancel(ctx context.Context, req BotConversationRequest) error
}

// BotConversationRequest is the per-message context for an active conversation.
type BotConversationRequest interface {
	UserID() uint64
	Text() string
	Inbound() BotInbound
	Reply(text string) error
	State(dst any) error
	SetState(v any) error
	Transition(conversation string, state any) error
	End() error
}

// BotCommandRegistry registers bot commands and conversation handlers.
type BotCommandRegistry interface {
	Register(cmd BotCommand) error
	RegisterConversation(conv BotConversation) error
}
