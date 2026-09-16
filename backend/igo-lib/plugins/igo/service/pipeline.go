// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxPipelineConfigIDLength = 64

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validatePipelineConfigID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "配置 ID 不能为空")
	}
	if len(id) > maxPipelineConfigIDLength {
		return consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "配置 ID 长度不能超过 64 个字符")
	}
	if !idRegex.MatchString(id) {
		return consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "配置 ID 只能包含英文字母、数字、中划线和下划线")
	}
	return nil
}

func (s *Service) resolveConfigCheckinToken(ctx context.Context, userID uint64, reqCheckinToken string) (string, *time.Time, error) {
	if strings.TrimSpace(reqCheckinToken) == "" {
		return "", nil, nil
	}
	checkinToken, checkinExp, err := s.resolveCheckinToken(ctx, userID, reqCheckinToken)
	if err != nil {
		return "", nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if _, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, checkinToken); err != nil {
		return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeTraceInt, fmt.Sprintf("签到授权凭据验证失败: %v", err))
	}
	return checkinToken, checkinExp, nil
}

// CreatePipelineConfig creates a new pipeline configuration.
func (s *Service) CreatePipelineConfig(ctx context.Context, userID uint64, req do.CreatePipelineConfigRequest) (*do.PipelineConfigDTO, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	if err := validatePipelineConfigID(req.ID); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "配置名称不能为空")
	}
	if req.SeatKey == "" {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "必须选择目标座位")
	}
	if req.AutoCheckin && req.CheckinInfoID == 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "开启自动签到时必须选择签到信息")
	}

	existing, err := dao.GetPipelineConfig(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeConflict, fmt.Sprintf("配置 ID「%s」已被占用，请更换其他名称", req.ID))
	}

	occupy, info, checkinAcc, err := s.loadPipelineRefs(ctx, userID, req.AccountID, req.CheckinAccountID, req.CheckinInfoID, req.AutoCheckin)
	if err != nil {
		return nil, err
	}

	row := &entity.PipelineConfig{
		ID:               req.ID,
		UserID:           userID,
		Name:             req.Name,
		Cookie:           occupy.Cookie,
		CookieExpiresAt:  occupy.CookieExpiresAt,
		LibraryID:        req.LibraryID,
		LibraryName:      req.LibraryName,
		Floor:            req.Floor,
		SeatKey:          req.SeatKey,
		SeatName:         req.SeatName,
		AutoCheckin:      req.AutoCheckin,
		CheckinToken:     checkinAcc.CheckinToken,
		CheckinExpiresAt: checkinAcc.CheckinExpiresAt,
		AccountID:        occupy.ID,
		CheckinAccountID: req.CheckinAccountID,
		CheckinInfoID:    req.CheckinInfoID,
	}
	if info != nil {
		row.BeaconUUID = info.BeaconUUID
		row.Major = info.Major
		row.Minor = info.Minor
		row.Latitude = info.Latitude
		row.Longitude = info.Longitude
	}
	if err := dao.CreatePipelineConfig(ctx, row); err != nil {
		return nil, err
	}
	dto := s.toPipelineDTO(ctx, row)
	return &dto, nil
}

// ListPipelineConfigs returns all pipeline cards owned by the user.
func (s *Service) ListPipelineConfigs(ctx context.Context, userID uint64) ([]do.PipelineConfigDTO, error) {
	rows, err := dao.ListPipelineConfigsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]do.PipelineConfigDTO, 0, len(rows))
	for i := range rows {
		out = append(out, s.toPipelineDTO(ctx, &rows[i]))
	}
	return out, nil
}

// GetPipelineConfig returns one pipeline configuration by ID.
func (s *Service) GetPipelineConfig(ctx context.Context, userID uint64, id string) (*do.PipelineConfigDTO, error) {
	row, err := dao.GetPipelineConfigByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该一条龙配置")
	}
	dto := s.toPipelineDTO(ctx, row)
	return &dto, nil
}

