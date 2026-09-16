// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package igo_test

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	igoentity "Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/service"
	"Wavelet/pkg/idgen"
	"Wavelet/pkg/response"
	"Wavelet/pkg/testhelper"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type userSwitchableAuthService struct {
	contracts.AuthService
	mu     sync.RWMutex
	userID uint64
}

func (m *userSwitchableAuthService) SetUserID(id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userID = id
}

func (m *userSwitchableAuthService) RequireAuthMiddleware() any {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()
	})
}

func (m *userSwitchableAuthService) GetCurrentUserID(context.Context) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.userID, nil
}

func (m *userSwitchableAuthService) GetCurrentUser(context.Context) (*contracts.UserDTO, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return &contracts.UserDTO{ID: m.userID, Username: fmt.Sprintf("user_%d", m.userID)}, nil
}

type stubDBService struct {
	db *gorm.DB
}

func (s stubDBService) DB(context.Context) *gorm.DB {
	return s.db
}

func (s stubDBService) GORM() *gorm.DB {
	return s.db
}

func (s stubDBService) Named(string) *gorm.DB {
	return s.db
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = r.target.Scheme
	req2.URL.Host = r.target.Host
	if r.base == nil {
		return http.DefaultTransport.RoundTrip(req2)
	}
	return r.base.RoundTrip(req2)
}

func rewriteHost(targetURL string) http.RoundTripper {
	u, _ := url.Parse(targetURL)
	return &rewriteTransport{target: u}
}

