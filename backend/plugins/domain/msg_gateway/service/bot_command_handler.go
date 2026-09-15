// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	igoconsts "Wavelet/igo-lib/plugins/igo/consts"
	igodao "Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/service"
	"Wavelet/igo-lib/plugins/igo/traceint"
	mgconsts "Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/dao"
	mgdo "Wavelet/plugins/domain/msg_gateway/model/do"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	botSessionKeyPrefix = "igo_bot_session"
	botSessionTTL       = 5 * time.Minute

	stepWaitingLoginAuth   = "waiting_login_auth"
	stepWaitingCheckinAuth = "waiting_checkin_auth"
)

// BotSessionState represents the temporary interactive session state for re-authentication.
type BotSessionState struct {
	ConfigID string `json:"config_id"`
	Step     string `json:"step"`
	ExpireAt int64  `json:"expire_at"`
}

// BotCommandHandler processes inbound private-chat messages and commands.
type BotCommandHandler struct {
	igoSvc *service.Service
}

// NewBotCommandHandler creates a command handler.
func NewBotCommandHandler(igoSvc *service.Service) *BotCommandHandler {
	if igoSvc == nil {
		igoSvc = service.New()
	}
	return &BotCommandHandler{igoSvc: igoSvc}
}

// HandleInbound processes one incoming bot message.
func (h *BotCommandHandler) HandleInbound(ctx context.Context, in mgdo.InboundMessage) error {
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil
	}

	// 1. Check if platform user is bound
	binding, err := dao.GetBindingByChannelPlatform(ctx, in.ChannelID, in.PlatformUserID)
	if err != nil && !errors.Is(err, mgconsts.ErrRecordNotFound) {
		return err
	}

	// Not bound: generate pairing code and guide binding
	if binding == nil {
		code, err := GenerateCode()
		if err != nil {
			return err
		}
		_, _ = dao.UpsertPairingCode(ctx, in.ChannelID, in.PlatformUserID, code, time.Now().Add(15*time.Minute))
		reply := fmt.Sprintf("👋 您好！您尚未绑定 Wavelet 账号。\n您的专属配对码为：`%s`\n请前往 Web 控制台「个人设置 - 绑定机器人」中输入此配对码完成绑定。", FormatCode(code))
		return ReplyToInbound(ctx, in, reply)
	}

	userID := binding.UserID
	sessionKey := fmt.Sprintf("%s:%d:%s", botSessionKeyPrefix, in.ChannelID, in.PlatformUserID)

	// 2. Check for active multi-step session
	session := h.loadSession(ctx, sessionKey)
	if session != nil {
		if strings.EqualFold(text, "/cancel") {
			h.clearSession(ctx, sessionKey)
			return ReplyToInbound(ctx, in, "已取消当前的鉴权补录流程。")
		}

		if session.Step == stepWaitingLoginAuth {
			return h.handleWaitingLoginAuth(ctx, in, userID, sessionKey, session, text)
		}

		if session.Step == stepWaitingCheckinAuth {
			return h.handleWaitingCheckinAuth(ctx, in, userID, sessionKey, session, text)
		}
	}

	// 3. Command dispatcher
	if strings.EqualFold(text, "/help") || strings.EqualFold(text, "/start") {
		helpMsg := "🤖 IGo 自动化一条龙机器人指令：\n\n" +
			"• /help 或 /start - 查看所有可用指令\n" +
			"• /show - 列出您名下的全部一条龙自动化配置\n" +
			"• /run [配置 ID] - 立即执行指定的一条龙任务\n" +
			"• /cancel - 取消当前正在进行的补录会话"
		return ReplyToInbound(ctx, in, helpMsg)
	}

	if strings.EqualFold(text, "/show") {
		return h.handleShow(ctx, in, userID)
	}

	if strings.HasPrefix(strings.ToLower(text), "/run") {
		parts := strings.Fields(text)
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			return ReplyToInbound(ctx, in, "⚠️ 请指定要执行的配置 ID，例如：`/run myseat01`\n可通过 `/show` 查看您已保存的配置 ID。")
		}
		configID := strings.TrimSpace(parts[1])
		return h.executePipeline(ctx, in, userID, sessionKey, configID, nil)
	}

	return nil
}

