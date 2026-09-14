// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package controller provides HTTP handlers for the igo plugin.
package controller

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/service"
	"Wavelet/pkg/response"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Controller handles igo HTTP requests.
type Controller struct {
	svc     *service.Service
	authSvc contracts.AuthService
}

// New creates a Controller.
func New(svc *service.Service, authSvc contracts.AuthService) *Controller {
	return &Controller{svc: svc, authSvc: authSvc}
}

func (ctrl *Controller) withUser(c *gin.Context, fn func(userID uint64)) {
	id, err := ctrl.authSvc.GetCurrentUserID(c.Request.Context())
	if err != nil || id == 0 {
		response.AbortUnauthorized(c, "未登录")
		return
	}
	fn(id)
}

func (ctrl *Controller) reply(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, consts.ErrNotImplemented) {
		response.AbortWithErrorCode(c, http.StatusNotImplemented, consts.CodeNotImplemented, "接口尚未实现")
		return true
	}
	response.AbortInternal(c, "内部系统错误")
	return true
}

func bindJSON[T any](c *gin.Context) (T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequestWithCode(c, consts.CodeValidationError, "参数校验失败")
		var zero T
		return zero, false
	}
	return req, true
}

func parseLibraryID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.AbortBadRequestWithCode(c, consts.CodeInvalidID, "场馆 ID 格式错误")
		return 0, false
	}
	return id, true
}

func parsePage(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
