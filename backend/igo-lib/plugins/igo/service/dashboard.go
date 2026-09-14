// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"context"
)

// GetDashboard returns the home snapshot.
func (s *Service) GetDashboard(ctx context.Context, userID uint64) (*do.DashboardResponse, error) {
	return s.snapshot(ctx, userID)
}

// GetStatus returns the compact status payload (original /api/status).
func (s *Service) GetStatus(ctx context.Context, userID uint64) (*do.DashboardResponse, error) {
	return s.snapshot(ctx, userID)
}

// ListActivityLogs is not persisted in this stage.
func (s *Service) ListActivityLogs(_ context.Context, _ uint64, _, _ int) ([]do.ActivityLogEntry, int64, error) {
	return []do.ActivityLogEntry{}, 0, nil
}

func (s *Service) snapshot(ctx context.Context, userID uint64) (*do.DashboardResponse, error) {
	sess, err := dao.GetSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := &do.DashboardResponse{
		Authorized:       sess != nil,
		HeroStatus:       "等待授权",
		HeroStatusDetail: "完成登录与场馆绑定后即可启用全部引擎",
		EngineSummary:    "等待授权",
		Reservation:      do.ReservationResponse{},
		Tasks:            []do.CoordinatorStatus{},
	}
	if sess != nil {
		out.HeroStatus = "已登录"
		out.HeroStatusDetail = "会话有效"
		out.EngineSummary = "空闲"
	}
	if v, err := dao.GetVenue(ctx, userID); err == nil {
		out.BoundLibrary = toVenueSummary(v)
	}
	if m, err := dao.GetDashboardMetrics(ctx, userID); err == nil && m != nil {
		out.HistoricalSuccessCount = m.HistoricalSuccessCount
		out.TotalGuardSeconds = m.TotalGuardSeconds
	}
	if sess != nil {
		if info, err := s.GetReservation(ctx, userID); err == nil && info != nil {
			out.Reservation = *info
			if info.HasReservation {
				out.HeroStatus = "已预约"
				out.HeroStatusDetail = info.SeatName
			}
		}
	}
	if tasks, err := s.ListTasks(ctx, userID); err == nil && tasks != nil {
		out.Tasks = tasks.Tasks
		for _, t := range tasks.Tasks {
			if t.IsActive {
				out.EngineSummary = t.Title + "运行中"
				break
			}
		}
	}
	return out, nil
}
