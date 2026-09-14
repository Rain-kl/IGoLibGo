// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"context"
	"net/http"
)

// GetReservation returns the current reservation.
func (s *Service) GetReservation(ctx context.Context, userID uint64) (*do.ReservationResponse, error) {
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	info, err := s.api(ctx, userID).GetReservation(ctx, tpl, cookie)
	return info, wrapTrace(err)
}

// RefreshReservation reloads reservation info from TraceInt.
func (s *Service) RefreshReservation(ctx context.Context, userID uint64) (*do.ReservationOperationResponse, error) {
	info, err := s.GetReservation(ctx, userID)
	if err != nil {
		return nil, err
	}
	msg := "当前没有预约"
	if info != nil && info.HasReservation {
		msg = "已刷新预约"
	}
	if info == nil {
		info = &do.ReservationResponse{}
	}
	return &do.ReservationOperationResponse{Reservation: *info, Message: msg}, nil
}

// CancelReservation cancels the current reservation.
func (s *Service) CancelReservation(ctx context.Context, userID uint64, req do.CancelReservationRequest) (*do.ReservationOperationResponse, error) {
	if req.StopOccupyFirst {
		_ = s.CancelTask(ctx, userID, consts.TaskKindOccupy)
	}
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	info, err := s.GetReservation(ctx, userID)
	if err != nil {
		return nil, err
	}
	if info == nil || !info.HasReservation || info.ReservationToken == "" {
		return nil, consts.NewError(http.StatusConflict, consts.CodeConflict, "当前没有可取消的预约")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	ok, err := s.api(ctx, userID).CancelReservation(ctx, tpl, cookie, info.ReservationToken)
	if err != nil {
		return nil, wrapTrace(err)
	}
	if !ok {
		return nil, consts.NewError(http.StatusBadGateway, consts.CodeTraceInt, "取消预约失败")
	}
	return &do.ReservationOperationResponse{
		Reservation: do.ReservationResponse{HasReservation: false},
		Message:     "已取消预约",
	}, nil
}
