// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package igo_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/core/extpoints"
	"Wavelet/igo-lib/plugins/igo"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	contracts.AuthService
}

type mockPushRegistry struct {
	events []contracts.PushEventMeta
}

func (m *mockPushRegistry) RegisterBuiltInEvent(meta contracts.PushEventMeta) {
	m.events = append(m.events, meta)
}

func (m *mockPushRegistry) SyncEvents(context.Context) error { return nil }

func (m *mockAuthService) RequireAuthMiddleware() any {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()
	})
}

func (m *mockAuthService) GetCurrentUserID(context.Context) (uint64, error) {
	return 1, nil
}

func (m *mockAuthService) GetCurrentUser(context.Context) (*contracts.UserDTO, error) {
	return &contracts.UserDTO{ID: 1, Username: "test_user"}, nil
}

var expectedRoutes = []struct {
	method string
	path   string
}{
	{"GET", "/api/v1/igo/dashboard"},
	{"GET", "/api/v1/igo/status"},
	{"GET", "/api/v1/igo/activity-logs"},
	{"GET", "/api/v1/igo/session"},
	{"GET", "/api/v1/igo/session/auth-qrcode"},
	{"POST", "/api/v1/igo/session/from-code"},
	{"POST", "/api/v1/igo/session/from-cookie"},
	{"POST", "/api/v1/igo/session/cookie/refresh"},
	{"DELETE", "/api/v1/igo/session"},
	{"GET", "/api/v1/igo/libraries"},
	{"GET", "/api/v1/igo/libraries/bound"},
	{"POST", "/api/v1/igo/libraries/bound/refresh"},
	{"GET", "/api/v1/igo/libraries/:id"},
	{"GET", "/api/v1/igo/libraries/:id/layout"},
	{"GET", "/api/v1/igo/libraries/:id/rule"},
	{"POST", "/api/v1/igo/libraries/:id/bind"},
	{"POST", "/api/v1/igo/libraries/:id/preview"},
	{"GET", "/api/v1/igo/libraries/:id/favorites"},
	{"PUT", "/api/v1/igo/libraries/:id/favorites"},
	{"GET", "/api/v1/igo/libraries/:id/seat-labels"},
	{"PUT", "/api/v1/igo/libraries/:id/seat-labels"},
	{"DELETE", "/api/v1/igo/libraries/:id/seat-labels"},
	{"GET", "/api/v1/igo/reservation"},
	{"POST", "/api/v1/igo/reservation/refresh"},
	{"POST", "/api/v1/igo/reservation/cancel"},
	{"GET", "/api/v1/igo/tasks"},
	{"GET", "/api/v1/igo/task-records"},
	{"POST", "/api/v1/igo/tasks/tomorrow/run-now"},
	{"POST", "/api/v1/igo/tasks/:kind/start"},
	{"POST", "/api/v1/igo/tasks/:kind/cancel"},
	{"GET", "/api/v1/igo/global-leak/blacklist"},
	{"PUT", "/api/v1/igo/global-leak/blacklist"},
	{"GET", "/api/v1/igo/global-leak/selected-libraries"},
	{"PUT", "/api/v1/igo/global-leak/selected-libraries"},
	{"GET", "/api/v1/igo/accounts"},
	{"POST", "/api/v1/igo/accounts"},
	{"GET", "/api/v1/igo/accounts/:id"},
	{"PUT", "/api/v1/igo/accounts/:id"},
	{"DELETE", "/api/v1/igo/accounts/:id"},
	{"POST", "/api/v1/igo/accounts/:id/login"},
	{"POST", "/api/v1/igo/accounts/:id/checkin-auth"},
	{"GET", "/api/v1/igo/checkin/infos"},
	{"POST", "/api/v1/igo/checkin/infos"},
	{"GET", "/api/v1/igo/checkin/infos/:id"},
	{"PUT", "/api/v1/igo/checkin/infos/:id"},
	{"DELETE", "/api/v1/igo/checkin/infos/:id"},
	{"POST", "/api/v1/igo/checkin/infos/:id/sign"},
	{"GET", "/api/v1/igo/checkin/session"},
	{"GET", "/api/v1/igo/checkin/auth-qrcode"},
	{"POST", "/api/v1/igo/checkin/from-code"},
	{"GET", "/api/v1/igo/checkin/devices"},
	{"POST", "/api/v1/igo/checkin/sign"},
	{"DELETE", "/api/v1/igo/checkin/session"},
	{"GET", "/api/v1/igo/checkin/profiles"},
	{"GET", "/api/v1/igo/checkin/profiles/:id"},
	{"PUT", "/api/v1/igo/checkin/profiles/:id"},
	{"GET", "/api/v1/igo/pipeline/configs"},
	{"POST", "/api/v1/igo/pipeline/configs"},
	{"GET", "/api/v1/igo/pipeline/configs/:id"},
	{"PUT", "/api/v1/igo/pipeline/configs/:id"},
	{"DELETE", "/api/v1/igo/pipeline/configs/:id"},
	{"POST", "/api/v1/igo/pipeline/configs/:id/run"},
	{"POST", "/api/v1/igo/pipeline/verify-session"},
	{"POST", "/api/v1/igo/pipeline/library-layout"},
	{"POST", "/api/v1/igo/pipeline/verify-checkin"},
	{"GET", "/api/v1/igo/protocol/templates"},
	{"GET", "/api/v1/igo/protocol/templates/defaults"},
	{"PUT", "/api/v1/igo/protocol/templates"},
	{"POST", "/api/v1/igo/protocol/templates/reset"},
	{"GET", "/api/v1/igo/settings"},
	{"PUT", "/api/v1/igo/settings"},
	{"POST", "/api/v1/igo/backup/export"},
	{"POST", "/api/v1/igo/backup/import"},
}

