// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"

	"Wavelet/core/contracts"
	"Wavelet/pkg/logger"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
)

// Note: 入站斜杠命令先于配对分发；已绑定非命令有占位则进 OnMessage — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

const unknownCommandReply = "未知命令，发送 /help 查看可用指令"

// SendTextFn sends a text reply on an already-connected channel.
// Signature matches Runner.SendText; Dispatch must not Connect a new adapter.
type SendTextFn func(ctx context.Context, channelID uint64, to do.Recipient, text string) error

// ParseBotCommand extracts a slash command from inbound text.
// Matching is case-insensitive; Telegram @bot suffixes are stripped.
func ParseBotCommand(text string) (name string, args []string, isCmd bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil, false
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", nil, false
	}
	token := fields[0]
	if !strings.HasPrefix(token, "/") {
		return "", nil, false
	}
	token = strings.TrimPrefix(token, "/")
	if i := strings.Index(token, "@"); i >= 0 {
		token = token[:i]
	}
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return "", nil, false
	}
	var rest []string
	if len(fields) > 1 {
		rest = fields[1:]
	}
	return token, rest, true
}

// Dispatch routes one inbound private-chat message: allow-unbound builtins,
// pairing for other unbound traffic, registered commands for bound users,
// occupancy for in-conversation text, and drops bound non-commands.
func (r *BotRegistry) Dispatch(ctx context.Context, msg do.InboundMessage, send SendTextFn) error {
	if strings.TrimSpace(msg.Text) == "" {
		return nil
	}

	binding, err := dao.GetBindingByChannelPlatform(ctx, msg.ChannelID, msg.PlatformUserID)
	if err != nil && !errors.Is(err, consts.ErrRecordNotFound) {
		logger.ErrorF(ctx, "bot inbound: query binding channel=%d platform_user=%s: %v", msg.ChannelID, msg.PlatformUserID, err)
		return err
	}

	name, args, isCmd := ParseBotCommand(msg.Text)
	if binding == nil {
		if isCmd {
			cmd, _, allowUnbound, ok := r.LookupCommand(name)
			if ok && allowUnbound {
				r.invokeHandle(ctx, cmd, newCommandRequest(ctx, r, 0, args, msg.Text, botInboundFrom(msg, 0), send))
				return nil
			}
		}
		return ReplyPairingCode(ctx, msg, send)
	}

	userID := binding.UserID
	msg.BindingUserID = &userID
	in := botInboundFrom(msg, userID)
	req := newCommandRequest(ctx, r, userID, args, msg.Text, in, send)
	if isCmd {
		return r.dispatchBoundCommand(ctx, name, req, in, send)
	}
	if r.deliverOccupancy(ctx, in, send) {
		return nil
	}
	return nil
}

func (r *BotRegistry) dispatchBoundCommand(ctx context.Context, name string, req *commandRequest, in contracts.BotInbound, send SendTextFn) error {
	if name == consts.CommandCancel {
		if cmd, _, _, ok := r.LookupCommand(name); ok {
			r.invokeHandle(ctx, cmd, req)
			return nil
		}
	}
	if cmd, _, _, ok := r.LookupCommand(name); ok {
		if err := deleteOccupancy(ctx, in.ChannelID, in.PlatformUserID); err != nil {
			logger.ErrorF(ctx, "bot inbound: clear occupancy before command %q: %v", name, err)
		}
		r.invokeHandle(ctx, cmd, req)
		return nil
	}
	if r.deliverOccupancy(ctx, in, send) {
		return nil
	}
	replyUnknownCommand(ctx, in, send)
	return nil
}

func (r *BotRegistry) invokeHandle(ctx context.Context, cmd contracts.BotCommand, req contracts.BotCommandRequest) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.ErrorF(ctx, "bot inbound: command %q panic: %v\n%s", cmd.Name(), rec, debug.Stack())
		}
	}()
	if err := cmd.Handle(ctx, req); err != nil {
		logger.ErrorF(ctx, "bot inbound: command %q: %v", cmd.Name(), err)
	}
}

func replyUnknownCommand(ctx context.Context, in contracts.BotInbound, send SendTextFn) {
	if send == nil {
		return
	}
	to := do.Recipient{ChatID: in.ChatID, PlatformUserID: in.PlatformUserID}
	if err := send(ctx, in.ChannelID, to, unknownCommandReply); err != nil {
		logger.ErrorF(ctx, "bot inbound: reply unknown command: %v", err)
	}
}

func botInboundFrom(msg do.InboundMessage, userID uint64) contracts.BotInbound {
	return contracts.BotInbound{
		ChannelID:      msg.ChannelID,
		ChannelType:    msg.ChannelType,
		PlatformUserID: msg.PlatformUserID,
		ChatID:         msg.ChatID,
		MessageID:      msg.MessageID,
		Text:           msg.Text,
		UserID:         userID,
	}
}

type commandRequest struct {
	ctx    context.Context
	reg    *BotRegistry
	userID uint64
	args   []string
	text   string
	in     contracts.BotInbound
	send   SendTextFn
}

var _ contracts.BotCommandRequest = (*commandRequest)(nil)

func newCommandRequest(ctx context.Context, reg *BotRegistry, userID uint64, args []string, text string, in contracts.BotInbound, send SendTextFn) *commandRequest {
	return &commandRequest{ctx: ctx, reg: reg, userID: userID, args: args, text: text, in: in, send: send}
}

func (r *commandRequest) UserID() uint64                { return r.userID }
func (r *commandRequest) Args() []string                { return r.args }
func (r *commandRequest) Text() string                  { return r.text }
func (r *commandRequest) Inbound() contracts.BotInbound { return r.in }

func (r *commandRequest) Reply(text string) error {
	if r.send == nil {
		return nil
	}
	return r.send(r.ctx, r.in.ChannelID, doRecipient(r.in), text)
}
