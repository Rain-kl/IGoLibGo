// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"
)

const (
	convLoginAuth   = "igo.login_auth"
	convCheckinAuth = "igo.checkin_auth"
	botAuthTTL      = 5 * time.Minute

	needAuthLogin   = "LOGIN"
	needAuthCheckin = "CHECKIN"
)

type pipelineAPI interface {
	ListPipelineConfigs(ctx context.Context, userID uint64) ([]do.PipelineConfigDTO, error)
	RunPipeline(ctx context.Context, userID uint64, id string, overrideReq *do.RunPipelineRequest) (*do.PipelineRunResult, error)
}

type authState struct {
	ConfigID string `json:"config_id"`
}

func loginPrompt(configID, authURL string) string {
	if strings.TrimSpace(authURL) == "" {
		authURL = consts.WeChatLoginAuthURL
	}
	return fmt.Sprintf("一条龙配置「%s」的登录凭据已过期或未授权。\n\n请在微信中打开下方授权链接完成登录，并将跳转后的链接或登录 Cookie 发送给机器人：\n%s\n\n发送 /cancel 可退出引导。", configID, authURL)
}

func checkinPrompt(configID, authURL string) string {
	if strings.TrimSpace(authURL) == "" {
		authURL = consts.WeChatCheckinAuthURL
	}
	return fmt.Sprintf("一条龙配置「%s」的微信签到授权已过期或未授权。\n\n请在微信中打开下方授权链接完成签到授权，并将复制的链接发送给机器人：\n%s\n\n发送 /cancel 可退出引导。", configID, authURL)
}

func formatRunResult(res *do.PipelineRunResult) string {
	if res == nil {
		return "一条龙自动化执行未成功，请稍后重试。"
	}
	if res.Success {
		checkinInfo := ""
		if res.CheckinStatus != "" {
			checkinInfo = fmt.Sprintf("• 签到结果: %s\n", res.CheckinStatus)
		}
		return fmt.Sprintf("一条龙自动化执行成功！\n\n• 配置: %s (%s)\n• 占座结果: %s\n%s• 执行时间: %s",
			res.Name, res.ConfigID, res.ReservationStatus, checkinInfo, res.ExecutedAt)
	}
	msg := res.Message
	if strings.TrimSpace(msg) == "" {
		msg = "执行未成功"
	}
	return fmt.Sprintf("一条龙自动化执行未成功\n\n• 配置: %s (%s)\n• 结果提示: %s\n• 执行时间: %s",
		res.Name, res.ConfigID, msg, res.ExecutedAt)
}

func isNotFound(err error) bool {
	var coded *consts.CodedError
	return errors.As(err, &coded) && coded.Code == consts.CodeNotFound
}
