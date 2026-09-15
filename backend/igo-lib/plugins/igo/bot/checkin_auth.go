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

// CheckinAuthConversation collects a WeChat check-in authorization URL or token.
type CheckinAuthConversation struct {
	api pipelineAPI
}

var _ contracts.BotConversation = (*CheckinAuthConversation)(nil)

// Name returns igo.checkin_auth.
func (*CheckinAuthConversation) Name() string { return convCheckinAuth }

// OnMessage treats inbound text as a check-in token or auth URL and retries the pipeline.
func (c *CheckinAuthConversation) OnMessage(ctx context.Context, req contracts.BotConversationRequest) error {
	var st authState
	if err := req.State(&st); err != nil || strings.TrimSpace(st.ConfigID) == "" {
		if endErr := req.End(); endErr != nil {
			return endErr
		}
		return req.Reply("会话已失效，请重新发送 /run。")
	}
	res, err := c.api.RunPipeline(ctx, req.UserID(), st.ConfigID, &do.RunPipelineRequest{CheckinToken: req.Text()})
	if err != nil {
		if isNotFound(err) {
			if endErr := req.End(); endErr != nil {
				return endErr
			}
			return req.Reply(fmt.Sprintf("配置「%s」已不存在，已退出引导。", st.ConfigID))
		}
		if replyErr := req.Reply("无法识别签到授权，请重新在微信中授权并发送链接（或发送 /cancel 退出）。"); replyErr != nil {
			return replyErr
		}
		return err
	}
	if res.NeedAuth == needAuthCheckin || res.NeedAuth == needAuthLogin {
		return req.Reply("签到授权仍无效。请重新获取微信签到授权链接发送（或发送 /cancel 退出）。")
	}
	if err := req.End(); err != nil {
		return err
	}
	return req.Reply("签到凭据已更新！\n\n" + formatRunResult(res))
}

// OnCancel lets the gateway reply the standard cancel text.
func (*CheckinAuthConversation) OnCancel(context.Context, contracts.BotConversationRequest) error {
	return nil
}
