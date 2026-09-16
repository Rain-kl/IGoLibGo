// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package igo is the IGoLibrary downstream Cordis plugin.
//
// Note: IGoLibrary 下游插件落在 backend/igo-lib — 见 .agents/notes/implemented/architecture/2026-09-14-igo-lib-plugin.md
package igo

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/core/extpoints"
	"Wavelet/igo-lib/plugins/igo/bot"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/controller"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/service"
	"embed"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

// MigrationsFS embeds migration SQL files for PostgreSQL and SQLite.
//
//go:embed migrations/*/*.sql
var MigrationsFS embed.FS

// Option configures the igo plugin.
type Option func(*Plugin)

// WithService allows injecting a custom service for testing.
func WithService(s *service.Service) Option {
	return func(p *Plugin) {
		p.svc = s
	}
}

// Plugin implements core.Plugin for IGoLibrary.
type Plugin struct {
	svc *service.Service
}

// New creates the igo plugin.
func New(opts ...Option) *Plugin {
	p := &Plugin{}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// Name returns the unique plugin identifier.
func (p *Plugin) Name() string {
	return consts.PluginName
}

// Apply registers migrations, DAO, and authenticated /api/v1/igo routes.
func (p *Plugin) Apply(ctx *core.Context) error {
	ctx.Migrations().Register(consts.PluginName, MigrationsFS)

	authSvc, err := ctx.Inject[contracts.AuthService]()
	if err != nil || authSvc == nil {
		return fmt.Errorf("igo: resolve AuthService: %w", err)
	}
	authMW, ok := authSvc.RequireAuthMiddleware().(gin.HandlerFunc)
	if !ok || authMW == nil {
		return errors.New("igo: AuthService.RequireAuthMiddleware is not gin.HandlerFunc")
	}

	if p.svc == nil {
		p.svc = service.New()
	}
	ctx.Bind(func(db contracts.DBService) {
		dao.SetDBService(db)
		if err := p.svc.BackfillAccountsAndCheckinInfos(ctx.GoContext()); err != nil {
			fmt.Printf("[igo] account backfill: %v\n", err)
		}
	})
	p.svc.SetEmitter(ctx.Events())
	if taskSvc, err := ctx.Inject[contracts.TaskService](); err == nil && taskSvc != nil {
		p.svc.SetTasks(taskSvc)
	}
	ctx.Bind(func(reg contracts.PushRegistry) {
		service.RegisterPushEvents(reg)
	})
	// Note: IGo 业务命令在 bot/ 注册，不进 msg_gateway — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md
	ctx.Bind(func(reg contracts.BotCommandRegistry) {
		bot.Register(reg, p.svc)
	})
	ctx.Task().Register(consts.TaskTypeTick, p.svc.HandleTick,
		extpoints.WithTaskType("igo_tick"),
		extpoints.WithTaskName("IGo 任务节拍"),
		extpoints.WithTaskDescription("抢座、占座、全域捡漏与明日预约的轮询节拍"),
		extpoints.WithTaskCategory("igo"),
		extpoints.WithTaskRetry(1),
		extpoints.WithTaskQueue("default"),
		extpoints.WithTaskRetryable(true),
	)
	ctx.Task().Register(consts.TaskTypeCookieWatch, p.svc.HandleCookieWatch,
		extpoints.WithTaskType("igo_cookie_watch"),
		extpoints.WithTaskName("IGo Cookie 到期扫描"),
		extpoints.WithTaskDescription("扫描即将过期的 TraceInt Cookie 并触发通知中心事件"),
		extpoints.WithTaskCategory("igo"),
		extpoints.WithTaskRetry(1),
		extpoints.WithTaskQueue("default"),
		extpoints.WithTaskRetryable(true),
	)
	ctx.Schedule().RegisterCron("*/5 * * * *", consts.TaskTypeCookieWatch, map[string]any{"action": "scan"})
	ctx.Provide[*service.Service](p.svc)
	ctrl := controller.New(p.svc, authSvc)

	g := ctx.Router().Group(consts.APIPrefix, authMW)

	g.GET("/dashboard", ctrl.GetDashboard)
	g.GET("/status", ctrl.GetStatus)
	g.GET("/activity-logs", ctrl.ListActivityLogs)

	g.GET("/session", ctrl.GetSession)
	g.GET("/session/auth-qrcode", ctrl.GetAuthQRCode)
	g.POST("/session/from-code", ctrl.AuthenticateFromCode)
	g.POST("/session/from-cookie", ctrl.AuthenticateFromCookie)
	g.POST("/session/cookie/refresh", ctrl.RefreshCookie)
	g.DELETE("/session", ctrl.SignOut)

	g.GET("/libraries", ctrl.ListLibraries)
	g.GET("/libraries/bound", ctrl.GetBoundLibrary)
	g.POST("/libraries/bound/refresh", ctrl.RefreshBoundLibrary)
	g.GET("/libraries/:id", ctrl.GetLibrary)
	g.GET("/libraries/:id/layout", ctrl.GetLibraryLayout)
	g.GET("/libraries/:id/rule", ctrl.GetLibraryRule)
	g.POST("/libraries/:id/bind", ctrl.BindLibrary)
	g.POST("/libraries/:id/preview", ctrl.PreviewLibrary)
	g.GET("/libraries/:id/favorites", ctrl.GetFavorites)
	g.PUT("/libraries/:id/favorites", ctrl.SaveFavorites)
	g.GET("/libraries/:id/seat-labels", ctrl.GetSeatLabels)
	g.PUT("/libraries/:id/seat-labels", ctrl.SetSeatLabels)
	g.DELETE("/libraries/:id/seat-labels", ctrl.DeleteSeatLabels)

	g.GET("/reservation", ctrl.GetReservation)
	g.POST("/reservation/refresh", ctrl.RefreshReservation)
	g.POST("/reservation/cancel", ctrl.CancelReservation)

	g.GET("/tasks", ctrl.ListTasks)
	g.GET("/task-records", ctrl.ListTaskRecords)
	g.POST("/tasks/tomorrow/run-now", ctrl.RunTomorrowNow)
	g.POST("/tasks/:kind/start", ctrl.StartTask)
	g.POST("/tasks/:kind/cancel", ctrl.CancelTask)

	g.GET("/global-leak/blacklist", ctrl.GetGlobalLeakBlacklist)
	g.PUT("/global-leak/blacklist", ctrl.SaveGlobalLeakBlacklist)
	g.GET("/global-leak/selected-libraries", ctrl.GetGlobalLeakSelectedLibraries)
	g.PUT("/global-leak/selected-libraries", ctrl.SaveGlobalLeakSelectedLibraries)

	g.GET("/accounts", ctrl.ListAccounts)
	g.POST("/accounts", ctrl.CreateAccount)
	g.GET("/accounts/:id", ctrl.GetAccount)
	g.PUT("/accounts/:id", ctrl.UpdateAccount)
	g.DELETE("/accounts/:id", ctrl.DeleteAccount)
	g.POST("/accounts/:id/login", ctrl.LoginAccount)
	g.POST("/accounts/:id/checkin-auth", ctrl.AuthorizeAccountCheckin)

	g.GET("/checkin/infos", ctrl.ListCheckInInfos)
	g.POST("/checkin/infos", ctrl.CreateCheckInInfo)
	g.GET("/checkin/infos/:id", ctrl.GetCheckInInfo)
	g.PUT("/checkin/infos/:id", ctrl.UpdateCheckInInfo)
	g.DELETE("/checkin/infos/:id", ctrl.DeleteCheckInInfo)
	g.POST("/checkin/infos/:id/sign", ctrl.SignCheckInInfo)

	g.GET("/checkin/session", ctrl.GetCheckInSession)
	g.GET("/checkin/auth-qrcode", ctrl.GetCheckInAuthQRCode)
	g.POST("/checkin/from-code", ctrl.AuthorizeCheckInFromCode)
	g.GET("/checkin/devices", ctrl.GetCheckInDevices)
	g.POST("/checkin/sign", ctrl.SignCheckIn)
	g.DELETE("/checkin/session", ctrl.ClearCheckInSession)
	g.GET("/checkin/profiles", ctrl.ListCheckInVenueProfiles)
	g.GET("/checkin/profiles/:id", ctrl.GetCheckInVenueProfile)
	g.PUT("/checkin/profiles/:id", ctrl.SaveCheckInVenueProfile)

	g.GET("/pipeline/configs", ctrl.ListPipelineConfigs)
	g.POST("/pipeline/configs", ctrl.CreatePipelineConfig)
	g.GET("/pipeline/configs/:id", ctrl.GetPipelineConfig)
	g.PUT("/pipeline/configs/:id", ctrl.UpdatePipelineConfig)
	g.DELETE("/pipeline/configs/:id", ctrl.DeletePipelineConfig)
	g.POST("/pipeline/configs/:id/run", ctrl.RunPipeline)
	g.POST("/pipeline/verify-session", ctrl.HelperVerifySession)
	g.POST("/pipeline/library-layout", ctrl.HelperGetLibraryLayout)
	g.POST("/pipeline/verify-checkin", ctrl.HelperVerifyCheckin)

	g.GET("/protocol/templates", ctrl.GetProtocolTemplates)
	g.GET("/protocol/templates/defaults", ctrl.GetDefaultProtocolTemplates)
	g.PUT("/protocol/templates", ctrl.SaveProtocolTemplates)
	g.POST("/protocol/templates/reset", ctrl.ResetProtocolTemplates)

	g.GET("/settings", ctrl.GetSettings)
	g.PUT("/settings", ctrl.SaveSettings)

	g.POST("/backup/export", ctrl.ExportBackup)
	g.POST("/backup/import", ctrl.ImportBackup)

	return nil
}
