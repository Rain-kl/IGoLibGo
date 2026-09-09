// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APIError 统一的 API 业务错误类型，可被全局错误处理中间件捕获
type APIError struct {
	Code    int
	ErrCode string
	Msg     string
	Details []ErrorDetail
}

func (e *APIError) Error() string {
	return e.Msg
}

// NewError 实例化一个 APIError
func NewError(code int, msg string) *APIError {
	return &APIError{
		Code:    code,
		ErrCode: httpStatusToErrorCode(code),
		Msg:     msg,
	}
}

// NewErrorWithCode 实例化一个带自定义业务 code 与 details 的 APIError
func NewErrorWithCode(code int, errCode, msg string, details ...ErrorDetail) *APIError {
	return &APIError{
		Code:    code,
		ErrCode: errCode,
		Msg:     msg,
		Details: details,
	}
}

// httpStatusToErrorCode 映射 HTTP 状态码为 api-design 规范的 snake_case 错误 code
func httpStatusToErrorCode(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnprocessableEntity:
		return "unprocessable_entity"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	case http.StatusInternalServerError:
		return "internal_server_error"
	case http.StatusBadGateway:
		return "bad_gateway"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		return "unknown_error"
	}
}

// AbortWithError 将 API 错误挂载到 Gin Context 并中断执行流
func AbortWithError(c *gin.Context, code int, msg string) {
	_ = c.Error(NewError(code, msg))
	c.Abort()
}

// AbortWithErrorCode 将带语义化 Code 的 API 错误挂载到 Gin Context 并中断执行流
func AbortWithErrorCode(c *gin.Context, statusCode int, errCode, msg string, details ...ErrorDetail) {
	_ = c.Error(NewErrorWithCode(statusCode, errCode, msg, details...))
	c.Abort()
}

// AbortBadRequest 以 400 中断请求并将错误挂载到 Gin Error 链，供全局中间件统一记录 Trace 并响应。
func AbortBadRequest(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusBadRequest, msg)
}

// AbortBadRequestWithCode 以 400 中断请求并附带特定 code
func AbortBadRequestWithCode(c *gin.Context, errCode, msg string, details ...ErrorDetail) {
	AbortWithErrorCode(c, http.StatusBadRequest, errCode, msg, details...)
}

// AbortUnauthorized 以 401 中断请求并将错误挂载到 Gin Error 链。
func AbortUnauthorized(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusUnauthorized, msg)
}

// AbortForbidden 以 403 中断请求并将错误挂载到 Gin Error 链。
func AbortForbidden(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusForbidden, msg)
}

// AbortNotFound 以 404 中断请求并将错误挂载到 Gin Error 链。
func AbortNotFound(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusNotFound, msg)
}

// AbortInternal 以 500 中断请求并将错误挂载到 Gin Error 链。
func AbortInternal(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusInternalServerError, msg)
}

// AbortTooManyRequests 以 429 中断请求并将错误挂载到 Gin Error 链。
func AbortTooManyRequests(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusTooManyRequests, msg)
}

// AbortConflict 以 409 中断请求并将错误挂载到 Gin Error 链。
func AbortConflict(c *gin.Context, msg string) {
	AbortWithError(c, http.StatusConflict, msg)
}

// AbortUnprocessableEntity 以 422 中断请求（语义验证失败）
func AbortUnprocessableEntity(c *gin.Context, msg string, details ...ErrorDetail) {
	AbortWithErrorCode(c, http.StatusUnprocessableEntity, "unprocessable_entity", msg, details...)
}

// AbortNotFoundIfMissing 在 err 非空时中断请求：gorm.ErrRecordNotFound 映射为 404（文案 notFoundMsg），其余映射为 400（文案 err.Error()）。
// 返回是否已中断。err 为 nil 时不写响应并返回 false。
func AbortNotFoundIfMissing(c *gin.Context, err error, notFoundMsg string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		AbortNotFound(c, notFoundMsg)
		return true
	}
	AbortBadRequest(c, err.Error())
	return true
}

// AbortBadRequestOnError 在 err 非空时以 400 中断请求，文案为 err.Error()。
// 返回是否已中断。err 为 nil 时不写响应并返回 false。
func AbortBadRequestOnError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	AbortBadRequest(c, err.Error())
	return true
}