func (h *BotCommandHandler) handleShow(ctx context.Context, in mgdo.InboundMessage, userID uint64) error {
	configs, err := igodao.ListPipelineConfigsByUser(ctx, userID)
	if err != nil {
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 获取配置列表失败: %v", err))
	}
	if len(configs) == 0 {
		return ReplyToInbound(ctx, in, "📋 您当前暂无配置任何一条龙自动化卡片。\n请前往 Web 控制台「自动化」页面新增配置。")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 您的一条龙自动化配置列表（共 %d 个）：\n", len(configs)))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	for i, c := range configs {
		autoCheckinText := "关闭"
		if c.AutoCheckin {
			autoCheckinText = "开启"
		}
		credStatus := "有效"
		if c.Cookie == "" {
			credStatus = "未授权"
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, c.ID, c.Name))
		sb.WriteString(fmt.Sprintf("   • 目标场馆: %s (%s楼)\n", c.LibraryName, c.Floor))
		sb.WriteString(fmt.Sprintf("   • 目标座位: %s (%s)\n", c.SeatName, c.SeatKey))
		sb.WriteString(fmt.Sprintf("   • 自动签到: %s\n", autoCheckinText))
		sb.WriteString(fmt.Sprintf("   • 凭据状态: %s\n\n", credStatus))
	}
	sb.WriteString("👉 发送 `/run [配置 ID]` 即可立即执行！")
	return ReplyToInbound(ctx, in, sb.String())
}

func (h *BotCommandHandler) executePipeline(ctx context.Context, in mgdo.InboundMessage, userID uint64, sessionKey string, configID string, overrideReq *do.RunPipelineRequest) error {
	cfg, err := igodao.GetPipelineConfigByUser(ctx, configID, userID)
	if err != nil {
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 查询配置失败: %v", err))
	}
	if cfg == nil {
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 未找到配置 ID「%s」，请发送 `/show` 查看您的有效配置列表。", configID))
	}

	res, err := h.igoSvc.RunPipeline(ctx, userID, configID, overrideReq)
	if err != nil {
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 执行异常: %v", err))
	}

	if res.NeedAuth == "LOGIN" {
		h.saveSession(ctx, sessionKey, &BotSessionState{
			ConfigID: configID,
			Step:     stepWaitingLoginAuth,
			ExpireAt: time.Now().Add(botSessionTTL).Unix(),
		})
		reply := fmt.Sprintf("⚠️ 一条龙配置「%s」的 TraceInt 登录凭据已过期或未授权。\n\n"+
			"请在微信中打开下方授权链接完成登录授权，并将跳转后的链接或登录 Cookie 发送给机器人：\n%s\n\n"+
			"（发送 /cancel 可随时退出引导）", configID, igoconsts.WeChatLoginAuthURL)
		return ReplyToInbound(ctx, in, reply)
	}

	if res.NeedAuth == "CHECKIN" {
		h.saveSession(ctx, sessionKey, &BotSessionState{
			ConfigID: configID,
			Step:     stepWaitingCheckinAuth,
			ExpireAt: time.Now().Add(botSessionTTL).Unix(),
		})
		reply := fmt.Sprintf("⚠️ 一条龙配置「%s」的微信签到授权已过期或未授权。\n\n"+
			"请在微信中打开下方授权链接完成签到授权，并将跳转后的链接发送给机器人：\n%s\n\n"+
			"（发送 /cancel 可随时退出引导）", configID, igoconsts.WeChatCheckinAuthURL)
		return ReplyToInbound(ctx, in, reply)
	}

	if res.Success {
		checkinInfo := ""
		if res.CheckinStatus != "" {
			checkinInfo = fmt.Sprintf("• 签到结果: %s\n", res.CheckinStatus)
		}
		reply := fmt.Sprintf("🎉 一条龙自动化执行成功！\n\n"+
			"• 配置: %s (%s)\n"+
			"• 占座结果: %s\n"+
			"%s"+
			"• 执行时间: %s", cfg.Name, cfg.ID, res.ReservationStatus, checkinInfo, res.ExecutedAt)
		return ReplyToInbound(ctx, in, reply)
	}

	reply := fmt.Sprintf("❌ 一条龙自动化执行未成功\n\n"+
		"• 配置: %s (%s)\n"+
		"• 结果提示: %s\n"+
		"• 执行时间: %s", cfg.Name, cfg.ID, res.Message, res.ExecutedAt)
	return ReplyToInbound(ctx, in, reply)
}

