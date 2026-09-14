// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/gin-gonic/gin"
)

// GetCheckInSession 远程签到会话
// @Summary 获取签到会话
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.CheckInSessionResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/session [get]
func (ctrl *Controller) GetCheckInSession(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetCheckInSession(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetCheckInAuthQRCode 远程签到授权二维码
// @Summary 获取签到授权二维码
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.QRCodeResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/auth-qrcode [get]
func (ctrl *Controller) GetCheckInAuthQRCode(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetCheckInAuthQRCode(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// AuthorizeCheckInFromCode 从授权码登录远程签到
// @Summary 签到授权
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.CheckInAuthFromCodeRequest true "授权码"
// @Success 200 {object} response.Any{data=do.CheckInAuthorizationResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/from-code [post]
func (ctrl *Controller) AuthorizeCheckInFromCode(c *gin.Context) {
	req, ok := bindJSON[do.CheckInAuthFromCodeRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.AuthorizeCheckInFromCode(c.Request.Context(), userID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// GetCheckInDevices 签到设备与 Beacon
// @Summary 获取签到设备
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.CheckInDeviceResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/devices [get]
func (ctrl *Controller) GetCheckInDevices(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetCheckInDevices(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// SignCheckIn 提交蓝牙签到
// @Summary 远程签到
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.CheckInSignRequest true "签到参数"
// @Success 200 {object} response.Any{data=do.CheckInSignResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/sign [post]
func (ctrl *Controller) SignCheckIn(c *gin.Context) {
	req, ok := bindJSON[do.CheckInSignRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.SignCheckIn(c.Request.Context(), userID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// ClearCheckInSession 清除签到会话
// @Summary 退出签到会话
// @Tags igo
// @Success 204 "无内容"
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/checkin/session [delete]
func (ctrl *Controller) ClearCheckInSession(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.ClearCheckInSession(c.Request.Context(), userID)
		ctrl.noContent(c, err)
	})
}
