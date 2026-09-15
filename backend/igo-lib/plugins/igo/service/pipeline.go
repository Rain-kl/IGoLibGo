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

	// Check if ID already exists
	existing, err := dao.GetPipelineConfig(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeConflict, fmt.Sprintf("配置 ID「%s」已被占用，请更换其他名称", req.ID))
	}

	// Resolve cookie (could be auth link or raw cookie)
	cookie, exp, err := s.resolveCookie(ctx, userID, req.Cookie)
	if err != nil {
		return nil, err
	}

	// Verify cookie by calling TraceInt
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	_, err = s.api(ctx, userID).ListLibraries(ctx, tpl, cookie)
	if err != nil {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeTraceInt, fmt.Sprintf("TraceInt 登录凭据验证失败: %v", err))
	}

	var checkinToken string
	var checkinExp *time.Time
	if req.AutoCheckin {
		checkinToken, checkinExp, err = s.resolveConfigCheckinToken(ctx, userID, req.CheckinToken)
		if err != nil {
			return nil, err
		}
	}

	beaconUUID := req.BeaconUUID
	if beaconUUID == "" {
		beaconUUID = req.BeaconMac
	}
	lat := req.Latitude
	if lat == "" {
		lat = req.BeaconLat
	}
	lng := req.Longitude
	if lng == "" {
		lng = req.BeaconLng
	}

	row := &entity.PipelineConfig{
		ID:               req.ID,
		UserID:           userID,
		Name:             req.Name,
		Cookie:           cookie,
		CookieExpiresAt:  exp,
		LibraryID:        req.LibraryID,
		LibraryName:      req.LibraryName,
		Floor:            req.Floor,
		SeatKey:          req.SeatKey,
		SeatName:         req.SeatName,
		AutoCheckin:      req.AutoCheckin,
		CheckinToken:     checkinToken,
		CheckinExpiresAt: checkinExp,
		BeaconUUID:       beaconUUID,
		Major:            req.Major,
		Minor:            req.Minor,
		Latitude:         lat,
		Longitude:        lng,
	}
	if err := dao.CreatePipelineConfig(ctx, row); err != nil {
		return nil, err
	}
	dto := toPipelineDTO(row)
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
		out = append(out, toPipelineDTO(&rows[i]))
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
	dto := toPipelineDTO(row)
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
	beaconUUID := req.BeaconUUID
	if beaconUUID == "" {
		beaconUUID = req.BeaconMac
	}
	if beaconUUID != "" {
		row.BeaconUUID = beaconUUID
	}
	row.Major = req.Major
	row.Minor = req.Minor
	lat := req.Latitude
	if lat == "" {
		lat = req.BeaconLat
	}
	if lat != "" {
		row.Latitude = lat
	}
	lng := req.Longitude
	if lng == "" {
		lng = req.BeaconLng
	}
	if lng != "" {
		row.Longitude = lng
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

	if strings.TrimSpace(req.Cookie) != "" {
		cookie, exp, err := s.resolveCookie(ctx, userID, req.Cookie)
		if err != nil {
			return nil, err
		}
		row.Cookie = cookie
		row.CookieExpiresAt = exp
	}

	if req.AutoCheckin && strings.TrimSpace(req.CheckinToken) != "" {
		token, exp, err := s.resolveCheckinToken(ctx, userID, req.CheckinToken)
		if err != nil {
			return nil, err
		}
		row.CheckinToken = token
		row.CheckinExpiresAt = exp
	}

	if err := dao.UpdatePipelineConfig(ctx, row); err != nil {
		return nil, err
	}
	dto := toPipelineDTO(row)
	return &dto, nil
}

// DeletePipelineConfig deletes a pipeline configuration.
func (s *Service) DeletePipelineConfig(ctx context.Context, userID uint64, id string) error {
	return dao.DeletePipelineConfig(ctx, id, userID)
}

func (s *Service) applyPipelineOverrides(ctx context.Context, id string, row *entity.PipelineConfig, overrideReq *do.RunPipelineRequest) {
	if overrideReq == nil {
		return
	}
	if strings.TrimSpace(overrideReq.Cookie) != "" {
		c, exp, err := s.resolveCookie(ctx, row.UserID, overrideReq.Cookie)
		if err == nil {
			row.Cookie = c
			row.CookieExpiresAt = exp
			_ = dao.UpdatePipelineCookie(ctx, id, c, exp)
		}
	}
	if strings.TrimSpace(overrideReq.CheckinToken) != "" {
		t, exp, err := s.resolveCheckinToken(ctx, row.UserID, overrideReq.CheckinToken)
		if err == nil {
			row.CheckinToken = t
			row.CheckinExpiresAt = exp
			_ = dao.UpdatePipelineCheckinToken(ctx, id, t, exp)
		}
	}
}

