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
	"net/http"
	"strconv"
	"strings"
)

// ListCheckInInfos returns check-in infos owned by the user.
func (s *Service) ListCheckInInfos(ctx context.Context, userID uint64) ([]do.CheckInInfoDTO, error) {
	rows, err := dao.ListCheckInInfosByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]do.CheckInInfoDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toCheckInInfoDTO(&rows[i]))
	}
	return out, nil
}

// GetCheckInInfo returns one check-in info owned by the user.
func (s *Service) GetCheckInInfo(ctx context.Context, userID, id uint64) (*do.CheckInInfoDTO, error) {
	row, err := dao.GetCheckInInfoByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该签到信息")
	}
	dto := toCheckInInfoDTO(row)
	return &dto, nil
}

// CreateCheckInInfo creates a check-in info. Beacon fields may be empty.
func (s *Service) CreateCheckInInfo(ctx context.Context, userID uint64, req do.CreateCheckInInfoRequest) (*do.CheckInInfoDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "签到信息名称不能为空")
	}
	uuid, err := normalizeOptionalBeacon(req.BeaconUUID, req.Major, req.Minor)
	if err != nil {
		return nil, err
	}
	row := &entity.CheckInInfo{
		UserID:     userID,
		Name:       name,
		BeaconUUID: uuid,
		Major:      req.Major,
		Minor:      req.Minor,
		Latitude:   strings.TrimSpace(req.Latitude),
		Longitude:  strings.TrimSpace(req.Longitude),
	}
	if err := dao.CreateCheckInInfo(ctx, row); err != nil {
		return nil, err
	}
	dto := toCheckInInfoDTO(row)
	return &dto, nil
}

// UpdateCheckInInfo updates a check-in info.
func (s *Service) UpdateCheckInInfo(ctx context.Context, userID, id uint64, req do.UpdateCheckInInfoRequest) (*do.CheckInInfoDTO, error) {
	row, err := dao.GetCheckInInfoByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该签到信息")
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		row.Name = name
	}
	uuid, err := normalizeOptionalBeacon(req.BeaconUUID, req.Major, req.Minor)
	if err != nil {
		return nil, err
	}
	if req.BeaconUUID != "" {
		row.BeaconUUID = uuid
	}
	row.Major = req.Major
	row.Minor = req.Minor
	if req.Latitude != "" {
		row.Latitude = strings.TrimSpace(req.Latitude)
	}
	if req.Longitude != "" {
		row.Longitude = strings.TrimSpace(req.Longitude)
	}
	if err := dao.UpdateCheckInInfo(ctx, row); err != nil {
		return nil, err
	}
	dto := toCheckInInfoDTO(row)
	return &dto, nil
}

// DeleteCheckInInfo removes a check-in info that is not referenced by pipeline cards.
func (s *Service) DeleteCheckInInfo(ctx context.Context, userID, id uint64) error {
	row, err := dao.GetCheckInInfoByUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该签到信息")
	}
	refs, err := dao.ListPipelineIDsByCheckInInfo(ctx, userID, id)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return consts.NewError(http.StatusConflict, consts.CodeConflict, "签到信息仍被一条龙引用: "+joinPipelineIDs(refs))
	}
	return dao.DeleteCheckInInfo(ctx, id, userID)
}

// SignCheckInInfo signs using the account token and the info Beacon parameters.
func (s *Service) SignCheckInInfo(ctx context.Context, userID, infoID uint64, req do.SignCheckInInfoRequest) (*do.CheckInSignResponse, error) {
	info, err := dao.GetCheckInInfoByUser(ctx, infoID, userID)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该签到信息")
	}
	acc, err := dao.GetAccountByUser(ctx, req.AccountID, userID)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	if strings.TrimSpace(acc.CheckinToken) == "" {
		return nil, &consts.NeedAuthError{
			Kind:      "CHECKIN",
			AccountID: acc.ID,
			AuthURL:   consts.WeChatCheckinAuthURL,
			Msg:       "签到凭据已失效，请更新该账户的签到凭证",
		}
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, acc.CheckinToken); err != nil {
		return nil, &consts.NeedAuthError{
			Kind:      "CHECKIN",
			AccountID: acc.ID,
			AuthURL:   consts.WeChatCheckinAuthURL,
			Msg:       "签到凭据已失效，请更新该账户的签到凭证",
		}
	}
	normUUID, ok := traceint.NormalizeUUID(info.BeaconUUID)
	if !ok || strings.TrimSpace(info.Latitude) == "" || strings.TrimSpace(info.Longitude) == "" {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请先补全签到信息的 Beacon 与坐标")
	}
	if info.Major < 0 || info.Major > 65535 || info.Minor < 0 || info.Minor > 65535 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Major / Minor 必须介于 0 和 65535 之间")
	}
	lat, err := strconv.ParseFloat(info.Latitude, 64)
	if err != nil {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请先补全签到信息的 Beacon 与坐标")
	}
	lng, err := strconv.ParseFloat(info.Longitude, 64)
	if err != nil {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请先补全签到信息的 Beacon 与坐标")
	}
	if req.ExpectedLibraryID <= 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "场馆编号无效")
	}
	ts, err := s.api(ctx, userID).GetCheckInServerTime(ctx, tpl)
	if err != nil {
		return nil, wrapTrace(err)
	}
	signReq := do.CheckInSignRequest{
		ExpectedLibraryID:   req.ExpectedLibraryID,
		ExpectedLibraryName: req.ExpectedLibraryName,
		BeaconUUID:          normUUID,
		Major:               info.Major,
		Minor:               info.Minor,
		Latitude:            lat,
		Longitude:           lng,
	}
	res, err := s.api(ctx, userID).SignCheckIn(ctx, tpl, acc.CheckinToken, signReq, ts)
	if err != nil {
		return nil, wrapTrace(err)
	}
	return res, nil
}

func normalizeOptionalBeacon(raw string, major, minor int) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	norm, ok := traceint.NormalizeUUID(raw)
	if !ok {
		return "", consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请填写有效的 32/36 位 Beacon UUID")
	}
	if major < 0 || major > 65535 || minor < 0 || minor > 65535 {
		return "", consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Major / Minor 必须介于 0 和 65535 之间")
	}
	return norm, nil
}

func toCheckInInfoDTO(row *entity.CheckInInfo) do.CheckInInfoDTO {
	if row == nil {
		return do.CheckInInfoDTO{}
	}
	return do.CheckInInfoDTO{
		ID:         row.ID,
		Name:       row.Name,
		BeaconUUID: row.BeaconUUID,
		Major:      row.Major,
		Minor:      row.Minor,
		Latitude:   row.Latitude,
		Longitude:  row.Longitude,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
