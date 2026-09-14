// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"strings"
	"sync"
	"time"
)

// Emitter publishes domain events onto the Cordis EventBus.
type Emitter interface {
	Emit(ctx context.Context, topic string, payload any) error
}

func (s *Service) SetEmitter(e Emitter) { s.events = e }

// RegisterPushEvents declares IGo events in the notification center.
func RegisterPushEvents(reg contracts.PushRegistry) {
	if reg == nil {
		return
	}
	for _, meta := range builtinPushEvents() {
		reg.RegisterBuiltInEvent(meta)
	}
}

func builtinPushEvents() []contracts.PushEventMeta {
	return []contracts.PushEventMeta{
		{
			Key:         consts.PushGrabSucceeded,
			Name:        "抢座成功",
			Description: "抢座任务预约到目标座位时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "抢座成功",
				Content: "{{library_name}} 的座位 {{seat_name}} 已预约成功。",
				Level:   "INFO",
			},
		},
		{
			Key:         consts.PushOccupySucceeded,
			Name:        "占座成功",
			Description: "占座任务完成重新预约时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "占座成功",
				Content: "座位 {{seat_name}} 已重新预约成功。",
				Level:   "INFO",
			},
		},
		{
			Key:         consts.PushGlobalLeakSucceeded,
			Name:        "全域捡漏成功",
			Description: "全域捡漏预约到空座时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "全域捡漏成功",
				Content: "{{library_name}} 的座位 {{seat_name}} 捡漏成功。",
				Level:   "INFO",
			},
		},
		{
			Key:         consts.PushTomorrowSucceeded,
			Name:        "明日预约成功",
			Description: "明日预约提交成功时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "明日预约成功",
				Content: "{{library_name}} 的座位 {{seat_name}} 明日预约已提交。",
				Level:   "INFO",
			},
		},
		{
			Key:         consts.PushTaskFailed,
			Name:        "任务失败",
			Description: "抢座、占座、捡漏或明日预约失败时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "任务失败",
				Content: "{{task_name}} 失败：{{reason}}",
				Level:   "ERROR",
			},
		},
		{
			Key:         consts.PushCookieExpiring,
			Name:        "Cookie 即将过期",
			Description: "TraceInt Cookie 到期前约 10 分钟触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "Cookie 即将过期",
				Content: "TraceInt Cookie 将于 {{expires_at}} 过期，请尽快重新授权。",
				Level:   "WARN",
			},
		},
		{
			Key:         consts.PushSessionInvalid,
			Name:        "Cookie 已失效",
			Description: "TraceInt 会话失效或授权被拒绝时触发。可在通知中心绑定推送渠道。",
			DefaultTemplate: contracts.PushNotificationTemplate{
				Title:   "Cookie 已失效",
				Content: "TraceInt 会话已失效：{{reason}}",
				Level:   "ERROR",
			},
		},
	}
}

var cookieAlertSent sync.Map

func (s *Service) notify(ctx context.Context, userID uint64, key, title, content string, data map[string]any) {
	if s.events == nil {
		return
	}
	if data == nil {
		data = map[string]any{}
	}
	data["time"] = time.Now().Format("2006-01-02 15:04:05")
	_ = s.events.Emit(ctx, contracts.EventTopicNotificationPush, contracts.PushNotificationEvent{
		EventKey: key,
		UserID:   userID,
		Title:    title,
		Content:  content,
		Metadata: data,
	})
}

func (s *Service) notifySuccess(ctx context.Context, userID uint64, kind, libraryName, seatName, message string) {
	key := consts.PushGrabSucceeded
	title := "抢座成功"
	switch kind {
	case consts.TaskKindOccupy:
		key, title = consts.PushOccupySucceeded, "占座成功"
	case consts.TaskKindGlobalLeak:
		key, title = consts.PushGlobalLeakSucceeded, "全域捡漏成功"
	case consts.TaskKindTomorrow:
		key, title = consts.PushTomorrowSucceeded, "明日预约成功"
	}
	s.notify(ctx, userID, key, title, message, map[string]any{
		"task_name":    kindTitle(kind),
		"library_name": libraryName,
		"seat_name":    seatName,
		"message":      message,
	})
}

func (s *Service) notifyFailure(ctx context.Context, userID uint64, kind, reason string) {
	if isAuthFailure(reason) {
		s.notify(ctx, userID, consts.PushSessionInvalid, "Cookie 已失效", reason, map[string]any{
			"task_name": kindTitle(kind),
			"reason":    reason,
		})
		return
	}
	s.notify(ctx, userID, consts.PushTaskFailed, "任务失败", kindTitle(kind)+" 失败："+reason, map[string]any{
		"task_name": kindTitle(kind),
		"reason":    reason,
	})
}

func (s *Service) maybeCookieAlert(ctx context.Context, userID uint64, cookie string) {
	exp := traceint.CookieExpiration(cookie)
	if exp == nil {
		return
	}
	remaining := time.Until(*exp)
	if remaining <= 0 || remaining > time.Duration(consts.CookieExpiringLead)*time.Second {
		return
	}
	token := exp.Unix()
	if prev, ok := cookieAlertSent.Load(userID); ok {
		if prev.(int64) == token {
			return
		}
	}
	cookieAlertSent.Store(userID, token)
	s.notify(ctx, userID, consts.PushCookieExpiring, "Cookie 即将过期",
		"TraceInt Cookie 将于 "+exp.Local().Format("2006-01-02 15:04:05")+" 过期",
		map[string]any{"expires_at": exp.Local().Format("2006-01-02 15:04:05")})
}

// HandleCookieWatch scans stored sessions for cookies about to expire.
func (s *Service) HandleCookieWatch(ctx context.Context, _ []byte) error {
	rows, err := dao.ListSessions(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		s.maybeCookieAlert(ctx, rows[i].UserID, rows[i].Cookie)
	}
	return nil
}

func isAuthFailure(msg string) bool {
	m := strings.ToLower(msg)
	for _, k := range []string{"未登录", "授权", "cookie", "unauthorized", "forbidden", "过期", "失效"} {
		if strings.Contains(m, strings.ToLower(k)) {
			return true
		}
	}
	return false
}
