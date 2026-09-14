// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/gin-gonic/gin"
)

// GetSession 当前 TraceInt Cookie 会话
// @Summary 获取会话
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.SessionResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session [get]
func (ctrl *Controller) GetSession(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.GetSession(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}

// GetAuthQRCode 微信授权二维码
// @Summary 获取授权二维码
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.QRCodeResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session/auth-qrcode [get]
func (ctrl *Controller) GetAuthQRCode(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.GetAuthQRCode(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}

// AuthenticateFromCode 从微信授权码/链接登录
// @Summary 从授权码登录
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.AuthFromCodeRequest true "授权码"
// @Success 200 {object} response.Any{data=do.SessionWorkflowResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session/from-code [post]
func (ctrl *Controller) AuthenticateFromCode(c *gin.Context) {
	req, ok := bindJSON[do.AuthFromCodeRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.AuthenticateFromCode(c.Request.Context(), userID, req)
		ctrl.reply(c, err)
	})
}

// AuthenticateFromCookie 从原始 Cookie 登录
// @Summary 从 Cookie 登录
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.AuthFromCookieRequest true "Cookie"
// @Success 200 {object} response.Any{data=do.SessionWorkflowResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session/from-cookie [post]
func (ctrl *Controller) AuthenticateFromCookie(c *gin.Context) {
	req, ok := bindJSON[do.AuthFromCookieRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.AuthenticateFromCookie(c.Request.Context(), userID, req)
		ctrl.reply(c, err)
	})
}

// RefreshCookie 刷新 Cookie
// @Summary 刷新 Cookie
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.SessionWorkflowResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session/cookie/refresh [post]
func (ctrl *Controller) RefreshCookie(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		_, err := ctrl.svc.RefreshCookie(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}

// SignOut 退出 TraceInt 会话
// @Summary 退出会话
// @Tags igo
// @Success 204 "无内容"
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/session [delete]
func (ctrl *Controller) SignOut(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.SignOut(c.Request.Context(), userID)
		ctrl.reply(c, err)
	})
}
