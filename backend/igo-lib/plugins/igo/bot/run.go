// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"fmt"
	"strings"

	"Wavelet/core/contracts"
)

// RunCommand executes a pipeline config immediately.
type RunCommand struct {
	api pipelineAPI
}

var _ contracts.BotCommand = (*RunCommand)(nil)

// Name returns "run".
func (*RunCommand) Name() string { return "run" }

// Aliases returns none.
func (*RunCommand) Aliases() []string { return nil }

// Description is listed in /help.
func (*RunCommand) Description() string { return "立即执行指定的一条龙配置" }

// Usage returns "/run <配置ID>".
func (*RunCommand) Usage() string { return "/run <配置ID>" }

// Handle runs the named pipeline or starts a credential conversation.
func (c *RunCommand) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
	args := req.Args()
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return req.Reply("请指定要执行的配置 ID，例如：/run myseat01\n可通过 /show 查看您已保存的配置 ID。")
	}
	configID := strings.TrimSpace(args[0])
	res, err := c.api.RunPipeline(ctx, req.UserID(), configID, nil)
	if err != nil {
		if isNotFound(err) {
			return req.Reply(fmt.Sprintf("未找到配置 ID「%s」，请发送 /show 查看您的有效配置列表。", configID))
		}
		if replyErr := req.Reply("执行异常，请稍后重试。"); replyErr != nil {
			return replyErr
		}
		return err
	}
	switch res.NeedAuth {
	case needAuthLogin:
		if err := req.Reply(loginPrompt(configID, res.AuthURL)); err != nil {
			return err
		}
		return req.Begin(convLoginAuth, authState{ConfigID: configID}, botAuthTTL)
	case needAuthCheckin:
		if err := req.Reply(checkinPrompt(configID, res.AuthURL)); err != nil {
			return err
		}
		return req.Begin(convCheckinAuth, authState{ConfigID: configID}, botAuthTTL)
	default:
		return req.Reply(formatRunResult(res))
	}
}
