// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListAccounts 账户列表
// @Summary 获取账户列表
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=[]do.AccountDTO}
// @Router /api/v1/igo/accounts [get]
func (ctrl *Controller) ListAccounts(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		list, err := ctrl.svc.ListAccounts(c.Request.Context(), userID)
		ctrl.jsonOK(c, list, err)
	})
}

// CreateAccount 创建账户
// @Summary 创建账户
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.CreateAccountRequest true "账户"
// @Success 201 {object} response.Any{data=do.AccountDTO}
// @Router /api/v1/igo/accounts [post]
func (ctrl *Controller) CreateAccount(c *gin.Context) {
	req, ok := bindJSON[do.CreateAccountRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.CreateAccount(c.Request.Context(), userID, req)
		if ctrl.reply(c, err) {
			return
		}
		response.Created(c, "/api/v1/igo/accounts/"+strconv.FormatUint(res.ID, 10), res)
	})
}

// GetAccount 账户详情
// @Summary 获取账户
// @Tags igo
// @Produce json
// @Param id path int true "账户 ID"
// @Success 200 {object} response.Any{data=do.AccountDTO}
// @Router /api/v1/igo/accounts/{id} [get]
func (ctrl *Controller) GetAccount(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetAccount(c.Request.Context(), userID, id)
		ctrl.jsonOK(c, res, err)
	})
}

// UpdateAccount 更新账户
// @Summary 更新账户
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "账户 ID"
// @Param request body do.UpdateAccountRequest true "账户"
// @Success 200 {object} response.Any{data=do.AccountDTO}
// @Router /api/v1/igo/accounts/{id} [put]
func (ctrl *Controller) UpdateAccount(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.UpdateAccountRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.UpdateAccount(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}

// DeleteAccount 删除账户
// @Summary 删除账户
// @Tags igo
// @Param id path int true "账户 ID"
// @Success 204 "无内容"
// @Router /api/v1/igo/accounts/{id} [delete]
func (ctrl *Controller) DeleteAccount(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.DeleteAccount(c.Request.Context(), userID, id)
		ctrl.noContent(c, err)
	})
}

// LoginAccount 写入占座凭据
// @Summary 账户登录授权
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "账户 ID"
// @Param request body do.AccountLoginRequest true "凭据"
// @Success 200 {object} response.Any{data=do.AccountDTO}
// @Router /api/v1/igo/accounts/{id}/login [post]
func (ctrl *Controller) LoginAccount(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.AccountLoginRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.LoginAccount(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}

// AuthorizeAccountCheckin 写入签到凭据
// @Summary 账户签到授权
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "账户 ID"
// @Param request body do.AccountCheckinAuthRequest true "凭据"
// @Success 200 {object} response.Any{data=do.AccountCheckinAuthResponse}
// @Router /api/v1/igo/accounts/{id}/checkin-auth [post]
func (ctrl *Controller) AuthorizeAccountCheckin(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.AccountCheckinAuthRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.AuthorizeAccountCheckin(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}
