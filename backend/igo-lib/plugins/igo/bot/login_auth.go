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

// LoginAuthConversation collects a login cookie or WeChat auth URL.
type LoginAuthConversation struct {
	api pipelineAPI
}

var _ contracts.BotConversation = (*LoginAuthConversation)(nil)

// Name returns igo.login_auth.
func (*LoginAuthConversation) Name() string { return convLoginAuth }

// OnMessage treats inbound text as a cookie or auth URL and retries the pipeline.
func (c *LoginAuthConversation) OnMessage(ctx context.Context, req contracts.BotConversationRequest) error {
	var st authState
	if err := req.State(&st); err != nil || strings.TrimSpace(st.ConfigID) == "" {
		if endErr := req.End(); endErr != nil {
			return endErr
		}
		return req.Reply("会话已失效，请重新发送 /run。")
	}
	res, err := c.api.RunPipeline(ctx, req.UserID(), st.ConfigID, &do.RunPipelineRequest{Cookie: req.Text()})
	if err != nil {
		if isNotFound(err) {
			if endErr := req.End(); endErr != nil {
				return endErr
			}
			return req.Reply(fmt.Sprintf("配置「%s」已不存在，已退出引导。", st.ConfigID))
		}
		if replyErr := req.Reply("无法识别登录凭据，请重新复制完整授权链接或 Cookie 发送（或发送 /cancel 退出）。"); replyErr != nil {
			return replyErr
		}
		return err
	}
	switch res.NeedAuth {
	case needAuthLogin:
		return req.Reply("登录凭据仍无效。请重新获取微信授权链接或 Cookie 发送（或发送 /cancel 退出）。")
	case needAuthCheckin:
		if err := req.Reply("登录凭据已更新。\n\n" + checkinPrompt(st.ConfigID, res.AuthURL)); err != nil {
			return err
		}
		return req.Transition(convCheckinAuth, st)
	default:
		if err := req.End(); err != nil {
			return err
		}
		return req.Reply("登录凭据已更新！\n\n" + formatRunResult(res))
	}
}

// OnCancel lets the gateway reply the standard cancel text.
func (*LoginAuthConversation) OnCancel(context.Context, contracts.BotConversationRequest) error {
	return nil
}
