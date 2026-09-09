// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package custom_example_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/downstream/plugins/custom_example"
	"Wavelet/downstream/plugins/custom_example/controller/hello"
	"Wavelet/downstream/plugins/custom_example/service"
	"Wavelet/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	contracts.AuthService
}

func (m *mockAuthService) RequireAuthMiddleware() any {
	return func(c *gin.Context) {
		c.Next()
	}
}

func TestCustomExamplePlugin_Apply(t *testing.T) {
	ctx := core.NewContext(context.Background())
	core.Provide[contracts.AuthService](ctx, &mockAuthService{})

	p := custom_example.New()
	assert.Equal(t, "custom_example", p.Name())

	err := p.Apply(ctx)
	require.NoError(t, err)
}

func TestCustomExample_Controller(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := service.NewHelloService()
	ctrl := hello.NewController(svc)

	engine := gin.New()
	engine.Use(response.ErrorHandlerMiddleware())
	engine.POST("/api/v1/custom/greetings", ctrl.CreateGreeting)
	engine.GET("/api/v1/custom/greetings/:id", ctrl.GetGreeting)

	// 1. Create greeting valid
	reqBody := `{"recipient":"Alice","message":"Welcome!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/custom/greetings", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/api/v1/custom/greetings/")
	assert.Contains(t, w.Body.String(), `"recipient":"Alice"`)

	// 2. Create greeting invalid (validation error -> 400)
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/custom/greetings", strings.NewReader(`{}`))
	badReq.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	engine.ServeHTTP(wBad, badReq)

	assert.Equal(t, http.StatusBadRequest, wBad.Code)
	assert.Contains(t, wBad.Body.String(), `"validation_error"`)

	// 3. Get greeting non-existent -> 404
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/custom/greetings/9999", nil)
	wGet := httptest.NewRecorder()
	engine.ServeHTTP(wGet, getReq)

	assert.Equal(t, http.StatusNotFound, wGet.Code)
	assert.Contains(t, wGet.Body.String(), `"not_found"`)
}
