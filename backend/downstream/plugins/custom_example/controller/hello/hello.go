// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package hello provides HTTP API handlers for the custom_example plugin.
package hello

import (
	"Wavelet/downstream/plugins/custom_example/model/do"
	"Wavelet/downstream/plugins/custom_example/service"
	"Wavelet/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Controller handles HTTP requests for custom greetings according to api-design standards.
type Controller struct {
	svc *service.HelloService
}

// NewController creates a new Controller instance.
func NewController(svc *service.HelloService) *Controller {
	return &Controller{svc: svc}
}

// CreateGreeting 创建自定义问候语
// @Summary 创建自定义问候
// @Tags custom_example
// @Accept json
// @Produce json
// @Param request body do.CreateGreetingRequest true "问候请求"
// @Success 201 {object} response.Any{data=do.GreetingResponse} "创建成功"
// @Failure 400 {object} response.AnyError "参数校验失败"
// @Router /api/v1/custom/greetings [post]
func (ctrl *Controller) CreateGreeting(c *gin.Context) {
	var req do.CreateGreetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequestWithCode(c, "validation_error", "参数校验失败")
		return
	}

	res, err := ctrl.svc.CreateGreeting(c.Request.Context(), req)
	if err != nil {
		response.AbortInternal(c, "创建问候失败")
		return
	}

	response.Created(c, "/api/v1/custom/greetings/"+strconv.FormatInt(res.ID, 10), res)
}

// GetGreeting 获取问候详情
// @Summary 获取问候详情
// @Tags custom_example
// @Produce json
// @Param id path int true "问候 ID"
// @Success 200 {object} response.Any{data=do.GreetingResponse} "查询成功"
// @Failure 404 {object} response.AnyError "未找到"
// @Router /api/v1/custom/greetings/{id} [get]
func (ctrl *Controller) GetGreeting(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.AbortBadRequestWithCode(c, "invalid_id", "ID 格式错误")
		return
	}

	res, err := ctrl.svc.GetGreeting(c.Request.Context(), id)
	if err != nil {
		response.AbortInternal(c, "获取问候失败")
		return
	}
	if res == nil {
		response.AbortNotFound(c, "问候记录不存在")
		return
	}

	c.JSON(http.StatusOK, response.OK(res))
}
