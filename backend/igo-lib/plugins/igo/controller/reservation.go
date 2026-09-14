// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/gin-gonic/gin"
)

// GetReservation 当前预约
// @Summary 获取当前预约
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.ReservationResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/reservation [get]
func (ctrl *Controller) GetReservation(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.GetReservation(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}

// RefreshReservation 刷新预约
// @Summary 刷新预约
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.ReservationOperationResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/reservation/refresh [post]
func (ctrl *Controller) RefreshReservation(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.RefreshReservation(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}

// CancelReservation 取消预约（原 /api/reservation/cancel）
// @Summary 取消预约
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.CancelReservationRequest false "取消选项"
// @Success 200 {object} response.Any{data=do.ReservationOperationResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/reservation/cancel [post]
func (ctrl *Controller) CancelReservation(c *gin.Context) {
	var req do.CancelReservationRequest
	if c.Request.ContentLength > 0 {
		bound, ok := bindJSON[do.CancelReservationRequest](c)
		if !ok {
			return
		}
		req = bound
	}
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.CancelReservation(c.Request.Context(), userID, req)
		ctrl.reply(c, err)
	})
}