func (s *Service) validatePipelineAuth(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, id, execTime string) (*do.PipelineRunResult, bool) {
	if strings.TrimSpace(row.Cookie) == "" {
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

	if _, err := cli.ListLibraries(ctx, tpl, row.Cookie); err != nil {
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
		if strings.TrimSpace(row.CheckinToken) == "" {
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
		if _, err := cli.GetCheckInDevices(ctx, tpl, row.CheckinToken); err != nil {
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

func (s *Service) checkAndReserveSeat(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, id, execTime string) (*do.PipelineRunResult, string, bool) {
	layout, err := cli.GetLayout(ctx, tpl, row.Cookie, row.LibraryID)
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

	resOk, err := cli.ReserveSeat(ctx, tpl, row.Cookie, row.LibraryID, row.SeatKey)
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

func (s *Service) executeBeaconCheckin(ctx context.Context, cli *traceint.Client, tpl do.ProtocolTemplatesResponse, row *entity.PipelineConfig, id, execTime, reservationStatus string) *do.PipelineRunResult {
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

	lat, _ := strconv.ParseFloat(row.Latitude, 64)
	lng, _ := strconv.ParseFloat(row.Longitude, 64)
	signReq := do.CheckInSignRequest{
		BeaconUUID: row.BeaconUUID,
		Major:      row.Major,
		Minor:      row.Minor,
		Latitude:   lat,
		Longitude:  lng,
	}
	signResp, err := cli.SignCheckIn(ctx, tpl, row.CheckinToken, signReq, serverTime)
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

	s.applyPipelineOverrides(ctx, id, row, overrideReq)

	execTime := time.Now().Format("2006-01-02 15:04:05")
	tpl, err := s.templates(ctx, row.UserID)
	if err != nil {
		return nil, err
	}
	cli := s.api(ctx, row.UserID)

	if authErrRes, ok := s.validatePipelineAuth(ctx, cli, tpl, row, id, execTime); !ok {
		return authErrRes, nil
	}

	resErr, resStatus, ok := s.checkAndReserveSeat(ctx, cli, tpl, row, id, execTime)
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

	return s.executeBeaconCheckin(ctx, cli, tpl, row, id, execTime, resStatus), nil
}

// HelperVerifySession probes a cookie and lists accessible libraries.
func (s *Service) HelperVerifySession(ctx context.Context, userID uint64, rawInput string) ([]do.LibrarySummary, string, *time.Time, error) {
	cookie, exp, err := s.resolveCookie(ctx, userID, rawInput)
	if err != nil {
		return nil, "", nil, err
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

// HelperGetLibraryLayout fetches seat layout for a library using provided cookie.
func (s *Service) HelperGetLibraryLayout(ctx context.Context, userID uint64, cookie string, libID int) (*do.LibraryLayoutResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	layout, err := s.api(ctx, userID).GetLayout(ctx, tpl, cookie, libID)
	return layout, wrapTrace(err)
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

func toPipelineDTO(row *entity.PipelineConfig) do.PipelineConfigDTO {
	if row == nil {
		return do.PipelineConfigDTO{}
	}
	hasCookie := row.Cookie != ""
	hasCheckin := row.CheckinToken != ""
	return do.PipelineConfigDTO{
		ID:                row.ID,
		UserID:            row.UserID,
		Name:              row.Name,
		HasCookie:         hasCookie,
		CookieValid:       hasCookie,
		CookieMasked:      traceint.MaskCookie(row.Cookie),
		CookieExpiresAt:   row.CookieExpiresAt,
		LibraryID:         row.LibraryID,
		LibraryName:       row.LibraryName,
		Floor:             row.Floor,
		SeatKey:           row.SeatKey,
		SeatName:          row.SeatName,
		AutoCheckin:       row.AutoCheckin,
		HasCheckinToken:   hasCheckin,
		CheckinTokenValid: hasCheckin,
		CheckinExpiresAt:  row.CheckinExpiresAt,
		BeaconUUID:        row.BeaconUUID,
		BeaconMac:         row.BeaconUUID,
		Major:             row.Major,
		Minor:             row.Minor,
		Latitude:          row.Latitude,
		BeaconLat:         row.Latitude,
		Longitude:         row.Longitude,
		BeaconLng:         row.Longitude,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