func applyPipelineConfigFields(row *entity.PipelineConfig, req *do.UpdatePipelineConfigRequest) {
	if req.Name != "" {
		row.Name = strings.TrimSpace(req.Name)
	}
	if req.LibraryID > 0 {
		row.LibraryID = req.LibraryID
	}
	if req.LibraryName != "" {
		row.LibraryName = req.LibraryName
	}
	if req.Floor != "" {
		row.Floor = req.Floor
	}
	if req.SeatKey != "" {
		row.SeatKey = req.SeatKey
	}
	if req.SeatName != "" {
		row.SeatName = req.SeatName
	}
	row.AutoCheckin = req.AutoCheckin
	if req.AccountID != 0 {
		row.AccountID = req.AccountID
	}
	row.CheckinAccountID = req.CheckinAccountID
	if req.CheckinInfoID != 0 || !req.AutoCheckin {
		row.CheckinInfoID = req.CheckinInfoID
	}
}

// UpdatePipelineConfig updates an existing pipeline configuration.
func (s *Service) UpdatePipelineConfig(ctx context.Context, userID uint64, id string, req do.UpdatePipelineConfigRequest) (*do.PipelineConfigDTO, error) {
	row, err := dao.GetPipelineConfigByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该一条龙配置")
	}

	applyPipelineConfigFields(row, &req)
	if row.AutoCheckin && row.CheckinInfoID == 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "开启自动签到时必须选择签到信息")
	}
	occupy, info, checkinAcc, err := s.loadPipelineRefs(ctx, userID, row.AccountID, row.CheckinAccountID, row.CheckinInfoID, row.AutoCheckin)
	if err != nil {
		return nil, err
	}
	row.Cookie = occupy.Cookie
	row.CookieExpiresAt = occupy.CookieExpiresAt
	row.CheckinToken = checkinAcc.CheckinToken
	row.CheckinExpiresAt = checkinAcc.CheckinExpiresAt
	if info != nil {
		row.BeaconUUID = info.BeaconUUID
		row.Major = info.Major
		row.Minor = info.Minor
		row.Latitude = info.Latitude
		row.Longitude = info.Longitude
	}

	if err := dao.UpdatePipelineConfig(ctx, row); err != nil {
		return nil, err
	}
	dto := s.toPipelineDTO(ctx, row)
	return &dto, nil
}

// DeletePipelineConfig deletes a pipeline configuration.
func (s *Service) DeletePipelineConfig(ctx context.Context, userID uint64, id string) error {
	return dao.DeletePipelineConfig(ctx, id, userID)
}

func (s *Service) applyPipelineOverrides(ctx context.Context, row *entity.PipelineConfig, occupy, checkinAcc *entity.Account, overrideReq *do.RunPipelineRequest) {
	if overrideReq == nil {
		return
	}
	if occupy != nil && strings.TrimSpace(overrideReq.Cookie) != "" {
		c, exp, err := s.resolveCookie(ctx, row.UserID, overrideReq.Cookie)
		if err == nil {
			occupy.Cookie = c
			occupy.CookieExpiresAt = exp
			_ = dao.UpdateAccount(ctx, occupy)
		}
	}
	if checkinAcc != nil && strings.TrimSpace(overrideReq.CheckinToken) != "" {
		t, exp, err := s.resolveCheckinToken(ctx, row.UserID, overrideReq.CheckinToken)
		if err == nil {
			checkinAcc.CheckinToken = t
			checkinAcc.CheckinExpiresAt = exp
			_ = dao.UpdateAccount(ctx, checkinAcc)
		}
	}
}

func (s *Service) validatePipelineAuth(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, occupy, checkinAcc *entity.Account, id, execTime string) (*do.PipelineRunResult, bool) {
	cookie := ""
	if occupy != nil {
		cookie = occupy.Cookie
	}
	if strings.TrimSpace(cookie) == "" {
		return &do.PipelineRunResult{
			Success:    false,
			ConfigID:   id,
			Name:       row.Name,
			NeedAuth:   "LOGIN",
			AuthURL:    consts.WeChatLoginAuthURL,
			Message:    "TraceInt 账户未授权或凭据为空，请重新登录授权",
			ExecutedAt: execTime,
		}, false
	}

	if _, err := cli.ListLibraries(ctx, tpl, cookie); err != nil {
		return &do.PipelineRunResult{
			Success:    false,
			ConfigID:   id,
			Name:       row.Name,
			NeedAuth:   "LOGIN",
			AuthURL:    consts.WeChatLoginAuthURL,
			Message:    "TraceInt 账户授权已过期，请重新登录授权",
			ExecutedAt: execTime,
		}, false
	}

	if row.AutoCheckin {
		token := ""
		if checkinAcc != nil {
			token = checkinAcc.CheckinToken
		}
		if strings.TrimSpace(token) == "" {
			return &do.PipelineRunResult{
				Success:    false,
				ConfigID:   id,
				Name:       row.Name,
				NeedAuth:   "CHECKIN",
				AuthURL:    consts.WeChatCheckinAuthURL,
				Message:    "已开启自动签到但签到凭据缺失，请授权签到链接",
				ExecutedAt: execTime,
			}, false
		}
		if _, err := cli.GetCheckInDevices(ctx, tpl, token); err != nil {
			return &do.PipelineRunResult{
				Success:    false,
				ConfigID:   id,
				Name:       row.Name,
				NeedAuth:   "CHECKIN",
				AuthURL:    consts.WeChatCheckinAuthURL,
				Message:    "签到微信授权已失效，请重新授权签到链接",
				ExecutedAt: execTime,
			}, false
		}
	}

	return nil, true
}

