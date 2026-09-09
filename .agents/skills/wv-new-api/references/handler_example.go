// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package references

import (
	"fmt"

	"Wavelet/pkg/response"
	"github.com/gin-gonic/gin"
)

// createChannelRequest 客户端请求体 DTO
type createChannelRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// createChannelResponse API 响应体 DTO
type createChannelResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateChannel 示例：插件内 HTTP Handler（位于 plugins/domain/channel/controller/channel.go）
// @Summary 创建频道
// @Description 示例：符合 api-design 规范的 RESTful POST 接口
// @Tags channel
// @Accept json
// @Produce json
// @Param request body createChannelRequest true "业务请求参数"
// @Success 201 {object} response.Any{data=createChannelResponse} "创建成功"
// @Failure 400 {object} response.AnyError "参数校验失败"
// @Failure 401 {object} response.AnyError "未登录"
// @Router /api/v1/channels [post]
func CreateChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequestWithCode(c, "validation_error", "参数校验失败")
		return
	}

	// 从请求上下文中提取认证用户
	userID := int64(9527)

	result, err := CreateChannelLogic(c.Request.Context(), userID, req.Name)
	if err != nil {
		response.AbortBadRequest(c, err.Error())
		return
	}

	// POST 创建资源符合 api-design：返回 201 Created 与 Location 响应头
	response.Created(c, fmt.Sprintf("/api/v1/channels/%d", result.ID), createChannelResponse{
		ID:   result.ID,
		Name: result.Name,
	})
}
