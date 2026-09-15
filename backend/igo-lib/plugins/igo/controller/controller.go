// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package controller provides HTTP handlers for the igo plugin.
package controller

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
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
	if user, err := ctrl.authSvc.GetCurrentUser(c.Request.Context()); err == nil && user != nil && user.ID > 0 {
		fn(user.ID)
		return
	}
	if user, err := ctrl.authSvc.GetCurrentUser(c); err == nil && user != nil && user.ID > 0 {
		fn(user.ID)
		return
	}
	if id, err := ctrl.authSvc.GetCurrentUserID(c); err == nil && id > 0 {
		fn(id)
		return
	}
	if id, err := ctrl.authSvc.GetCurrentUserID(c.Request.Context()); err == nil && id > 0 {
		fn(id)
		return
	}
	response.AbortUnauthorized(c, "未登录")
}

func (ctrl *Controller) reply(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var coded *consts.CodedError
	if errors.As(err, &coded) {
		response.AbortWithErrorCode(c, coded.Status, coded.Code, coded.Msg)
		return true
	}
	if errors.Is(err, consts.ErrNotImplemented) {
		response.AbortWithErrorCode(c, http.StatusNotImplemented, consts.CodeNotImplemented, "接口尚未实现")
		return true
	}
	if errors.Is(err, consts.ErrNoSession) {
		response.AbortWithErrorCode(c, http.StatusConflict, consts.CodeSessionRequired, "请先完成 TraceInt 登录")
		return true
	}
	if errors.Is(err, dao.ErrDBNotReady) {
		response.AbortWithErrorCode(c, http.StatusServiceUnavailable, "service_unavailable", "数据库未就绪")
		return true
	}
	response.AbortInternal(c, "内部系统错误")
	return true
}

func (ctrl *Controller) jsonOK(c *gin.Context, data any, err error) {
	if ctrl.reply(c, err) {
		return
	}
	c.JSON(http.StatusOK, response.OK(data))
}

func (ctrl *Controller) noContent(c *gin.Context, err error) {
	if ctrl.reply(c, err) {
		return
	}
	response.NoContent(c)
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

func bindOptionalJSON[T any](c *gin.Context) (T, bool) {
	var req T
	if c.Request.ContentLength == 0 {
		return req, true
	}
	return bindJSON[T](c)
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
	perPageStr := c.Query("per_page")
	if perPageStr == "" {
		perPageStr = c.Query("limit")
	}
	if perPageStr == "" {
		perPageStr = "20"
	}
	perPage, _ = strconv.Atoi(perPageStr)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