type recordingBotReg struct {
	cmds  []string
	convs []string
}

func (r *recordingBotReg) Register(cmd contracts.BotCommand) error {
	r.cmds = append(r.cmds, cmd.Name())
	return nil
}

func (r *recordingBotReg) RegisterConversation(conv contracts.BotConversation) error {
	r.convs = append(r.convs, conv.Name())
	return nil
}

func applyPlugin(t *testing.T) *core.Context {
	t.Helper()
	ctx := core.NewContext(context.Background())
	ctx.Provide[contracts.AuthService](&mockAuthService{})
	ctx.Provide[contracts.PushRegistry](&mockPushRegistry{})
	p := igo.New()
	require.Equal(t, consts.PluginName, p.Name())
	require.NoError(t, p.Apply(ctx))
	return ctx
}

func TestPluginRegistersBotCommands(t *testing.T) {
	ctx := core.NewContext(context.Background())
	ctx.Provide[contracts.AuthService](&mockAuthService{})
	ctx.Provide[contracts.PushRegistry](&mockPushRegistry{})
	reg := &recordingBotReg{}
	ctx.Provide[contracts.BotCommandRegistry](reg)
	require.NoError(t, igo.New().Apply(ctx))
	assert.Equal(t, []string{"show", "run"}, reg.cmds)
	assert.Equal(t, []string{"igo.login_auth", "igo.checkin_auth"}, reg.convs)
}

func asGinHandler(t *testing.T, h any, where string) gin.HandlerFunc {
	t.Helper()
	switch fn := h.(type) {
	case gin.HandlerFunc:
		return fn
	case func(*gin.Context):
		return fn
	default:
		t.Fatalf("%s is %T, not a gin handler", where, h)
		return nil
	}
}

func mountRoutes(t *testing.T, routes []extpoints.RouteDefinition) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(response.ErrorHandlerMiddleware())
	for _, rt := range routes {
		handlers := make([]gin.HandlerFunc, 0, len(rt.Middlewares)+len(rt.Handlers))
		for _, mw := range rt.Middlewares {
			handlers = append(handlers, asGinHandler(t, mw, rt.Method+" "+rt.Path+" middleware"))
		}
		for _, h := range rt.Handlers {
			handlers = append(handlers, asGinHandler(t, h, rt.Method+" "+rt.Path+" handler"))
		}
		engine.Handle(rt.Method, rt.Path, handlers...)
	}
	return engine
}

func TestPlugin_RegistersExpectedRoutes(t *testing.T) {
	ctx := applyPlugin(t)
	routes := ctx.Router().Routes()
	got := make(map[string]struct{}, len(routes))
	for _, rt := range routes {
		got[rt.Method+" "+rt.Path] = struct{}{}
	}
	assert.Len(t, routes, len(expectedRoutes))
	for _, want := range expectedRoutes {
		_, ok := got[want.method+" "+want.path]
		assert.True(t, ok, "missing route %s %s", want.method, want.path)
	}
}

