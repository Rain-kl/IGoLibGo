// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
)

// Note: /cancel AllowUnbound；无对话不发配对码 — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// CancelCommand ends the active conversation occupancy, if any.
type CancelCommand struct{}

var _ contracts.BotCommand = (*CancelCommand)(nil)

// Name returns the canonical command token "cancel".
func (*CancelCommand) Name() string { return consts.CommandCancel }

// Aliases reports that cancel has no aliases.
func (*CancelCommand) Aliases() []string { return nil }

// Description returns the listing blurb shown in /help.
func (*CancelCommand) Description() string { return "取消当前进行中的操作" }

// Usage returns "/cancel".
func (*CancelCommand) Usage() string { return "/cancel" }

// Handle cancels the active conversation or says none is in progress.
func (*CancelCommand) Handle(_ context.Context, req contracts.BotCommandRequest) error {
	if !req.HasConversation() {
		return req.Reply("当前没有进行中的操作")
	}
	if err := req.CancelActive(); err != nil {
		return err
	}
	return req.Reply("已取消当前操作")
}
