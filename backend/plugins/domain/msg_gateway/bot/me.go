// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"fmt"
	"strings"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
	"Wavelet/plugins/domain/msg_gateway/model/entity"
	"Wavelet/plugins/domain/msg_gateway/service"
)

// Note: /me 未绑定走配对；已绑定只展示身份与绑定，不含 email/phone — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// MeCommand reports bind status and Wavelet identity for the current sender.
type MeCommand struct {
	UnboundReply func(ctx context.Context, req contracts.BotCommandRequest) error
	ListBindings func(ctx context.Context, userID uint64) ([]entity.MessageBinding, error)
	GetChannel   func(ctx context.Context, id uint64) (*entity.MessageChannel, error)
	GetUser      func(ctx context.Context, id uint64) (*contracts.UserDTO, error)
}

var _ contracts.BotCommand = (*MeCommand)(nil)

// Name returns the canonical command token "me".
func (*MeCommand) Name() string { return consts.CommandMe }

// Aliases reports that me has no aliases.
func (*MeCommand) Aliases() []string { return nil }

// Description returns the listing blurb shown in /help.
func (*MeCommand) Description() string { return "查看绑定状态与个人信息" }

// Usage returns "/me".
func (*MeCommand) Usage() string { return "/me" }

// Handle replies with pairing guidance when unbound, or identity and bindings when bound.
func (c *MeCommand) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
	if req.UserID() == 0 {
		return c.replyUnbound(ctx, req)
	}
	return req.Reply(c.formatBound(ctx, req))
}

func (c *MeCommand) replyUnbound(ctx context.Context, req contracts.BotCommandRequest) error {
	if c != nil && c.UnboundReply != nil {
		return c.UnboundReply(ctx, req)
	}
	in := req.Inbound()
	msg := do.InboundMessage{
		ChannelID:      in.ChannelID,
		ChannelType:    in.ChannelType,
		PlatformUserID: in.PlatformUserID,
		ChatID:         in.ChatID,
		MessageID:      in.MessageID,
		Text:           in.Text,
	}
	return service.ReplyPairingCode(ctx, msg, func(_ context.Context, _ uint64, _ do.Recipient, text string) error {
		return req.Reply(text)
	})
}

func (c *MeCommand) formatBound(ctx context.Context, req contracts.BotCommandRequest) string {
	userID := req.UserID()
	in := req.Inbound()
	var b strings.Builder
	fmt.Fprintf(&b, "用户 ID：%d", userID)

	if user, err := c.lookupUser(ctx, userID); err == nil && user != nil {
		if name := strings.TrimSpace(user.Username); name != "" {
			fmt.Fprintf(&b, "\n用户名：%s", name)
		}
		if nick := strings.TrimSpace(user.Nickname); nick != "" {
			fmt.Fprintf(&b, "\n昵称：%s", nick)
		}
	}

	currentName, currentType := c.channelLabel(ctx, in.ChannelID, in.ChannelType)
	fmt.Fprintf(&b, "\n当前频道：%s", formatChannelLine(currentName, currentType))
	if pid := strings.TrimSpace(in.PlatformUserID); pid != "" {
		fmt.Fprintf(&b, "\n平台用户 ID：%s", pid)
	}

	others := c.otherBindings(ctx, userID, in.ChannelID)
	if len(others) == 0 {
		return b.String()
	}
	b.WriteString("\n\n其它绑定：")
	for _, line := range others {
		fmt.Fprintf(&b, "\n- %s", line)
	}
	return b.String()
}

func formatChannelLine(name, typ string) string {
	name = strings.TrimSpace(name)
	typ = strings.TrimSpace(typ)
	switch {
	case name != "" && typ != "":
		return name + "（" + typ + "）"
	case name != "":
		return name
	case typ != "":
		return typ
	default:
		return "-"
	}
}

func (c *MeCommand) lookupUser(ctx context.Context, id uint64) (*contracts.UserDTO, error) {
	if c != nil && c.GetUser != nil {
		return c.GetUser(ctx, id)
	}
	svc := service.GetUserService(ctx)
	if svc == nil {
		return nil, nil
	}
	return svc.GetUserByID(ctx, id)
}

func (c *MeCommand) listBindings(ctx context.Context, userID uint64) ([]entity.MessageBinding, error) {
	if c != nil && c.ListBindings != nil {
		return c.ListBindings(ctx, userID)
	}
	return dao.ListBindingsByUser(ctx, userID)
}

func (c *MeCommand) getChannel(ctx context.Context, id uint64) (*entity.MessageChannel, error) {
	if c != nil && c.GetChannel != nil {
		return c.GetChannel(ctx, id)
	}
	return dao.GetMessageChannel(ctx, id)
}

func (c *MeCommand) channelLabel(ctx context.Context, channelID uint64, fallbackType string) (name, typ string) {
	typ = fallbackType
	if channelID == 0 {
		return "", typ
	}
	ch, err := c.getChannel(ctx, channelID)
	if err != nil || ch == nil {
		return "", typ
	}
	if ch.Name != "" {
		name = ch.Name
	}
	if ch.Type != "" {
		typ = ch.Type
	}
	return name, typ
}

func (c *MeCommand) otherBindings(ctx context.Context, userID, currentChannelID uint64) []string {
	rows, err := c.listBindings(ctx, userID)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(rows))
	for i := range rows {
		if rows[i].ChannelID == currentChannelID {
			continue
		}
		name, typ := c.channelLabel(ctx, rows[i].ChannelID, "")
		line := formatChannelLine(name, typ)
		if line == "-" {
			line = fmt.Sprintf("%d", rows[i].ChannelID)
		}
		out = append(out, line)
	}
	return out
}