func (s *Service) checkAndReserveSeat(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, occupy *entity.Account, id, execTime string) (*do.PipelineRunResult, string, bool) {
	cookie := ""
	if occupy != nil {
		cookie = occupy.Cookie
	}
	layout, err := cli.GetLayout(ctx, tpl, cookie, row.LibraryID)
	if err != nil {
		return &do.PipelineRunResult{ //nolint:nilerr // result encapsulates failure
			Success:    false,
			ConfigID:   id,
			Name:       row.Name,
			Message:    fmt.Sprintf("获取场馆「%s」布局失败: %v", row.LibraryName, err),
			ExecutedAt: execTime,
		}, "", false
	}

	var targetSeat *do.SeatSnapshot
	for i := range layout.Seats {
		if layout.Seats[i].SeatKey == row.SeatKey {
			targetSeat = &layout.Seats[i]
			break
		}
	}
	if targetSeat == nil {
		return &do.PipelineRunResult{
			Success:    false,
			ConfigID:   id,
			Name:       row.Name,
			Message:    fmt.Sprintf("在场馆「%s」中未找到目标座位「%s」", row.LibraryName, row.SeatName),
			ExecutedAt: execTime,
		}, "", false
	}
	if targetSeat.IsOccupied {
		return &do.PipelineRunResult{
			Success:    false,
			ConfigID:   id,
			Name:       row.Name,
			Message:    fmt.Sprintf("目标座位 [%s %s] 当前已被占用，占座失败退出", row.LibraryName, row.SeatName),
			ExecutedAt: execTime,
		}, "", false
	}

	resOk, err := cli.ReserveSeat(ctx, tpl, cookie, row.LibraryID, row.SeatKey)
	if err != nil || !resOk {
		errMsg := "未知原因"
		if err != nil {
			errMsg = err.Error()
		}
		return &do.PipelineRunResult{ //nolint:nilerr // result encapsulates failure
			Success:           false,
			ConfigID:          id,
			Name:              row.Name,
			ReservationStatus: "占座失败",
			Message:           fmt.Sprintf("预约占座失败: %s", errMsg),
			ExecutedAt:        execTime,
		}, "", false
	}

	reservationStatus := fmt.Sprintf("成功预约 [%s %s]", row.LibraryName, row.SeatName)
	return nil, reservationStatus, true
}

