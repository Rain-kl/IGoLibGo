// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDashboard 首页快照
// @Summary 首页仪表盘
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.DashboardResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/dashboard [get]
func (ctrl *Controller) GetDashboard(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetDashboard(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetStatus 兼容原 /api/status 的压缩状态
// @Summary 运行状态
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.DashboardResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/status [get]
func (ctrl *Controller) GetStatus(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetStatus(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// ListActivityLogs 活动日志
// @Summary 活动日志
// @Tags igo
// @Produce json
// @Param page query int false "页码"
// @Param per_page query int false "每页条数"
// @Success 200 {object} response.PagedResponse[[]do.ActivityLogEntry]
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/activity-logs [get]
func (ctrl *Controller) ListActivityLogs(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		page, perPage := parsePage(c)
		items, total, err := ctrl.svc.ListActivityLogs(c.Request.Context(), userID, page, perPage)
		if ctrl.reply(c, err) {
			return
		}
		c.JSON(http.StatusOK, response.Paged(items, response.Meta{Total: total, Page: page, PerPage: perPage}))
	})
}
