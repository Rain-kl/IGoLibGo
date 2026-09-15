// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package bot implements msg_gateway builtin slash commands (/help, /start, /me, /cancel).
package bot

import "Wavelet/plugins/domain/msg_gateway/service"

// Note: 自带指令 AllowUnbound 挂 help/cancel/me — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// RegisterBuiltins registers gateway builtin commands that may run unbound.
func RegisterBuiltins(reg *service.BotRegistry) error {
	help := &HelpCommand{List: reg.List}
	if err := reg.RegisterBuiltin(help, true); err != nil {
		return err
	}
	if err := reg.RegisterBuiltin(&CancelCommand{}, true); err != nil {
		return err
	}
	if err := reg.RegisterBuiltin(&MeCommand{}, true); err != nil {
		return err
	}
	return nil
}
