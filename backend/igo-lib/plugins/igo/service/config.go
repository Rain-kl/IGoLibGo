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
	"encoding/json"
)

// GetProtocolTemplates returns effective (default + override) templates.
func (s *Service) GetProtocolTemplates(ctx context.Context, userID uint64) (*do.ProtocolTemplatesResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// GetDefaultProtocolTemplates returns built-in templates.
func (s *Service) GetDefaultProtocolTemplates(_ context.Context, _ uint64) (*do.ProtocolTemplatesResponse, error) {
	tpl := traceint.DefaultTemplates()
	return &tpl, nil
}

// SaveProtocolTemplates stores per-user overrides.
func (s *Service) SaveProtocolTemplates(ctx context.Context, userID uint64, req do.SaveProtocolTemplatesRequest) (*do.ProtocolTemplatesResponse, error) {
	raw, err := json.Marshal(req.Overrides)
	if err != nil {
		return nil, err
	}
	if err := dao.UpsertProtocolOverride(ctx, &entity.ProtocolOverride{UserID: userID, Overrides: string(raw)}); err != nil {
		return nil, err
	}
	return s.GetProtocolTemplates(ctx, userID)
}

// ResetProtocolTemplates clears overrides.
func (s *Service) ResetProtocolTemplates(ctx context.Context, userID uint64) (*do.ProtocolTemplatesResponse, error) {
	if err := dao.UpsertProtocolOverride(ctx, &entity.ProtocolOverride{UserID: userID, Overrides: "{}"}); err != nil {
		return nil, err
	}
	return s.GetDefaultProtocolTemplates(ctx, userID)
}

// GetSettings returns migratable settings.
func (s *Service) GetSettings(ctx context.Context, userID uint64) (*do.SettingsResponse, error) {
	out := s.settings(ctx, userID)
	return &out, nil
}

// SaveSettings updates migratable settings.
func (s *Service) SaveSettings(ctx context.Context, userID uint64, req do.SaveSettingsRequest) (*do.SettingsResponse, error) {
	cur := s.settings(ctx, userID)
	applyInt := func(dst *int, src *int) {
		if src != nil {
			*dst = *src
		}
	}
	applyBool := func(dst *bool, src *bool) {
		if src != nil {
			*dst = *src
		}
	}
	applyStr := func(dst *string, src *string) {
		if src != nil {
			*dst = *src
		}
	}
	applyInt(&cur.RequestTimeoutSeconds, req.RequestTimeoutSeconds)
	applyInt(&cur.NetworkMaxRetries, req.NetworkMaxRetries)
	applyBool(&cur.TraceIntGraphQLOverridesEnabled, req.TraceIntGraphQLOverridesEnabled)
	applyStr(&cur.GrabReservationStrategy, req.GrabReservationStrategy)
	applyBool(&cur.OptimalGrabStrategyReminderEnabled, req.OptimalGrabStrategyReminderEnabled)
	applyStr(&cur.GrabScheduledStartDefault, req.GrabScheduledStartDefault)
	applyStr(&cur.TomorrowScheduledStartDefault, req.TomorrowScheduledStartDefault)
	applyInt(&cur.OccupyReReserveDelaySeconds, req.OccupyReReserveDelaySeconds)
	applyStr(&cur.OccupyCheckIntervalMode, req.OccupyCheckIntervalMode)
	applyInt(&cur.GlobalLeakScanIntervalSeconds, req.GlobalLeakScanIntervalSeconds)
	applyBool(&cur.AutoReleaseEnabled, req.AutoReleaseEnabled)
	applyInt(&cur.AutoReleaseLeadSeconds, req.AutoReleaseLeadSeconds)
	applyStr(&cur.HomeReservationProgressMode, req.HomeReservationProgressMode)
	raw, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}
	if err := dao.UpsertSettings(ctx, &entity.Settings{UserID: userID, Payload: string(raw)}); err != nil {
		return nil, err
	}
	return &cur, nil
}

// ExportBackup is not implemented in this stage.
func (s *Service) ExportBackup(_ context.Context, _ uint64, _ do.BackupExportRequest) (*do.BackupExportResponse, error) {
	return nil, consts.ErrNotImplemented
}

// ImportBackup is not implemented in this stage.
func (s *Service) ImportBackup(_ context.Context, _ uint64, _ do.BackupImportRequest) error {
	return consts.ErrNotImplemented
}
