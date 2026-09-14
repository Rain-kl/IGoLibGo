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
	"time"
)

// GetCheckInSession returns the remote-check-in session.
func (s *Service) GetCheckInSession(ctx context.Context, userID uint64) (*do.CheckInSessionResponse, error) {
	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toCheckInSession(row), nil
}

// GetCheckInAuthQRCode returns the independent WeChat QR entry for check-in.
func (s *Service) GetCheckInAuthQRCode(ctx context.Context, userID uint64) (*do.QRCodeResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &do.QRCodeResponse{AuthURL: tpl.RemoteCheckInAuthorizationReturnURL}, nil
}

// AuthorizeCheckInFromCode exchanges a WeChat code for a check-in session.
func (s *Service) AuthorizeCheckInFromCode(ctx context.Context, userID uint64, req do.CheckInAuthFromCodeRequest) (*do.CheckInAuthorizationResponse, error) {
	code, ok := traceint.ExtractCode(req.Code)
	if !ok {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "签到授权链接中未找到 32 位 code")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	token, exp, err := s.api(ctx, userID).ExchangeCheckInCode(ctx, tpl, code)
	if err != nil {
		return nil, wrapTrace(err)
	}
	now := time.Now().UTC()
	row := &entity.CheckInSession{
		UserID:         userID,
		Token:          token,
		SavedAt:        now,
		ExpiresAt:      exp,
		CanAutoRestore: req.Remember,
	}
	if err := dao.UpsertCheckInSession(ctx, row); err != nil {
		return nil, err
	}
	device, devErr := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, token)
	out := &do.CheckInAuthorizationResponse{Session: *toCheckInSession(row), Device: device}
	if devErr != nil {
		out.DeviceRefreshWarning = devErr.Error()
	}
	return out, nil
}

// GetCheckInDevices returns profile and beacon UUIDs.
func (s *Service) GetCheckInDevices(ctx context.Context, userID uint64) (*do.CheckInDeviceResponse, error) {
	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeSessionRequired, "请先完成签到授权")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	info, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, row.Token)
	return info, wrapTrace(err)
}

// SignCheckIn submits a Bluetooth beacon check-in.
func (s *Service) SignCheckIn(ctx context.Context, userID uint64, req do.CheckInSignRequest) (*do.CheckInSignResponse, error) {
	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeSessionRequired, "请先完成签到授权")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	ts, err := s.api(ctx, userID).GetCheckInServerTime(ctx, tpl)
	if err != nil {
		return nil, wrapTrace(err)
	}
	res, err := s.api(ctx, userID).SignCheckIn(ctx, tpl, row.Token, req, ts)
	return res, wrapTrace(err)
}

// ClearCheckInSession drops the remote-check-in session.
func (s *Service) ClearCheckInSession(ctx context.Context, userID uint64) error {
	return dao.DeleteCheckInSession(ctx, userID)
}

func toCheckInSession(row *entity.CheckInSession) *do.CheckInSessionResponse {
	if row == nil {
		return &do.CheckInSessionResponse{Authorized: false}
	}
	out := &do.CheckInSessionResponse{
		Authorized:     true,
		SavedAt:        row.SavedAt.UTC().Format(time.RFC3339),
		CanAutoRestore: row.CanAutoRestore,
	}
	if row.ExpiresAt != nil {
		out.ExpiresAt = row.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}
