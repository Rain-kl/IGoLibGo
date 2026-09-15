// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListTasks 全部协调器状态
// @Summary 列出任务状态
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.TaskListResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/tasks [get]
func (ctrl *Controller) ListTasks(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.ListTasks(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// ListTaskRecords 任务启动历史（原 /api/task-records）
// ListTaskRecords 任务启动历史（原 /api/task-records）
// @Summary 任务启动记录
// @Tags igo
// @Produce json
// @Param kind query string false "任务类型 (grab / global-leak)"
// @Success 200 {object} response.Any{data=[]do.TaskLaunchRecord}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/task-records [get]
func (ctrl *Controller) ListTaskRecords(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.ListTaskRecords(c.Request.Context(), userID, c.Query("kind"))
		ctrl.jsonOK(c, res, err)
	})
}

// StartTask 启动任务（原 /api/tasks/{kind}/start）
// @Summary 启动任务
// @Tags igo
// @Accept json
// @Produce json
// @Param kind path string true "grab / occupy / global-leak / tomorrow"
// @Success 200 {object} response.Any{data=do.CoordinatorStatus}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/tasks/{kind}/start [post]
func (ctrl *Controller) StartTask(c *gin.Context) {
	kind := c.Param("kind")
	if _, ok := consts.SupportedTaskKinds[kind]; !ok {
		response.AbortBadRequestWithCode(c, consts.CodeInvalidTaskKind, "不支持的任务类型")
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		var err error
		switch kind {
		case consts.TaskKindGrab:
			req, ok := bindJSON[do.GrabStartRequest](c)
			if !ok {
				return
			}
			err = ctrl.svc.StartGrab(c.Request.Context(), userID, req)
		case consts.TaskKindOccupy:
			req, ok := bindJSON[do.OccupyStartRequest](c)
			if !ok {
				return
			}
			err = ctrl.svc.StartOccupy(c.Request.Context(), userID, req)
		case consts.TaskKindGlobalLeak:
			req, ok := bindJSON[do.GlobalLeakStartRequest](c)
			if !ok {
				return
			}
			err = ctrl.svc.StartGlobalLeak(c.Request.Context(), userID, req)
		case consts.TaskKindTomorrow:
			req, ok := bindJSON[do.TomorrowStartRequest](c)
			if !ok {
				return
			}
			err = ctrl.svc.StartTomorrow(c.Request.Context(), userID, req)
		}
		if ctrl.reply(c, err) {
			return
		}
		status, err := ctrl.svc.GetTaskStatus(c.Request.Context(), userID, kind)
		ctrl.jsonOK(c, status, err)
	})
}

// CancelTask 取消任务（原 /api/tasks/{kind}/cancel）
// @Summary 取消任务
// @Tags igo
// @Param kind path string true "grab / occupy / global-leak / tomorrow"
// @Success 200 {object} response.Any{data=do.CoordinatorStatus}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/tasks/{kind}/cancel [post]
func (ctrl *Controller) CancelTask(c *gin.Context) {
	kind := c.Param("kind")
	if _, ok := consts.SupportedTaskKinds[kind]; !ok {
		response.AbortBadRequestWithCode(c, consts.CodeInvalidTaskKind, "不支持的任务类型")
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.CancelTask(c.Request.Context(), userID, kind)
		if ctrl.reply(c, err) {
			return
		}
		status, err := ctrl.svc.GetTaskStatus(c.Request.Context(), userID, kind)
		ctrl.jsonOK(c, status, err)
	})
}

// RunTomorrowNow 立即执行一次明日预约
// @Summary 立即执行明日预约
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.TomorrowStartRequest true "明日预约计划"
// @Success 200 {object} response.Any{data=do.CoordinatorStatus}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/tasks/tomorrow/run-now [post]
func (ctrl *Controller) RunTomorrowNow(c *gin.Context) {
	req, ok := bindJSON[do.TomorrowStartRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.RunTomorrowNow(c.Request.Context(), userID, req)
		if ctrl.reply(c, err) {
			return
		}
		status, err := ctrl.svc.GetTaskStatus(c.Request.Context(), userID, consts.TaskKindTomorrow)
		ctrl.jsonOK(c, status, err)
	})
}

// GetGlobalLeakBlacklist 捡漏黑名单
// @Summary 获取捡漏黑名单
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.GlobalLeakBlacklistResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/global-leak/blacklist [get]
func (ctrl *Controller) GetGlobalLeakBlacklist(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetGlobalLeakBlacklist(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// SaveGlobalLeakBlacklist 保存捡漏黑名单
// @Summary 保存捡漏黑名单
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.SaveGlobalLeakBlacklistRequest true "黑名单"
// @Success 200 {object} response.Any{data=do.GlobalLeakBlacklistResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/global-leak/blacklist [put]
func (ctrl *Controller) SaveGlobalLeakBlacklist(c *gin.Context) {
	req, ok := bindJSON[do.SaveGlobalLeakBlacklistRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.SaveGlobalLeakBlacklist(c.Request.Context(), userID, req)
		ctrl.noContent(c, err)
	})
}

// GetGlobalLeakSelectedLibraries 捡漏扫描场馆
// @Summary 获取捡漏场馆
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=[]do.GlobalLeakLibraryTarget}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/global-leak/selected-libraries [get]
func (ctrl *Controller) GetGlobalLeakSelectedLibraries(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetGlobalLeakSelectedLibraries(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// SaveGlobalLeakSelectedLibraries 保存捡漏扫描场馆
// @Summary 保存捡漏场馆
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.SaveGlobalLeakSelectedLibrariesRequest true "场馆列表"
// @Success 200 {object} response.Any{data=[]do.GlobalLeakLibraryTarget}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/global-leak/selected-libraries [put]
func (ctrl *Controller) SaveGlobalLeakSelectedLibraries(c *gin.Context) {
	req, ok := bindJSON[do.SaveGlobalLeakSelectedLibrariesRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.SaveGlobalLeakSelectedLibraries(c.Request.Context(), userID, req)
		ctrl.noContent(c, err)
	})
}
