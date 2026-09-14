// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"net/http"
	"strings"
	"time"

	"Wavelet/igo-lib/plugins/igo/consts"
)

// GetSession returns the TraceInt cookie session.
func (s *Service) GetSession(ctx context.Context, userID uint64) (*do.SessionResponse, error) {
	row, err := dao.GetSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toSessionResponse(row), nil
}

// GetAuthQRCode returns the WeChat authorization entry URL.
func (s *Service) GetAuthQRCode(ctx context.Context, userID uint64) (*do.QRCodeResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &do.QRCodeResponse{
		ImageDataURL: defaultQRCodeDataURL(),
		AuthURL:      tpl.CookieAuthorizationReturnURL,
	}, nil
}

// AuthenticateFromCode exchanges a WeChat code/link for a cookie session.
func (s *Service) AuthenticateFromCode(ctx context.Context, userID uint64, req do.AuthFromCodeRequest) (*do.SessionWorkflowResponse, error) {
	code, ok := traceint.ExtractCode(req.Code)
	if !ok {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "授权链接中未找到 32 位 code")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	cookie, err := s.api(ctx, userID).GetCookie(ctx, tpl, code)
	if err != nil {
		return nil, wrapTrace(err)
	}
	return s.persistCookie(ctx, userID, cookie, "qr_code", req.Remember)
}

// AuthenticateFromCookie stores and validates a raw cookie.
// If the input is detected to be a WeChat authorization link or code, it automatically exchanges it.
func (s *Service) AuthenticateFromCookie(ctx context.Context, userID uint64, req do.AuthFromCookieRequest) (*do.SessionWorkflowResponse, error) {
	raw := strings.TrimSpace(req.Cookie)
	if code, ok := traceint.ExtractCode(raw); ok && (!strings.Contains(raw, "Authorization=") || strings.HasPrefix(raw, "http")) {
		return s.AuthenticateFromCode(ctx, userID, do.AuthFromCodeRequest{
			Code:     code,
			Remember: req.Remember,
		})
	}
	return s.persistCookie(ctx, userID, req.Cookie, "manual_cookie", req.Remember)
}

// RefreshCookie re-validates the stored cookie, or exchanges a new WeChat code when provided.
func (s *Service) RefreshCookie(ctx context.Context, userID uint64, req do.RefreshCookieRequest) (*do.SessionWorkflowResponse, error) {
	if strings.TrimSpace(req.Code) != "" {
		return s.AuthenticateFromCode(ctx, userID, do.AuthFromCodeRequest{Code: req.Code, Remember: true})
	}
	cookie, row, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	remember := true
	source := "refresh"
	if row != nil {
		remember = row.CanAutoRestore
		source = row.Source
	}
	return s.persistCookie(ctx, userID, cookie, source, remember)
}

// SignOut clears the TraceInt session.
func (s *Service) SignOut(ctx context.Context, userID uint64) error {
	return dao.DeleteSession(ctx, userID)
}

func (s *Service) persistCookie(ctx context.Context, userID uint64, cookie, source string, remember bool) (*do.SessionWorkflowResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	libs, err := s.api(ctx, userID).ListLibraries(ctx, tpl, cookie)
	if err != nil {
		return nil, wrapTrace(err)
	}
	now := time.Now().UTC()
	exp := traceint.CookieExpiration(cookie)
	row := &entity.Session{
		UserID:         userID,
		Cookie:         cookie,
		Source:         source,
		SavedAt:        now,
		ExpiresAt:      exp,
		CanAutoRestore: remember,
	}
	if remember {
		if err := dao.UpsertSession(ctx, row); err != nil {
			return nil, err
		}
	} else {
		_ = dao.DeleteSession(ctx, userID)
	}
	stored, _ := dao.GetSession(ctx, userID)
	if stored == nil {
		stored = row
	}
	return &do.SessionWorkflowResponse{
		Session:   *toSessionResponse(stored),
		Libraries: libs,
		Message:   "登录成功",
	}, nil
}
