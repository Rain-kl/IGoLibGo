// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"fmt"
	"strings"

	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/model/do"
)

// ShowCommand lists the user's pipeline configs.
type ShowCommand struct {
	api pipelineAPI
}

var _ contracts.BotCommand = (*ShowCommand)(nil)

// Name returns "show".
func (*ShowCommand) Name() string { return "show" }

// Aliases returns none.
func (*ShowCommand) Aliases() []string { return nil }

// Description is listed in /help.
func (*ShowCommand) Description() string { return "列出一条龙自动化配置" }

// Usage returns "/show".
func (*ShowCommand) Usage() string { return "/show" }

// Handle replies with the user's pipeline cards.
func (c *ShowCommand) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
	configs, err := c.api.ListPipelineConfigs(ctx, req.UserID())
	if err != nil {
		if replyErr := req.Reply("获取配置列表失败，请稍后重试。"); replyErr != nil {
			return replyErr
		}
		return err
	}
	if len(configs) == 0 {
		return req.Reply("您当前暂无配置任何一条龙自动化卡片。\n请前往 Web 控制台「自动化」页面新增配置。")
	}
	return req.Reply(formatConfigList(configs))
}

func formatConfigList(configs []do.PipelineConfigDTO) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "您的一条龙自动化配置列表（共 %d 个）：\n", len(configs))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	for i, cfg := range configs {
		autoCheckinText := "关闭"
		if cfg.AutoCheckin {
			autoCheckinText = "开启"
		}
		authStatus := "未授权"
		if cfg.HasCookie {
			authStatus = "有效"
		}
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, cfg.ID, cfg.Name)
		fmt.Fprintf(&sb, "   • 目标场馆: %s (%s楼)\n", cfg.LibraryName, cfg.Floor)
		fmt.Fprintf(&sb, "   • 目标座位: %s (%s)\n", cfg.SeatName, cfg.SeatKey)
		fmt.Fprintf(&sb, "   • 自动签到: %s\n", autoCheckinText)
		fmt.Fprintf(&sb, "   • 凭据状态: %s\n\n", authStatus)
	}
	sb.WriteString("发送 /run <配置ID> 即可立即执行。")
	return sb.String()
}
