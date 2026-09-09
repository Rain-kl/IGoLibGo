// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package response provides shared HTTP API response structures adhering to RESTful api-design patterns.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Meta defines pagination and collection metadata according to api-design standards.
type Meta struct {
	Total      int64 `json:"total,omitempty"`
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ErrorDetail defines field-level validation or contextual error details.
type ErrorDetail struct {
	Field string `json:"field,omitempty"`
	Issue string `json:"issue"`
}

// ErrorBody defines the inner error payload according to api-design standards.
type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorResponse defines the standard error envelope with an "error" object.
type ErrorResponse struct {
	ErrorMsg string    `json:"error_msg,omitempty"` // Backwards compatibility with legacy clients
	Error    ErrorBody `json:"error"`
	Data     any       `json:"data"`
}

// Response defines the unified API response envelope with data payload.
type Response[T any] struct {
	Data     T          `json:"data"`
	ErrorMsg string     `json:"error_msg,omitempty"`
	Error    *ErrorBody `json:"error,omitempty"`
}

// PagedResponse defines collection response with metadata.
type PagedResponse[T any] struct {
	Data  T          `json:"data"`
	Meta  Meta       `json:"meta"`
	Error *ErrorBody `json:"error,omitempty"`
}

// Any 用于 Swagger 文档的通用成功响应类型
type Any struct {
	Data     any    `json:"data"`
	ErrorMsg string `json:"error_msg,omitempty" example:""`
}

// AnyError 用于 Swagger 文档的错误响应类型
type AnyError struct {
	Error ErrorBody `json:"error"`
}

// OK 构造 200 成功响应体
func OK[T any](data T) Response[T] {
	return Response[T]{Data: data}
}

// OKNil 构造 200 空数据成功响应体
func OKNil() Response[any] {
	return Response[any]{Data: nil}
}

// Paged 构造分页集合成功响应体
func Paged[T any](data T, meta Meta) PagedResponse[T] {
	return PagedResponse[T]{Data: data, Meta: meta}
}

// Created 写出 201 Created 成功响应并附带 Location 头
func Created[T any](c *gin.Context, location string, data T) {
	if location != "" {
		c.Header("Location", location)
	}
	c.JSON(http.StatusCreated, Response[T]{Data: data})
}

// NoContent 写出 204 No Content 成功响应
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

// Err 构造错误响应
func Err(msg string) ErrorResponse {
	return ErrorResponse{
		ErrorMsg: msg,
		Error: ErrorBody{
			Code:    "bad_request",
			Message: msg,
		},
		Data: nil,
	}
}

// ErrWithCode 构造带状态码与 Code 的错误响应
func ErrWithCode(errCode, msg string, details ...ErrorDetail) ErrorResponse {
	return ErrorResponse{
		ErrorMsg: msg,
		Error: ErrorBody{
			Code:    errCode,
			Message: msg,
			Details: details,
		},
		Data: nil,
	}
}
