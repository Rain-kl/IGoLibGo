// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Created(c, "/api/v1/items/123", map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "/api/v1/items/123", w.Header().Get("Location"))

	var body Response[map[string]string]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "123", body.Data["id"])
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.Bytes())
}

func TestPagedResponse(t *testing.T) {
	meta := Meta{Total: 100, Page: 1, PerPage: 10, TotalPages: 10}
	items := []string{"item1", "item2"}
	resp := Paged(items, meta)

	assert.Equal(t, items, resp.Data)
	assert.Equal(t, int64(100), resp.Meta.Total)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.PerPage)
	assert.Equal(t, 10, resp.Meta.TotalPages)
}

func TestAbortWithErrorCode_Middleware(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandlerMiddleware())
	r.GET("/validation", func(c *gin.Context) {
		AbortBadRequestWithCode(c, "validation_failed", "invalid payload", ErrorDetail{
			Field: "email",
			Issue: "must be valid email",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/validation", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "validation_failed", errResp.Error.Code)
	assert.Equal(t, "invalid payload", errResp.Error.Message)
	require.Len(t, errResp.Error.Details, 1)
	assert.Equal(t, "email", errResp.Error.Details[0].Field)
	assert.Equal(t, "must be valid email", errResp.Error.Details[0].Issue)
}