func (h *BotCommandHandler) handleWaitingLoginAuth(ctx context.Context, in mgdo.InboundMessage, userID uint64, sessionKey string, session *BotSessionState, input string) error {
	configID := session.ConfigID
	cfg, err := igodao.GetPipelineConfigByUser(ctx, configID, userID)
	if err != nil || cfg == nil {
		h.clearSession(ctx, sessionKey)
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 配置「%s」已不存在，已退出引导流程。", configID))
	}

	var cookie string
	var exp *time.Time

	if strings.Contains(input, "code=") || len(input) == 32 {
		code, ok := traceint.ExtractCode(input)
		if !ok {
			return ReplyToInbound(ctx, in, "⚠️ 无法识别微信授权链接中的 Code，请重新复制完整链接发送（或发送 /cancel 退出）：")
		}
		tpl := traceint.DefaultTemplates()
		c, err := (&traceint.Client{}).GetCookie(ctx, tpl, code)
		if err != nil {
			return ReplyToInbound(ctx, in, fmt.Sprintf("⚠️ 通过授权码换取凭据失败: %v\n请重新获取微信授权链接发送（或发送 /cancel 退出）：", err))
		}
		cookie = c
		exp = traceint.CookieExpiration(c)
	} else {
		cookie = input
		exp = traceint.CookieExpiration(input)
	}

	_ = igodao.UpdatePipelineCookie(ctx, configID, cookie, exp)

	// Check if auto checkin is enabled and checkin auth is also needed
	if cfg.AutoCheckin && (cfg.CheckinToken == "" || (cfg.CheckinExpiresAt != nil && cfg.CheckinExpiresAt.Before(time.Now()))) {
		session.Step = stepWaitingCheckinAuth
		h.saveSession(ctx, sessionKey, session)
		reply := fmt.Sprintf("✅ 登录凭据已更新！\n\n"+
			"由于配置「%s」开启了自动签到，请在微信中打开下方签到授权链接完成授权，并将复制的链接发送给机器人：\n%s\n\n"+
			"（发送 /cancel 可随时退出引导）", configID, igoconsts.WeChatCheckinAuthURL)
		return ReplyToInbound(ctx, in, reply)
	}

	// Done with auth, clear session and execute pipeline!
	h.clearSession(ctx, sessionKey)
	_ = ReplyToInbound(ctx, in, "✅ 登录凭据已更新！正在立即执行一条龙自动化任务...")
	return h.executePipeline(ctx, in, userID, sessionKey, configID, nil)
}

func (h *BotCommandHandler) handleWaitingCheckinAuth(ctx context.Context, in mgdo.InboundMessage, userID uint64, sessionKey string, session *BotSessionState, input string) error {
	configID := session.ConfigID
	cfg, err := igodao.GetPipelineConfigByUser(ctx, configID, userID)
	if err != nil || cfg == nil {
		h.clearSession(ctx, sessionKey)
		return ReplyToInbound(ctx, in, fmt.Sprintf("❌ 配置「%s」已不存在，已退出引导流程。", configID))
	}

	var token string
	var exp *time.Time

	if strings.Contains(input, "code=") || (len(input) == 32 && !strings.Contains(input, "=")) {
		code, ok := traceint.ExtractCode(input)
		if !ok {
			return ReplyToInbound(ctx, in, "⚠️ 无法识别微信签到授权链接中的 Code，请重新复制完整链接发送（或发送 /cancel 退出）：")
		}
		tpl := traceint.DefaultTemplates()
		tok, tExp, err := (&traceint.Client{}).ExchangeCheckInCode(ctx, tpl, code)
		if err != nil {
			return ReplyToInbound(ctx, in, fmt.Sprintf("⚠️ 换取签到授权失败: %v\n请重新在微信中授权并发送链接（或发送 /cancel 退出）：", err))
		}
		token = tok
		exp = tExp
	} else {
		token = input
	}

	_ = igodao.UpdatePipelineCheckinToken(ctx, configID, token, exp)
	h.clearSession(ctx, sessionKey)
	_ = ReplyToInbound(ctx, in, "✅ 签到凭据已更新！正在立即执行一条龙自动化任务...")
	return h.executePipeline(ctx, in, userID, sessionKey, configID, nil)
}

func (h *BotCommandHandler) loadSession(ctx context.Context, key string) *BotSessionState {
	cache := dao.GetCache(ctx)
	if cache == nil {
		return nil
	}
	var s BotSessionState
	if err := cache.Get(ctx, key, &s); err != nil {
		return nil
	}
	if s.ExpireAt > 0 && time.Now().Unix() > s.ExpireAt {
		_ = cache.Delete(ctx, key)
		return nil
	}
	return &s
}

func (h *BotCommandHandler) saveSession(ctx context.Context, key string, s *BotSessionState) {
	cache := dao.GetCache(ctx)
	if cache == nil {
		return
	}
	_ = cache.Set(ctx, key, s, botSessionTTL)
}

func (h *BotCommandHandler) clearSession(ctx context.Context, key string) {
	cache := dao.GetCache(ctx)
	if cache == nil {
		return
	}
	_ = cache.Delete(ctx, key)
}

// ReplyToInbound sends a text reply to the sender of an inbound message.
func ReplyToInbound(ctx context.Context, in mgdo.InboundMessage, text string) error {
	row, err := dao.GetMessageChannel(ctx, in.ChannelID)
	if err != nil {
		return err
	}
	factory, ok := Lookup(row.Type)
	if !ok {
		return fmt.Errorf("bot channel factory not found: %s", row.Type)
	}
	cfg, err := channelConfigFromRow(row)
	if err != nil {
		return err
	}
	ch, err := factory(cfg, nil)
	if err != nil {
		return err
	}
	if err := ch.Connect(ctx); err != nil {
		return err
	}
	defer func() { _ = ch.Disconnect(ctx) }()

	to := mgdo.Recipient{
		ChatID:         in.ChatID,
		PlatformUserID: in.PlatformUserID,
	}
	return ch.Send(ctx, to, mgdo.OutboundMessage{Text: text})
}