func (s *Service) executeBeaconCheckin(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, checkinAcc *entity.Account, info *entity.CheckInInfo, id, execTime, reservationStatus string) *do.PipelineRunResult {
	if info == nil || !beaconComplete(info) {
		return &do.PipelineRunResult{
			Success:           true,
			ConfigID:          id,
			Name:              row.Name,
			ReservationStatus: reservationStatus,
			CheckinStatus:     "签到信息不完整",
			Message:           fmt.Sprintf("已成功占座 [%s %s]，请先补全签到信息的 Beacon 与坐标后再打卡", row.LibraryName, row.SeatName),
			ExecutedAt:        execTime,
		}
	}
	serverTime, err := cli.GetCheckInServerTime(ctx, tpl)
	if err != nil {
		return &do.PipelineRunResult{ //nolint:nilerr // result encapsulates failure
			Success:           true,
			ConfigID:          id,
			Name:              row.Name,
			ReservationStatus: reservationStatus,
			CheckinStatus:     "获取服务器时间失败",
			Message:           fmt.Sprintf("已成功占座 [%s %s]，但获取签到服务器时间失败: %v", row.LibraryName, row.SeatName, err),
			ExecutedAt:        execTime,
		}
	}

	lat, _ := strconv.ParseFloat(info.Latitude, 64)
	lng, _ := strconv.ParseFloat(info.Longitude, 64)
	token := ""
	if checkinAcc != nil {
		token = checkinAcc.CheckinToken
	}
	signReq := do.CheckInSignRequest{
		ExpectedLibraryID: row.LibraryID,
		BeaconUUID:        info.BeaconUUID,
		Major:             info.Major,
		Minor:             info.Minor,
		Latitude:          lat,
		Longitude:         lng,
	}
	signResp, err := cli.SignCheckIn(ctx, tpl, token, signReq, serverTime)
	if err != nil {
		return &do.PipelineRunResult{ //nolint:nilerr // result encapsulates failure
			Success:           true,
			ConfigID:          id,
			Name:              row.Name,
			ReservationStatus: reservationStatus,
			CheckinStatus:     fmt.Sprintf("打卡失败: %v", err),
			Message:           fmt.Sprintf("已成功占座 [%s %s]，但远程自动签到打卡失败: %v", row.LibraryName, row.SeatName, err),
			ExecutedAt:        execTime,
		}
	}

	return &do.PipelineRunResult{
		Success:           true,
		ConfigID:          id,
		Name:              row.Name,
		ReservationStatus: reservationStatus,
		CheckinStatus:     fmt.Sprintf("打卡成功 (%s)", signResp.Message),
		Message:           fmt.Sprintf("一条龙全流程执行成功！已成功占座 [%s %s] 并完成签到打卡", row.LibraryName, row.SeatName),
		ExecutedAt:        execTime,
	}
}

// RunPipeline executes the full all-in-one automation pipeline.
func (s *Service) RunPipeline(ctx context.Context, userID uint64, id string, overrideReq *do.RunPipelineRequest) (*do.PipelineRunResult, error) {
	var row *entity.PipelineConfig
	var err error

	if userID != 0 {
		row, err = dao.GetPipelineConfigByUser(ctx, id, userID)
	} else {
		row, err = dao.GetPipelineConfig(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, fmt.Sprintf("未找到一条龙配置「%s」", id))
	}

	if row.AccountID == 0 {
		_ = s.backfillOnePipeline(ctx, row)
		if userID != 0 {
			row, err = dao.GetPipelineConfigByUser(ctx, id, userID)
		} else {
			row, err = dao.GetPipelineConfig(ctx, id)
		}
		if err != nil {
			return nil, err
		}
		if row == nil || row.AccountID == 0 {
			return &do.PipelineRunResult{
				Success:    false,
				ConfigID:   id,
				Name:       id,
				NeedAuth:   "LOGIN",
				AuthURL:    consts.WeChatLoginAuthURL,
				Message:    "TraceInt 账户未授权或凭据为空，请重新登录授权",
				ExecutedAt: time.Now().Format("2006-01-02 15:04:05"),
			}, nil
		}
	}

	occupy, info, checkinAcc, err := s.loadPipelineRefs(ctx, row.UserID, row.AccountID, row.CheckinAccountID, row.CheckinInfoID, row.AutoCheckin)
	if err != nil {
		return nil, err
	}
	s.applyPipelineOverrides(ctx, row, occupy, checkinAcc, overrideReq)

	execTime := time.Now().Format("2006-01-02 15:04:05")
	tpl, err := s.templates(ctx, row.UserID)
	if err != nil {
		return nil, err
	}
	cli := s.api(ctx, row.UserID)

	if authErrRes, ok := s.validatePipelineAuth(ctx, cli, tpl, row, occupy, checkinAcc, id, execTime); !ok {
		return authErrRes, nil
	}

	resErr, resStatus, ok := s.checkAndReserveSeat(ctx, cli, tpl, row, occupy, id, execTime)
	if !ok {
		return resErr, nil
	}

	if !row.AutoCheckin {
		return &do.PipelineRunResult{
			Success:           true,
			ConfigID:          id,
			Name:              row.Name,
			ReservationStatus: resStatus,
			Message:           fmt.Sprintf("一条龙自动化执行成功！已成功锁定座位 [%s %s]", row.LibraryName, row.SeatName),
			ExecutedAt:        execTime,
		}, nil
	}

	return s.executeBeaconCheckin(ctx, cli, tpl, row, checkinAcc, info, id, execTime, resStatus), nil
}