// TestPipeline_EndToEndUserFlow tests the full user journey:
// 1. Create pipeline configuration card
// 2. Query list and verify details
// 3. Update configuration
// 4. Run pipeline (Success scenario with seat reservation + beacon checkin)
// 5. Run pipeline when seat is occupied (Graceful abort)
// 6. Run pipeline when session expired (Returns need_auth: LOGIN)
// 7. Re-authorize and run to success
// 8. Multi-tenant isolation between user 101 and user 102
// 9. Delete configuration
func TestPipeline_EndToEndUserFlow(t *testing.T) {
	_ = idgen.Init(1)
	db, _, cleanup := testhelper.SetupTestEnvironment(t)
	defer cleanup()

	require.NoError(t, db.AutoMigrate(
		&igoentity.PipelineConfig{},
		&igoentity.ProtocolOverride{},
		&igoentity.Settings{},
		&igoentity.Session{},
		&igoentity.Venue{},
		&igoentity.Account{},
		&igoentity.CheckInInfo{},
		&igoentity.CheckInSession{},
	))

	dao.SetDBService(stubDBService{db: db})
	t.Cleanup(func() { dao.SetDBService(nil) })

	// Mock TraceInt & WeChat server
	seatStatusOccupied := false
	cookieValid := true

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// WeChat code exchange to cookie
		if strings.Contains(r.URL.Path, "urlSign") || strings.Contains(r.URL.RawQuery, "code=") {
			http.SetCookie(w, &http.Cookie{
				Name:    "Authorization",
				Value:   "traceint-auth-cookie-valid-12345",
				Expires: time.Now().Add(24 * time.Hour),
			})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}

		if strings.Contains(r.URL.Path, "devices") {
			_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`))
			return
		}
		if strings.Contains(r.URL.Path, "getTime") || strings.Contains(r.URL.Path, "currentTime") {
			_, _ = w.Write([]byte("1700000000"))
			return
		}
		if strings.Contains(r.URL.Path, "sign") {
			_, _ = w.Write([]byte(`{"code":0,"msg":"打卡成功","data":{"status":1}}`))
			return
		}

		// GraphQL API
		if strings.Contains(r.URL.Path, "graphql") {
			if !cookieValid {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"errors":[{"message":"user not login"}]}`))
				return
			}

			// Check if this is a Reserve query
			var bodyMap map[string]any
			_ = json.NewDecoder(r.Body).Decode(&bodyMap)
			q, _ := bodyMap["query"].(string)

			if strings.Contains(q, "userAuth") && strings.Contains(q, "reserve") {
				respJSON := fmt.Sprintf(`{"data":{"userAuth":{"reserve":{"libs":[
					{"lib_id":10,"lib_name":"总馆一楼","lib_floor":"1","is_open":true,"lib_layout":{"seats":[{"key":"S-101","name":"101号","type":1,"status":%t}]}},
					{"lib_id":20,"lib_name":"总馆二楼","lib_floor":"2","is_open":true,"lib_layout":{"seats":[{"key":"S-201","name":"201号","type":1,"status":false}]}}
				],"reserueSeat":true}}}}`, seatStatusOccupied)
				_, _ = w.Write([]byte(respJSON))
				return
			}

			if strings.Contains(q, "userAuth") && strings.Contains(q, "rule") {
				_, _ = w.Write([]byte(`{"data":{"userAuth":{"rule":{"signRule":{"rules":[{"library":{"lib_id":10},"open_time":28800,"close_time":79200,"validate_time":900}]}}}}}`))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	authMock := &userSwitchableAuthService{userID: 101}
	coreCtx := core.NewContext(context.Background())
	core.Provide[contracts.AuthService](coreCtx, authMock)
	core.Provide[contracts.DBService](coreCtx, stubDBService{db: db})

	dao.SetDBService(stubDBService{db: db})
	t.Cleanup(func() { dao.SetDBService(nil) })

	svc := service.New()
	svc.SetHTTPClient(&http.Client{Timeout: 5 * time.Second, Transport: rewriteHost(srv.URL)})

	igoPlugin := igo.New(igo.WithService(svc))
	require.NoError(t, igoPlugin.Apply(coreCtx))

	// Mount router
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(response.ErrorHandlerMiddleware())
	for _, r := range coreCtx.Router().Routes() {
		handlers := make([]gin.HandlerFunc, 0, len(r.Middlewares)+len(r.Handlers))
		for _, mw := range r.Middlewares {
			switch fn := mw.(type) {
			case gin.HandlerFunc:
				handlers = append(handlers, fn)
			case func(*gin.Context):
				handlers = append(handlers, fn)
			}
		}
		for _, h := range r.Handlers {
			switch fn := h.(type) {
			case gin.HandlerFunc:
				handlers = append(handlers, fn)
			case func(*gin.Context):
				handlers = append(handlers, fn)
			}
		}
		if len(handlers) > 0 {
			engine.Handle(r.Method, r.Path, handlers...)
		}
	}

	// ==========================================
	// 1. User 101 creates a pipeline card: my_exam_seat
	// ==========================================
	accBody := `{
		"name": "考研号",
		"cookie": "Authorization=traceint-auth-cookie-valid-12345",
		"checkin_token": "valid-checkin-token-xyz"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/igo/accounts", strings.NewReader(accBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var accResp struct {
		Data do.AccountDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &accResp))

	infoBody := `{
		"name": "考研签到",
		"beacon_uuid": "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		"major": 10001,
		"minor": 1984,
		"latitude": "30.123456",
		"longitude": "120.123456"
	}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/checkin/infos", strings.NewReader(infoBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var infoResp struct {
		Data do.CheckInInfoDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &infoResp))

	createBody := fmt.Sprintf(`{
		"id": "my_exam_seat",
		"name": "考研专座01",
		"account_id": "%s",
		"checkin_info_id": "%s",
		"library_id": 10,
		"library_name": "总馆一楼",
		"floor": "1",
		"seat_key": "S-101",
		"seat_name": "101号",
		"auto_checkin": true
	}`, strconv.FormatUint(accResp.Data.ID, 10), strconv.FormatUint(infoResp.Data.ID, 10))

	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var createResp struct {
		Data do.PipelineConfigDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createResp))
	assert.Equal(t, "my_exam_seat", createResp.Data.ID)
	assert.Equal(t, "考研专座01", createResp.Data.Name)
	assert.Equal(t, "S-101", createResp.Data.SeatKey)
	assert.True(t, createResp.Data.AutoCheckin)

	// ==========================================
	// 2. List configs for User 101
	// ==========================================
	req = httptest.NewRequest(http.MethodGet, "/api/v1/igo/pipeline/configs", nil)
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var listResp struct {
		Data []do.PipelineConfigDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	require.Len(t, listResp.Data, 1)
	assert.Equal(t, "my_exam_seat", listResp.Data[0].ID)

	// ==========================================
	// 3. Update config name to "考研专座-VIP"
	// ==========================================
	updateBody := `{
		"name": "考研专座-VIP",
		"auto_checkin": true
	}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/igo/pipeline/configs/my_exam_seat", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var updateResp struct {
		Data do.PipelineConfigDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updateResp))
	assert.Equal(t, "考研专座-VIP", updateResp.Data.Name)

	// ==========================================
	// 4. Run pipeline (Success scenario)
	// ==========================================
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs/my_exam_seat/run", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var runResp struct {
		Data do.PipelineRunResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &runResp))
	assert.True(t, runResp.Data.Success)
	assert.Contains(t, runResp.Data.ReservationStatus, "成功预约")
	assert.Contains(t, runResp.Data.CheckinStatus, "打卡成功")

	// ==========================================
	// 5. Run pipeline when seat is occupied (Graceful abort)
	// ==========================================
	seatStatusOccupied = true // mark S-101 occupied
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs/my_exam_seat/run", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &runResp))
	assert.False(t, runResp.Data.Success)
	assert.Contains(t, runResp.Data.Message, "已被占用")
	seatStatusOccupied = false // reset

	// ==========================================
	// 6. Run pipeline when session expired
	// ==========================================
	cookieValid = false // simulate session expiration
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs/my_exam_seat/run", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &runResp))
	assert.False(t, runResp.Data.Success)
	assert.Equal(t, "LOGIN", runResp.Data.NeedAuth)
	assert.Contains(t, runResp.Data.AuthURL, "open.weixin.qq.com")

	// ==========================================
	// 7. Re-authorize with new cookie and run successfully
	// ==========================================
	cookieValid = true
	overrideBody := `{
		"cookie": "Authorization=traceint-new-reauthed-cookie"
	}`
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs/my_exam_seat/run", strings.NewReader(overrideBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &runResp))
	assert.True(t, runResp.Data.Success)
	assert.Contains(t, runResp.Data.ReservationStatus, "成功预约")

	// ==========================================
	// 8. Multi-tenant Isolation Test
	// User 102 attempts to access/run/delete User 101's card
	// ==========================================
	authMock.SetUserID(102)

	// User 102 lists configs -> empty
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/igo/pipeline/configs", nil)
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	assert.Empty(t, listResp.Data)

	// User 102 queries User 101's card -> 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/igo/pipeline/configs/my_exam_seat", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// User 102 runs User 101's card -> 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs/my_exam_seat/run", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	u2AccBody := `{"name":"用户B","cookie":"Authorization=traceint-user102-cookie"}`
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/accounts", strings.NewReader(u2AccBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var u2Acc struct {
		Data do.AccountDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u2Acc))

	u2Body := fmt.Sprintf(`{
		"id": "user102_seat",
		"name": "用户B的座位",
		"account_id": "%s",
		"library_id": 20,
		"library_name": "总馆二楼",
		"floor": "2",
		"seat_key": "S-201",
		"seat_name": "201号",
		"auto_checkin": false
	}`, strconv.FormatUint(u2Acc.Data.ID, 10))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/configs", strings.NewReader(u2Body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Switch back to User 101
	authMock.SetUserID(101)

	// ==========================================
	// 9. Delete config
	// ==========================================
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/igo/pipeline/configs/my_exam_seat", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify deleted
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/igo/pipeline/configs/my_exam_seat", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// ==========================================
	// 10. Verify helper endpoints: verify-session & library-layout
	// ==========================================
	// 10.1 Empty cookie -> 400 Bad Request
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/verify-session", strings.NewReader(`{"cookie":""}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 10.2 Missing body -> 400 Bad Request
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/verify-session", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 10.3 Valid cookie -> 200 OK with libraries
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/verify-session", strings.NewReader(`{"cookie":"Authorization=traceint-user101-cookie"}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var verifyRes response.Response[map[string]any]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &verifyRes))
	assert.Equal(t, true, verifyRes.Data["valid"])
	assert.NotEmpty(t, verifyRes.Data["libraries"])

	// 10.4 library-layout without cookie and without stored session -> 400
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/library-layout", strings.NewReader(`{"cookie":"","library_id":20}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 10.5 library-layout with cookie -> 200 OK with seats
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/igo/pipeline/library-layout", strings.NewReader(`{"cookie":"Authorization=traceint-user101-cookie","library_id":20}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
