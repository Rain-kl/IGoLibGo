// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"strings"
	"unicode/utf8"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/service"
)

// Note: /start 与 /help 同一 Handle；系统行看到 alias start 时并进 help 那行 — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// HelpCommand lists registered builtin and business commands.
type HelpCommand struct {
	List func() []service.CommandMeta
}

var _ contracts.BotCommand = (*HelpCommand)(nil)

// Name returns the canonical command token "help".
func (*HelpCommand) Name() string { return consts.CommandHelp }

// Aliases returns start so /start shares this handler.
func (*HelpCommand) Aliases() []string { return []string{consts.CommandStart} }

// Description returns the listing blurb shown in /help.
func (*HelpCommand) Description() string { return "查看全部可用指令" }

// Usage returns "/help".
func (*HelpCommand) Usage() string { return "/help" }

// Handle replies with system and business command sections.
func (c *HelpCommand) Handle(_ context.Context, req contracts.BotCommandRequest) error {
	var metas []service.CommandMeta
	if c != nil && c.List != nil {
		metas = c.List()
	}
	return req.Reply(formatHelp(metas))
}

func formatHelp(metas []service.CommandMeta) string {
	var builtins, business []service.CommandMeta
	for _, meta := range metas {
		if meta.Builtin {
			builtins = append(builtins, meta)
			continue
		}
		business = append(business, meta)
	}

	var b strings.Builder
	writeHelpSection(&b, "系统指令：", builtins)
	if len(business) == 0 {
		return strings.TrimRight(b.String(), "\n")
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	writeHelpSection(&b, "业务指令：", business)
	return strings.TrimRight(b.String(), "\n")
}

func writeHelpSection(b *strings.Builder, title string, metas []service.CommandMeta) {
	if len(metas) == 0 {
		return
	}
	lefts := make([]string, len(metas))
	width := 0
	for i, meta := range metas {
		lefts[i] = commandLeft(meta)
		if n := utf8.RuneCountInString(lefts[i]); n > width {
			width = n
		}
	}
	b.WriteString(title)
	b.WriteByte('\n')
	for i, meta := range metas {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(lefts[i])
		if pad := width - utf8.RuneCountInString(lefts[i]); pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString("  - ")
		b.WriteString(meta.Description)
	}
	b.WriteByte('\n')
}

func commandLeft(meta service.CommandMeta) string {
	if hasStartAlias(meta.Aliases) {
		return "/" + consts.CommandStart + ", /" + meta.Name
	}
	if usage := strings.TrimSpace(meta.Usage); usage != "" {
		return usage
	}
	return "/" + meta.Name
}

func hasStartAlias(aliases []string) bool {
	for _, alias := range aliases {
		if strings.EqualFold(strings.TrimSpace(alias), consts.CommandStart) {
			return true
		}
	}
	return false
}
