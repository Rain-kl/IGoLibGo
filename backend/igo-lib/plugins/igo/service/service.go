// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package service implements IGoLibrary use cases.
package service

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Service is the igo business facade.
type Service struct {
	client *traceint.Client
	tasks  contracts.TaskService
	events Emitter
}

// New creates a Service with the default TraceInt HTTP client.
func New() *Service {
	return &Service{client: &traceint.Client{}}
}

// SetHTTPClient replaces the underlying HTTP client (tests).
func (s *Service) SetHTTPClient(c *http.Client) {
	if s.client == nil {
		s.client = &traceint.Client{}
	}
	s.client.HTTP = c
}

// SetTasks injects the platform task dispatcher.
func (s *Service) SetTasks(tasks contracts.TaskService) { s.tasks = tasks }

func wrapTrace(err error) error {
	if err == nil {
		return nil
	}
	var ce *consts.CodedError
	if errors.As(err, &ce) {
		return err
	}
	return consts.NewError(http.StatusBadGateway, consts.CodeTraceInt, err.Error())
}

func (s *Service) templates(ctx context.Context, userID uint64) (do.ProtocolTemplatesResponse, error) {
	base := traceint.DefaultTemplates()
	row, err := dao.GetProtocolOverride(ctx, userID)
	if err != nil {
		return base, err
	}
	if row == nil {
		return base, nil
	}
	return traceint.MergeTemplates(base, traceint.DecodeOverrides(row.Overrides)), nil
}

func (s *Service) cookie(ctx context.Context, userID uint64) (string, *entity.Session, error) {
	row, err := dao.GetSession(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if row == nil || row.Cookie == "" {
		return "", nil, consts.NewError(http.StatusConflict, consts.CodeSessionRequired, "请先完成 TraceInt 登录")
	}
	return row.Cookie, row, nil
}

func (s *Service) settings(ctx context.Context, userID uint64) do.SettingsResponse {
	out := defaultSettings()
	row, err := dao.GetSettings(ctx, userID)
	if err != nil || row == nil || row.Payload == "" {
		return out
	}
	_ = json.Unmarshal([]byte(row.Payload), &out)
	return out
}

func defaultSettings() do.SettingsResponse {
	return do.SettingsResponse{
		RequestTimeoutSeconds:              5,
		NetworkMaxRetries:                  3,
		GrabReservationStrategy:            "query_then_reserve",
		OptimalGrabStrategyReminderEnabled: true,
		OccupyReReserveDelaySeconds:        180,
		OccupyCheckIntervalMode:            "fixed_ten_seconds",
		GlobalLeakScanIntervalSeconds:      3,
	}
}

func toSessionResponse(row *entity.Session) *do.SessionResponse {
	if row == nil {
		return &do.SessionResponse{Authorized: false}
	}
	out := &do.SessionResponse{
		Authorized:     true,
		Source:         row.Source,
		SavedAt:        row.SavedAt.UTC().Format(time.RFC3339),
		CanAutoRestore: row.CanAutoRestore,
		CookieMasked:   traceint.MaskCookie(row.Cookie),
	}
	if exp := traceint.CookieExpiration(row.Cookie); exp != nil {
		out.ExpiresAt = exp.UTC().Format(time.RFC3339)
	} else if row.ExpiresAt != nil {
		out.ExpiresAt = row.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}

func toVenueSummary(v *entity.Venue) *do.LibrarySummary {
	if v == nil {
		return nil
	}
	return &do.LibrarySummary{
		LibraryID:   v.LibraryID,
		Name:        v.Name,
		Floor:       v.Floor,
		IsOpen:      v.IsOpen,
		TotalSeats:  v.TotalSeats,
		UsedSeats:   v.UsedSeats,
		BookedSeats: v.BookedSeats,
	}
}