// HelperVerifySession probes a cookie and lists accessible libraries.
func (s *Service) HelperVerifySession(ctx context.Context, userID uint64, rawInput string, accountID uint64) ([]do.LibrarySummary, string, *time.Time, error) {
	var cookie string
	var exp *time.Time
	var err error
	if strings.TrimSpace(rawInput) == "" && accountID != 0 {
		acc, accErr := dao.GetAccountByUser(ctx, accountID, userID)
		if accErr != nil {
			return nil, "", nil, accErr
		}
		if acc == nil {
			return nil, "", nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
		}
		cookie = acc.Cookie
		exp = acc.CookieExpiresAt
	} else {
		cookie, exp, err = s.resolveCookie(ctx, userID, rawInput)
		if err != nil {
			return nil, "", nil, err
		}
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, "", nil, err
	}
	libs, err := s.api(ctx, userID).ListLibraries(ctx, tpl, cookie)
	if err != nil {
		return nil, "", nil, wrapTrace(err)
	}
	return libs, cookie, exp, nil
}

// HelperGetLibraryLayout fetches seat layout for a library.
// Explicit cookie wins, then account_id, then a still-valid stored session.
func (s *Service) HelperGetLibraryLayout(ctx context.Context, userID uint64, cookie string, accountID uint64, libID int) (*do.LibraryLayoutResponse, error) {
	if strings.TrimSpace(cookie) == "" && accountID != 0 {
		acc, err := dao.GetAccountByUser(ctx, accountID, userID)
		if err != nil {
			return nil, err
		}
		if acc == nil {
			return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
		}
		cookie = acc.Cookie
	}
	resolved, _, err := s.resolveCookiePreferStored(ctx, userID, cookie)
	if err != nil {
		return nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	layout, err := s.api(ctx, userID).GetLayout(ctx, tpl, resolved, libID)
	return layout, wrapTrace(err)
}

func (s *Service) resolveCookiePreferStored(ctx context.Context, userID uint64, rawInput string) (string, *time.Time, error) {
	if strings.TrimSpace(rawInput) != "" {
		return s.resolveCookie(ctx, userID, rawInput)
	}
	row, err := dao.GetSession(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if row == nil || strings.TrimSpace(row.Cookie) == "" {
		return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请传入登录凭据或先完成 TraceInt 登录")
	}
	exp := traceint.CookieExpiration(row.Cookie)
	if exp == nil {
		exp = row.ExpiresAt
	}
	if exp != nil && !exp.After(time.Now()) {
		return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "登录凭据已过期，请传入 cookie")
	}
	return row.Cookie, exp, nil
}

// HelperVerifyCheckin validates checkin authorization token or code and retrieves device list.
func (s *Service) HelperVerifyCheckin(ctx context.Context, userID uint64, rawInput string) (*do.CheckInDeviceResponse, string, *time.Time, error) {
	token, exp, err := s.resolveCheckinToken(ctx, userID, rawInput)
	if err != nil {
		return nil, "", nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, "", nil, err
	}
	devs, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, token)
	if err != nil {
		return nil, "", nil, wrapTrace(err)
	}
	return devs, token, exp, nil
}

func (s *Service) resolveCookie(ctx context.Context, userID uint64, input string) (string, *time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请输入登录凭据或微信授权链接")
	}
	if strings.Contains(input, "code=") || len(input) == 32 {
		code, ok := traceint.ExtractCode(input)
		if !ok {
			return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "无效的授权链接或 Code")
		}
		tpl, err := s.templates(ctx, userID)
		if err != nil {
			return "", nil, err
		}
		c, err := s.api(ctx, userID).GetCookie(ctx, tpl, code)
		if err != nil {
			return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeTraceInt, fmt.Sprintf("通过授权码换取 Cookie 失败: %v", err))
		}
		exp := traceint.CookieExpiration(c)
		return c, exp, nil
	}
	exp := traceint.CookieExpiration(input)
	return input, exp, nil
}

func (s *Service) resolveCheckinToken(ctx context.Context, userID uint64, input string) (string, *time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请输入签到授权凭据或微信授权链接")
	}
	if strings.Contains(input, "code=") || (len(input) == 32 && !strings.Contains(input, "=")) {
		code, ok := traceint.ExtractCode(input)
		if !ok {
			return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "无效的签到授权链接或 Code")
		}
		tpl, err := s.templates(ctx, userID)
		if err != nil {
			return "", nil, err
		}
		tok, exp, err := s.api(ctx, userID).ExchangeCheckInCode(ctx, tpl, code)
		if err != nil {
			return "", nil, consts.NewError(http.StatusBadRequest, consts.CodeTraceInt, fmt.Sprintf("换取签到授权失败: %v", err))
		}
		return tok, exp, nil
	}
	return input, nil, nil
}