func TestPlugin_ApplyRequiresAuthService(t *testing.T) {
	ctx := core.NewContext(context.Background())
	err := igo.New().Apply(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AuthService")
}

func TestPlugin_RegistersPushEvents(t *testing.T) {
	reg := &mockPushRegistry{}
	ctx := core.NewContext(context.Background())
	ctx.Provide[contracts.AuthService](&mockAuthService{})
	ctx.Provide[contracts.PushRegistry](reg)
	require.NoError(t, igo.New().Apply(ctx))

	keys := map[string]struct{}{}
	for _, ev := range reg.events {
		keys[ev.Key] = struct{}{}
		assert.NotEmpty(t, ev.Name)
		assert.NotEmpty(t, ev.DefaultTemplate.Content)
	}
	for _, key := range []string{
		consts.PushGrabSucceeded,
		consts.PushOccupySucceeded,
		consts.PushGlobalLeakSucceeded,
		consts.PushTomorrowSucceeded,
		consts.PushTaskFailed,
		consts.PushCookieExpiring,
		consts.PushSessionInvalid,
	} {
		_, ok := keys[key]
		assert.True(t, ok, "missing push event %s", key)
	}
}

func TestPlugin_RegistersMigrations(t *testing.T) {
	ctx := applyPlugin(t)
	entry, ok := ctx.Migrations().Get(consts.PluginName)
	require.True(t, ok)
	require.NotNil(t, entry.FS)

	matches, err := fs.Glob(entry.FS, "migrations/*/*.sql")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"migrations/postgres/00001_initial.sql",
		"migrations/postgres/00002_pipeline_configs.sql",
		"migrations/postgres/00003_accounts_and_checkin_infos.sql",
		"migrations/sqlite/00001_initial.sql",
		"migrations/sqlite/00002_pipeline_configs.sql",
		"migrations/sqlite/00003_accounts_and_checkin_infos.sql",
	}, matches)
}

func TestPlugin_NotImplementedEnvelope(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/igo/backup/export", strings.NewReader(`{"password":"12345678"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)

	var body response.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, consts.CodeNotImplemented, body.Error.Code)
	assert.Equal(t, "接口尚未实现", body.Error.Message)
	assert.Nil(t, body.Data)
}

func TestPlugin_OriginalPathStatusEnvelope(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/igo/status", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"service_unavailable"`)
}

func TestPlugin_ValidationErrorEnvelope(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/igo/session/from-code", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body response.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, consts.CodeValidationError, body.Error.Code)
}

func TestPlugin_InvalidTaskKind(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/igo/tasks/foo/start", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), consts.CodeInvalidTaskKind)
}

func TestPlugin_InvalidLibraryID(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/igo/libraries/abc/layout", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), consts.CodeInvalidID)
}

func TestPlugin_ValidStartBodyStillNotImplemented(t *testing.T) {
	ctx := applyPlugin(t)
	engine := mountRoutes(t, ctx.Router().Routes())

	body := `{"library_id":1,"seats":[{"seat_key":"A-1","seat_name":"A1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/igo/tasks/grab/start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "service_unavailable")
}

func TestPlugin_DashboardAuthenticationResolvesUser(t *testing.T) {
	t.Run("authenticated user reaches dashboard logic without 401", func(t *testing.T) {
		ctx := applyPlugin(t)
		engine := mountRoutes(t, ctx.Router().Routes())

		req := httptest.NewRequest(http.MethodGet, "/api/v1/igo/dashboard", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Without DB initialized in this mock context, it should hit service_unavailable (503), NEVER 401 Unauthorized!
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		coreCtx := core.NewContext(context.Background())
		unauthMock := &unauthenticatedAuthService{}
		core.Provide[contracts.AuthService](coreCtx, unauthMock)

		p := igo.New()
		require.NoError(t, p.Apply(coreCtx))
		engine := mountRoutes(t, coreCtx.Router().Routes())

		req := httptest.NewRequest(http.MethodGet, "/api/v1/igo/dashboard", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "未登录")
	})
}

type unauthenticatedAuthService struct {
	contracts.AuthService
}

func (m *unauthenticatedAuthService) RequireAuthMiddleware() any {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()
	})
}

func (m *unauthenticatedAuthService) GetCurrentUserID(context.Context) (uint64, error) {
	return 0, errors.New("unauthorized")
}

func (m *unauthenticatedAuthService) GetCurrentUser(context.Context) (*contracts.UserDTO, error) {
	return nil, errors.New("unauthorized")
}
