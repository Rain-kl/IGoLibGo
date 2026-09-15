// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorHandlerMiddleware 捕获 c.Errors 并统一格式化为 JSON 返回给客户端，同时将其记录到 Span 异常中。
// 与 AbortWithError / AbortBadRequest 等配合使用，是全局 OTel 友好错误响应的唯一出口。
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		if apiErr, ok := errors.AsType[*APIError](err); ok {
			errCode := apiErr.ErrCode
			if errCode == "" {
				errCode = httpStatusToErrorCode(apiErr.Code)
			}
			c.JSON(apiErr.Code, ErrorResponse{
				Error: ErrorBody{
					Code:    errCode,
					Message: apiErr.Msg,
					Details: apiErr.Details,
				},
				Data: nil,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:    "internal_server_error",
				Message: "内部系统错误",
			},
			Data: nil,
		})
	}
}