func (s *Service) toPipelineDTO(ctx context.Context, row *entity.PipelineConfig) do.PipelineConfigDTO {
	if row == nil {
		return do.PipelineConfigDTO{}
	}
	dto := do.PipelineConfigDTO{
		ID:               row.ID,
		UserID:           row.UserID,
		Name:             row.Name,
		AccountID:        row.AccountID,
		CheckinAccountID: row.CheckinAccountID,
		CheckinInfoID:    row.CheckinInfoID,
		HasCookie:        row.Cookie != "",
		CookieMasked:     traceint.MaskCookie(row.Cookie),
		CookieExpiresAt:  row.CookieExpiresAt,
		LibraryID:        row.LibraryID,
		LibraryName:      row.LibraryName,
		Floor:            row.Floor,
		SeatKey:          row.SeatKey,
		SeatName:         row.SeatName,
		AutoCheckin:      row.AutoCheckin,
		HasCheckinToken:  row.CheckinToken != "",
		CheckinExpiresAt: row.CheckinExpiresAt,
		BeaconUUID:       row.BeaconUUID,
		Major:            row.Major,
		Minor:            row.Minor,
		Latitude:         row.Latitude,
		Longitude:        row.Longitude,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	if row.AccountID != 0 {
		if acc, err := dao.GetAccountByUser(ctx, row.AccountID, row.UserID); err == nil && acc != nil {
			a := toAccountDTO(acc)
			dto.Account = &a
			dto.HasCookie = acc.Cookie != ""
			dto.CookieMasked = traceint.MaskCookie(acc.Cookie)
			dto.CookieExpiresAt = acc.CookieExpiresAt
		}
	}
	checkinID := row.CheckinAccountID
	if checkinID == 0 {
		checkinID = row.AccountID
	}
	if checkinID != 0 {
		if acc, err := dao.GetAccountByUser(ctx, checkinID, row.UserID); err == nil && acc != nil {
			a := toAccountDTO(acc)
			if row.CheckinAccountID != 0 {
				dto.CheckinAccount = &a
			}
			dto.HasCheckinToken = acc.CheckinToken != ""
			dto.CheckinExpiresAt = acc.CheckinExpiresAt
		}
	}
	if row.CheckinInfoID != 0 {
		if info, err := dao.GetCheckInInfoByUser(ctx, row.CheckinInfoID, row.UserID); err == nil && info != nil {
			d := toCheckInInfoDTO(info)
			dto.CheckinInfo = &d
			dto.BeaconUUID = info.BeaconUUID
			dto.Major = info.Major
			dto.Minor = info.Minor
			dto.Latitude = info.Latitude
			dto.Longitude = info.Longitude
		}
	}
	return dto
}

func (s *Service) loadPipelineRefs(ctx context.Context, userID, accountID, checkinAccountID, checkinInfoID uint64, autoCheckin bool) (*entity.Account, *entity.CheckInInfo, *entity.Account, error) {
	if accountID == 0 {
		return nil, nil, nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "必须选择占座账户")
	}
	occupy, err := dao.GetAccountByUser(ctx, accountID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	if occupy == nil {
		return nil, nil, nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到占座账户")
	}
	checkinAcc := occupy
	if checkinAccountID != 0 {
		checkinAcc, err = dao.GetAccountByUser(ctx, checkinAccountID, userID)
		if err != nil {
			return nil, nil, nil, err
		}
		if checkinAcc == nil {
			return nil, nil, nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到打卡账户")
		}
	}
	var info *entity.CheckInInfo
	if checkinInfoID != 0 {
		info, err = dao.GetCheckInInfoByUser(ctx, checkinInfoID, userID)
		if err != nil {
			return nil, nil, nil, err
		}
		if info == nil {
			return nil, nil, nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到签到信息")
		}
	} else if autoCheckin {
		return nil, nil, nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "开启自动签到时必须选择签到信息")
	}
	return occupy, info, checkinAcc, nil
}

func beaconComplete(info *entity.CheckInInfo) bool {
	if info == nil {
		return false
	}
	_, ok := traceint.NormalizeUUID(info.BeaconUUID)
	return ok && strings.TrimSpace(info.Latitude) != "" && strings.TrimSpace(info.Longitude) != ""
}
