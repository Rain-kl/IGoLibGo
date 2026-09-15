// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListPipelineConfigs returns all all-in-one automation cards owned by the current user.
// @Summary 获取一条龙配置列表
// @Tags IGo-Pipeline
// @Produce json
// @Success 200 {object} response.Response{data=[]do.PipelineConfigDTO}
// @Router /api/v1/igo/pipeline/configs [get]
func (ctrl *Controller) ListPipelineConfigs(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		list, err := ctrl.svc.ListPipelineConfigs(c.Request.Context(), userID)
		ctrl.jsonOK(c, list, err)
	})
}

// GetPipelineConfig retrieves one pipeline config card by ID.
// @Summary 获取单个一条龙配置
// @Tags IGo-Pipeline
// @Produce json
// @Param id path string true "配置 ID"
// @Success 200 {object} response.Response{data=do.PipelineConfigDTO}
// @Router /api/v1/igo/pipeline/configs/{id} [get]
func (ctrl *Controller) GetPipelineConfig(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		id := strings.TrimSpace(c.Param("id"))
		item, err := ctrl.svc.GetPipelineConfig(c.Request.Context(), userID, id)
		ctrl.jsonOK(c, item, err)
	})
}

// CreatePipelineConfig creates a new pipeline config card.
// @Summary 创建一条龙配置
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param body body do.CreatePipelineConfigRequest true "创建参数"
// @Success 201 {object} response.Response{data=do.PipelineConfigDTO}
// @Router /api/v1/igo/pipeline/configs [post]
func (ctrl *Controller) CreatePipelineConfig(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		req, ok := bindJSON[do.CreatePipelineConfigRequest](c)
		if !ok {
			return
		}
		item, err := ctrl.svc.CreatePipelineConfig(c.Request.Context(), userID, req)
		if ctrl.reply(c, err) {
			return
		}
		c.JSON(http.StatusCreated, response.OK(item))
	})
}

// UpdatePipelineConfig updates an existing pipeline config card.
// @Summary 更新一条龙配置
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param id path string true "配置 ID"
// @Param body body do.UpdatePipelineConfigRequest true "更新参数"
// @Success 200 {object} response.Response{data=do.PipelineConfigDTO}
// @Router /api/v1/igo/pipeline/configs/{id} [put]
func (ctrl *Controller) UpdatePipelineConfig(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		id := strings.TrimSpace(c.Param("id"))
		req, ok := bindJSON[do.UpdatePipelineConfigRequest](c)
		if !ok {
			return
		}
		item, err := ctrl.svc.UpdatePipelineConfig(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, item, err)
	})
}

// DeletePipelineConfig removes a pipeline config card.
// @Summary 删除一条龙配置
// @Tags IGo-Pipeline
// @Param id path string true "配置 ID"
// @Success 204 "No Content"
// @Router /api/v1/igo/pipeline/configs/{id} [delete]
func (ctrl *Controller) DeletePipelineConfig(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		id := strings.TrimSpace(c.Param("id"))
		err := ctrl.svc.DeletePipelineConfig(c.Request.Context(), userID, id)
		ctrl.noContent(c, err)
	})
}

// RunPipeline triggers execution of an all-in-one automation card.
// @Summary 立即执行一条龙任务
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param id path string true "配置 ID"
// @Param body body do.RunPipelineRequest false "覆盖凭据"
// @Success 200 {object} response.Response{data=do.PipelineRunResult}
// @Router /api/v1/igo/pipeline/configs/{id}/run [post]
func (ctrl *Controller) RunPipeline(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		id := strings.TrimSpace(c.Param("id"))
		var req *do.RunPipelineRequest
		if c.Request.ContentLength > 0 {
			var body do.RunPipelineRequest
			if err := c.ShouldBindJSON(&body); err == nil {
				req = &body
			}
		}
		res, err := ctrl.svc.RunPipeline(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}

// HelperVerifySession validates a raw cookie or authorization link and retrieves venues.
// @Summary 辅助接口：验证登录凭据并获取场馆列表
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param body body do.HelperVerifySessionRequest true "凭据参数"
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /api/v1/igo/pipeline/helper/verify-session [post]
func (ctrl *Controller) HelperVerifySession(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		req, ok := bindJSON[do.HelperVerifySessionRequest](c)
		if !ok {
			return
		}
		libs, cookie, exp, err := ctrl.svc.HelperVerifySession(c.Request.Context(), userID, req.Cookie)
		if ctrl.reply(c, err) {
			return
		}
		c.JSON(http.StatusOK, response.OK(gin.H{
			"libraries":  libs,
			"cookie":     cookie,
			"expires_at": exp,
		}))
	})
}

// HelperGetLibraryLayout loads the seat layout for a library using the supplied cookie.
// @Summary 辅助接口：获取场馆座位排布图
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param body body do.HelperLibraryLayoutRequest true "场馆与凭据"
// @Success 200 {object} response.Response{data=do.LibraryLayoutResponse}
// @Router /api/v1/igo/pipeline/helper/library-layout [post]
func (ctrl *Controller) HelperGetLibraryLayout(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		req, ok := bindJSON[do.HelperLibraryLayoutRequest](c)
		if !ok {
			return
		}
		layout, err := ctrl.svc.HelperGetLibraryLayout(c.Request.Context(), userID, req.Cookie, req.LibraryID)
		ctrl.jsonOK(c, layout, err)
	})
}

// HelperVerifyCheckin tests check-in authorization credentials and retrieves beacon devices.
// @Summary 辅助接口：验证签到凭据并获取设备列表
// @Tags IGo-Pipeline
// @Accept json
// @Produce json
// @Param body body do.HelperVerifyCheckinRequest true "签到凭据"
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /api/v1/igo/pipeline/helper/verify-checkin [post]
func (ctrl *Controller) HelperVerifyCheckin(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		req, ok := bindJSON[do.HelperVerifyCheckinRequest](c)
		if !ok {
			return
		}
		devs, token, exp, err := ctrl.svc.HelperVerifyCheckin(c.Request.Context(), userID, req.TokenOrCode)
		if ctrl.reply(c, err) {
			return
		}
		c.JSON(http.StatusOK, response.OK(gin.H{
			"device":     devs,
			"token":      token,
			"expires_at": exp,
		}))
	})
}
