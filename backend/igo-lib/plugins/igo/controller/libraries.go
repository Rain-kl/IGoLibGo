// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/gin-gonic/gin"
)

// ListLibraries 账号下场馆列表
// @Summary 列出场馆
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=[]do.LibrarySummary}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries [get]
func (ctrl *Controller) ListLibraries(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.ListLibraries(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetBoundLibrary 当前锁定场馆
// @Summary 获取锁定场馆
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.BoundLibraryResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/bound [get]
func (ctrl *Controller) GetBoundLibrary(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetBoundLibrary(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// RefreshBoundLibrary 刷新锁定场馆
// @Summary 刷新锁定场馆
// @Tags igo
// @Produce json
// @Success 200 {object} response.Any{data=do.BoundLibraryResponse}
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/bound/refresh [post]
func (ctrl *Controller) RefreshBoundLibrary(c *gin.Context) {
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.RefreshBoundLibrary(c.Request.Context(), userID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetLibrary 单个场馆摘要
// @Summary 获取场馆
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=do.LibrarySummary}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id} [get]
func (ctrl *Controller) GetLibrary(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetLibrary(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetLibraryLayout 座位布局
// @Summary 获取场馆布局
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=do.LibraryLayoutResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/layout [get]
func (ctrl *Controller) GetLibraryLayout(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetLibraryLayout(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetLibraryRule 场馆规则
// @Summary 获取场馆规则
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=do.LibraryRuleResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/rule [get]
func (ctrl *Controller) GetLibraryRule(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetLibraryRule(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// BindLibrary 锁定场馆
// @Summary 锁定场馆
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=do.BoundLibraryResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/bind [post]
func (ctrl *Controller) BindLibrary(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.BindLibrary(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// PreviewLibrary 预览场馆（不锁定）
// @Summary 预览场馆
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=do.LibraryLayoutResponse}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/preview [post]
func (ctrl *Controller) PreviewLibrary(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.PreviewLibrary(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetFavorites 收藏座位
// @Summary 获取收藏座位
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=[]do.SeatRef}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/favorites [get]
func (ctrl *Controller) GetFavorites(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetFavorites(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// GetSeatLabels 座位标签
// @Summary 获取座位标签
// @Tags igo
// @Produce json
// @Param id path int true "场馆 ID"
// @Success 200 {object} response.Any{data=[]do.SeatLabel}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/seat-labels [get]
func (ctrl *Controller) GetSeatLabels(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.GetSeatLabels(c.Request.Context(), userID, libraryID)
		ctrl.jsonOK(c, res, err)
	})
}

// SaveFavorites 保存收藏座位
// @Summary 保存收藏座位
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "场馆 ID"
// @Param request body do.SaveFavoritesRequest true "收藏列表"
// @Success 200 {object} response.Any{data=[]do.SeatRef}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/favorites [put]
func (ctrl *Controller) SaveFavorites(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.SaveFavoritesRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.SaveFavorites(c.Request.Context(), userID, libraryID, req.Seats)
		ctrl.noContent(c, err)
	})
}

// SetSeatLabels 设置座位标签
// @Summary 设置座位标签
// @Tags igo
// @Accept json
// @Produce json
// @Param id path int true "场馆 ID"
// @Param request body do.SetSeatLabelsRequest true "标签"
// @Success 200 {object} response.Any{data=[]do.SeatLabel}
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/seat-labels [put]
func (ctrl *Controller) SetSeatLabels(c *gin.Context) {
	libraryID, ok := parseLibraryID(c)
	if !ok {
		return
	}
	req, ok := bindJSON[do.SetSeatLabelsRequest](c)
	if !ok {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		res, err := ctrl.svc.SetSeatLabels(c.Request.Context(), userID, libraryID, req)
		ctrl.jsonOK(c, res, err)
	})
}

// DeleteSeatLabels 删除座位标签
// @Summary 删除座位标签
// @Tags igo
// @Accept json
// @Param id path int true "场馆 ID"
// @Param request body do.DeleteSeatLabelsRequest true "座位 key"
// @Success 204 "无内容"
// @Failure 400 {object} response.AnyError
// @Failure 401 {object} response.AnyError
// @Failure 501 {object} response.AnyError
// @Router /api/v1/igo/libraries/{id}/seat-labels [delete]
func (ctrl *Controller) DeleteSeatLabels(c *gin.Context) {
	req, ok := bindJSON[do.DeleteSeatLabelsRequest](c)
	if !ok {
		return
	}
	libraryID, valid := parseLibraryID(c)
	if !valid {
		return
	}
	ctrl.withUser(c, func(userID uint64) {
		err := ctrl.svc.DeleteSeatLabels(c.Request.Context(), userID, libraryID, req.SeatKeys)
		ctrl.noContent(c, err)
	})
}
