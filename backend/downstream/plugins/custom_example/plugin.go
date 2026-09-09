// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package custom_example demonstrates how to build a downstream Cordis plugin
// strictly adhering to the physical subpackage architecture and api-design RESTful standards.
// Copy this directory to create your own plugin.
package custom_example

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/downstream/plugins/custom_example/controller/hello"
	"Wavelet/downstream/plugins/custom_example/dao"
	"Wavelet/downstream/plugins/custom_example/service"
	"embed"

	"github.com/gin-gonic/gin"
)

//go:embed migrations/*/*.sql
var customMigrations embed.FS

// Plugin implements core.Plugin for the custom_example downstream plugin.
type Plugin struct {
	svc *service.HelloService
}

// New creates a new custom_example plugin.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the unique identifier for this plugin.
func (p *Plugin) Name() string {
	return "custom_example"
}

// Apply registers routes and services into the Cordis micro-kernel Context.
func (p *Plugin) Apply(ctx *core.Context) error {
	// 1. 注册 Goose 双方言嵌入式迁移
	ctx.Migrations().Register("custom_example", customMigrations)

	// 2. 绑定平台基础设施（DBService 等）
	ctx.Bind[contracts.DBService](dao.SetDBService)

	// 3. 初始化服务层并示范注册到微内核 IoC 容器（示范原生泛型方法 ctx.Provide）
	p.svc = service.NewHelloService()
	ctx.Provide[*service.HelloService](p.svc)

	// 4. 解析认证服务并挂载中间件（示范原生泛型方法 ctx.Inject）
	var authMW gin.HandlerFunc
	if authSvc, err := ctx.Inject[contracts.AuthService](); err == nil && authSvc != nil {
		if mw, ok := authSvc.RequireAuthMiddleware().(gin.HandlerFunc); ok {
			authMW = mw
		}
	}

	// 5. 挂载路由组（遵循物理子包分层与 api-design 规范）
	ctrl := hello.NewController(p.svc)
	group := ctx.Router().Group("/api/v1/custom/greetings")
	if authMW != nil {
		group.Use(authMW)
	}
	group.POST("", ctrl.CreateGreeting)
	group.GET("/:id", ctrl.GetGreeting)

	return nil
}
