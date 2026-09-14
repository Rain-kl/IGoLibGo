// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/gin-gonic/gin"
)

// GetProtocolTemplates 当前协议模板（默认 + 覆盖）
// @Summary 获取协议模板
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.ProtocolTemplatesResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/protocol/templates [get]
func (ctrl *Controller) GetProtocolTemplates(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetProtocolTemplates(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetDefaultProtocolTemplates 内置协议模板
// @Summary 获取默认协议模板
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.ProtocolTemplatesResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/protocol/templates/defaults [get]
func (ctrl *Controller) GetDefaultProtocolTemplates(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetDefaultProtocolTemplates(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// SaveProtocolTemplates 保存协议覆盖
// @Summary 保存协议模板
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.SaveProtocolTemplatesRequest true "覆盖"
// @Success 200 {object} response.Any{data=do.ProtocolTemplatesResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/protocol/templates [put]
func (ctrl *Controller) SaveProtocolTemplates(c *gin.Context) {
	req, ok := bindJSON[do.SaveProtocolTemplatesRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.SaveProtocolTemplates(c.Request.Context(), userID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// ResetProtocolTemplates 重置协议覆盖
// @Summary 重置协议模板
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.ProtocolTemplatesResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/protocol/templates/reset [post]
func (ctrl *Controller) ResetProtocolTemplates(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.ResetProtocolTemplates(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetSettings 可迁移系统设置
// @Summary 获取设置
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.SettingsResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/settings [get]
func (ctrl *Controller) GetSettings(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetSettings(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// SaveSettings 保存可迁移系统设置
// @Summary 保存设置
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.SaveSettingsRequest true "设置"
// @Success 200 {object} response.Any{data=do.SettingsResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/settings [put]
func (ctrl *Controller) SaveSettings(c *gin.Context) {
	req, ok := bindJSON[do.SaveSettingsRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.SaveSettings(c.Request.Context(), userID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// ExportBackup 导出加密备份
// @Summary 导出备份
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.BackupExportRequest true "密码"
// @Success 200 {object} response.Any{data=do.BackupExportResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/backup/export [post]
func (ctrl *Controller) ExportBackup(c *gin.Context) {
	req, ok := bindJSON[do.BackupExportRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.ExportBackup(c.Request.Context(), userID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// ImportBackup 导入加密备份
// @Summary 导入备份
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.BackupImportRequest true "备份"
// @Success 204 "无内容"
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/backup/import [post]
func (ctrl *Controller) ImportBackup(c *gin.Context) {
	req, ok := bindJSON[do.BackupImportRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.ImportBackup(c.Request.Context(), userID, req)
		ctrl.noContent(c, err)
	})
}
