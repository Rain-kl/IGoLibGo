// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListCheckInInfos 签到信息列表
// @Summary 获取签到信息列表
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=[]do.CheckInInfoDTO}
// @Router /api/v1/igo/checkin/infos [get]
func (ctrl *Controller) ListCheckInInfos(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		list, err := ctrl.svc.ListCheckInInfos(c.Request.Context(), userID)
		ctrl.jsonOK(c, list, err)
	})
}

// CreateCheckInInfo 创建签到信息
// @Summary 创建签到信息
// @Tags igo
// @Accept json
// @Produce json
// @Param request body do.CreateCheckInInfoRequest true "签到信息"
// @Success 201 {object} response.Any{data=do.CheckInInfoDTO}
// @Router /api/v1/igo/checkin/infos [post]
func (ctrl *Controller) CreateCheckInInfo(c *gin.Context) {
	req, ok := bindJSON[do.CreateCheckInInfoRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.CreateCheckInInfo(c.Request.Context(), userID, req)
		if ctrl.reply(c, err) {
			return
		}
		response.Created(c, "/api/v1/igo/checkin/infos/"+strconv.FormatUint(res.ID, 10), res)
	})
}

// GetCheckInInfo 签到信息详情
// @Summary 获取签到信息
// @Tags igo
// @Produce json
// @Param id path int true "签到信息 ID"
// @Success 200 {object} response.Any{data=do.CheckInInfoDTO}
// @Router /api/v1/igo/checkin/infos/{id} [get]
func (ctrl *Controller) GetCheckInInfo(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetCheckInInfo(c.Request.Context(), userID, id)
		ctrl.jsonOK(c, res, err)
	})
}

// UpdateCheckInInfo 更新签到信息
// @Summary 更新签到信息
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "签到信息 ID"
// @Param request body do.UpdateCheckInInfoRequest true "签到信息"
// @Success 200 {object} response.Any{data=do.CheckInInfoDTO}
// @Router /api/v1/igo/checkin/infos/{id} [put]
func (ctrl *Controller) UpdateCheckInInfo(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.UpdateCheckInInfoRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.UpdateCheckInInfo(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}

// DeleteCheckInInfo 删除签到信息
// @Summary 删除签到信息
// @Tags igo
// @Param id path int true "签到信息 ID"
// @Success 204 "无内容"
// @Router /api/v1/igo/checkin/infos/{id} [delete]
func (ctrl *Controller) DeleteCheckInInfo(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.DeleteCheckInInfo(c.Request.Context(), userID, id)
		ctrl.noContent(c, err)
	})
}

// SignCheckInInfo 按签到信息打卡
// @Summary 使用账户凭证与签到信息打卡
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "签到信息 ID"
// @Param request body do.SignCheckInInfoRequest true "打卡参数"
// @Success 200 {object} response.Any{data=do.CheckInSignResponse}
// @Router /api/v1/igo/checkin/infos/{id}/sign [post]
func (ctrl *Controller) SignCheckInInfo(c *gin.Context) {
	id, ok := parseUint64ID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.SignCheckInInfoRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.SignCheckInInfo(c.Request.Context(), userID, id, req)
		ctrl.jsonOK(c, res, err)
	})
}
