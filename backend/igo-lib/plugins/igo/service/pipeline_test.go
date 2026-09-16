// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func expiredAuthCookie() string {
	payload, _ := json.Marshal(map[string]any{"exp": time.Now().Add(-time.Hour).Unix()})
	token := "eyJhbGciOiJub25lIn0." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
	return "Authorization=" + token
}

func TestPipelineConfigServiceCRUDAndValidation(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "graphql") {
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[
				{"lib_id":101,"lib_name":"总馆三楼","lib_floor":"3","is_open":true,"lib_rt":{"seats_total":50,"seats_used":10,"seats_booking":5}}
			]}}}}`)
			return
		}
		// Checkin devices endpoint
		if strings.Contains(r.URL.Path, "devices") || strings.Contains(r.URL.RawQuery, "devices") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试用户"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	acc, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "甲", Cookie: "Authorization=test-cookie", CheckinToken: "sess-token-123"})
	require.NoError(t, err)
	info, err := svc.CreateCheckInInfo(ctx, 1, do.CreateCheckInInfoRequest{
		Name: "一号座签到", BeaconUUID: "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major: 10001, Minor: 1984, Latitude: "39.9042", Longitude: "116.4074",
	})
	require.NoError(t, err)

	// 1. Invalid ID test
	_, err = svc.CreatePipelineConfig(ctx, 1, do.CreatePipelineConfigRequest{
		ID:        "invalid id with spaces!",
		Name:      "测试",
		AccountID: acc.ID,
		LibraryID: 101,
		SeatKey:   "SK-1",
	})
	require.Error(t, err)

	// 2. Create valid config
	created, err := svc.CreatePipelineConfig(ctx, 1, do.CreatePipelineConfigRequest{
		ID:            "test_seat01",
		Name:          "一号座",
		AccountID:     acc.ID,
		CheckinInfoID: info.ID,
		LibraryID:     101,
		LibraryName:   "总馆三楼",
		Floor:         "3",
		SeatKey:       "SK-1",
		SeatName:      "01号",
		AutoCheckin:   true,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "test_seat01", created.ID)
	assert.Equal(t, "一号座", created.Name)
	assert.True(t, created.HasCookie)
	assert.True(t, created.AutoCheckin)
	assert.True(t, created.HasCheckinToken)

	// Duplicate ID should error
	_, err = svc.CreatePipelineConfig(ctx, 1, do.CreatePipelineConfigRequest{
		ID:        "test_seat01",
		Name:      "重复",
		AccountID: acc.ID,
		LibraryID: 101,
		SeatKey:   "SK-2",
	})
	require.Error(t, err)

	// 3. List configs
	list, err := svc.ListPipelineConfigs(ctx, 1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "test_seat01", list[0].ID)

	// 4. Get config
	got, err := svc.GetPipelineConfig(ctx, 1, "test_seat01")
	require.NoError(t, err)
	assert.Equal(t, "test_seat01", got.ID)

	// 5. Update config
	updated, err := svc.UpdatePipelineConfig(ctx, 1, "test_seat01", do.UpdatePipelineConfigRequest{
		Name:    "一号座(修改)",
		SeatKey: "SK-2",
	})
	require.NoError(t, err)
	assert.Equal(t, "一号座(修改)", updated.Name)
	assert.Equal(t, "SK-2", updated.SeatKey)

	// 6. Delete config
	require.NoError(t, svc.DeletePipelineConfig(ctx, 1, "test_seat01"))
	_, err = svc.GetPipelineConfig(ctx, 1, "test_seat01")
	require.Error(t, err)
}

func TestRunPipelineScenarios(t *testing.T) {
	var seatOccupied bool
	var cookieInvalid bool
	var checkinInvalid bool
	var reserveFail bool

	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)

		if strings.Contains(r.URL.Path, "graphql") {
			if cookieInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"errors":[{"message":"登录已过期"}]}`)
				return
			}
			// Library list or layout or reserve
			if strings.Contains(bodyStr, "lib_layout") {
				statusVal := false
				if seatOccupied {
					statusVal = true
				}
				_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[{
					"lib_id":101,"lib_name":"总馆三楼","lib_floor":"3","is_open":true,
					"lib_layout":{
						"seats_total":10,"seats_booking":0,"seats_used":0,
						"seats":[{"key":"SK-100","name":"100号","type":1,"status":`+fmtBool(statusVal)+`,"x":1,"y":1}]
					}
				}]}}}}`)
				return
			}
			if strings.Contains(bodyStr, "reserueSeat") {
				if reserveFail {
					_, _ = io.WriteString(w, `{"errors":[{"message":"座位已被其他用户锁定"}]}`)
					return
				}
				_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"reserueSeat":true}}}}`)
				return
			}
			// General query / ListLibraries
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[
				{"lib_id":101,"lib_name":"总馆三楼","lib_floor":"3","is_open":true,"lib_rt":{"seats_total":50,"seats_used":10,"seats_booking":5}}
			]}}}}`)
			return
		}

		// Checkin endpoints
		if strings.Contains(r.URL.Path, "time") {
			_, _ = io.WriteString(w, "1726390000")
			return
		}
		if strings.Contains(r.URL.Path, "devices") || strings.Contains(r.URL.RawQuery, "devices") {
			if checkinInvalid {
				_, _ = io.WriteString(w, `{"code":1,"msg":"微信授权已过期"}`)
				return
			}
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试用户"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
			return
		}
		// Sign checkin
		if strings.Contains(r.URL.Path, "sign") || strings.Contains(r.URL.RawQuery, "sign") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"签到成功","data":{"status":1,"lib_id":101,"seat_name":"100号"}}`)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	acc, err := svc.CreateAccount(ctx, 9, do.CreateAccountRequest{
		Name: "执行号", Cookie: "Authorization=valid-cookie", CheckinToken: "valid-checkin-token",
	})
	require.NoError(t, err)
	info, err := svc.CreateCheckInInfo(ctx, 9, do.CreateCheckInInfoRequest{
		Name: "执行签到", BeaconUUID: "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major: 10001, Minor: 1984, Latitude: "39.9042", Longitude: "116.4074",
	})
	require.NoError(t, err)

	_, err = svc.CreatePipelineConfig(ctx, 9, do.CreatePipelineConfigRequest{
		ID:            "seat_exec_test",
		Name:          "执行测试",
		AccountID:     acc.ID,
		CheckinInfoID: info.ID,
		LibraryID:     101,
		LibraryName:   "总馆三楼",
		Floor:         "3",
		SeatKey:       "SK-100",
		SeatName:      "100号",
		AutoCheckin:   true,
	})
	require.NoError(t, err)

	// Scenario 1: Cookie expired -> return NeedAuth == "LOGIN"
	cookieInvalid = true
	res, err := svc.RunPipeline(ctx, 9, "seat_exec_test", nil)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "LOGIN", res.NeedAuth)
	assert.Equal(t, consts.WeChatLoginAuthURL, res.AuthURL)
	assert.Contains(t, res.Message, "TraceInt 账户授权已过期")

	// Scenario 2: Cookie valid, but checkin token invalid -> return NeedAuth == "CHECKIN"
	cookieInvalid = false
	checkinInvalid = true
	res, err = svc.RunPipeline(ctx, 9, "seat_exec_test", nil)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "CHECKIN", res.NeedAuth)
	assert.Equal(t, consts.WeChatCheckinAuthURL, res.AuthURL)
	assert.Contains(t, res.Message, "签到微信授权已失效")

	// Scenario 3: Credentials valid, but seat is occupied -> exit with failure
	checkinInvalid = false
	seatOccupied = true
	res, err = svc.RunPipeline(ctx, 9, "seat_exec_test", nil)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Empty(t, res.NeedAuth)
	assert.Contains(t, res.Message, "已被占用")

	// Scenario 4: Credentials valid, seat is free -> full success (seat reserved + checkin signed)
	seatOccupied = false
	res, err = svc.RunPipeline(ctx, 9, "seat_exec_test", nil)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "成功预约 [总馆三楼 100号]", res.ReservationStatus)
	assert.Contains(t, res.CheckinStatus, "打卡成功")
	assert.Contains(t, res.Message, "一条龙全流程执行成功")

	// Helper verification APIs
	libs, _, _, err := svc.HelperVerifySession(ctx, 9, "Authorization=test", 0)
	require.NoError(t, err)
	assert.Len(t, libs, 1)

	layout, err := svc.HelperGetLibraryLayout(ctx, 9, "Authorization=test", 0, 101)
	require.NoError(t, err)
	assert.Len(t, layout.Seats, 1)

	devs, _, _, err := svc.HelperVerifyCheckin(ctx, 9, "valid-checkin-token")
	require.NoError(t, err)
	assert.Equal(t, "测试用户", devs.Nickname)
}

func fmtBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestRunPipeline_SeparateOccupyAndCheckinAccounts(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if strings.Contains(r.URL.Path, "graphql") {
			if strings.Contains(bodyStr, "lib_layout") {
				_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[{
					"lib_id":101,"lib_name":"总馆三楼","lib_floor":"3","is_open":true,
					"lib_layout":{"seats":[{"key":"SK-100","name":"100号","type":1,"status":false,"x":1,"y":1}]}
				}]}}}}`)
				return
			}
			if strings.Contains(bodyStr, "reserueSeat") {
				_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"reserueSeat":true}}}}`)
				return
			}
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[{"lib_id":101,"lib_name":"总馆三楼","lib_floor":"3","is_open":true}]}}}}`)
			return
		}
		if strings.Contains(r.URL.Path, "time") {
			_, _ = io.WriteString(w, "1726390000")
			return
		}
		if strings.Contains(r.URL.Path, "devices") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"乙"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
			return
		}
		if strings.Contains(r.URL.Path, "sign") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"签到成功","data":{"status":1}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	occupy, err := svc.CreateAccount(ctx, 3, do.CreateAccountRequest{Name: "占座号", Cookie: "Authorization=occupy-cookie"})
	require.NoError(t, err)
	checkin, err := svc.CreateAccount(ctx, 3, do.CreateAccountRequest{Name: "打卡号", CheckinToken: "checkin-token-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"})
	require.NoError(t, err)
	info, err := svc.CreateCheckInInfo(ctx, 3, do.CreateCheckInInfoRequest{
		Name: "共用Beacon", BeaconUUID: "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major: 1, Minor: 2, Latitude: "31.2", Longitude: "121.4",
	})
	require.NoError(t, err)
	_, err = svc.CreatePipelineConfig(ctx, 3, do.CreatePipelineConfigRequest{
		ID: "dual", Name: "双号", AccountID: occupy.ID, CheckinAccountID: checkin.ID,
		CheckinInfoID: info.ID, LibraryID: 101, LibraryName: "总馆三楼", Floor: "3",
		SeatKey: "SK-100", SeatName: "100号", AutoCheckin: true,
	})
	require.NoError(t, err)
	res, err := svc.RunPipeline(ctx, 3, "dual", nil)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Contains(t, res.CheckinStatus, "打卡成功")
}

func TestCreatePipelineConfig_AutoCheckinRequiresInfo(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	acc, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "甲", Cookie: "Authorization=c"})
	require.NoError(t, err)
	_, err = svc.CreatePipelineConfig(ctx, 1, do.CreatePipelineConfigRequest{
		ID: "noinfo", Name: "x", AccountID: acc.ID, LibraryID: 1, SeatKey: "s", AutoCheckin: true,
	})
	require.Error(t, err)
	var coded *consts.CodedError
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, consts.CodeValidationError, coded.Code)
}

func layoutAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "graphql") {
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[{
				"lib_id":211,"lib_name":"总馆","lib_floor":"1","is_open":true,
				"lib_layout":{"seats_total":1,"seats_booking":0,"seats_used":0,"max_x":1,"max_y":1,
					"seats":[{"x":1,"y":1,"key":"A1","type":1,"name":"1号","status":false}]
				}}]}}}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func TestHelperGetLibraryLayout_PrefersStoredCookie(t *testing.T) {
	svc := setupService(t, layoutAPIHandler())
	ctx := context.Background()

	_, err := svc.HelperGetLibraryLayout(ctx, 7, "", 0, 211)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "请传入登录凭据")

	_, err = svc.AuthenticateFromCookie(ctx, 7, do.AuthFromCookieRequest{
		Cookie:   "Authorization=stored-cookie",
		Remember: true,
	})
	require.NoError(t, err)

	layout, err := svc.HelperGetLibraryLayout(ctx, 7, "", 0, 211)
	require.NoError(t, err)
	require.NotNil(t, layout)
	assert.Equal(t, 211, layout.LibraryID)
	require.Len(t, layout.Seats, 1)
	assert.Equal(t, "A1", layout.Seats[0].SeatKey)
}

func TestHelperGetLibraryLayout_ExpiredStoredCookieRequiresInput(t *testing.T) {
	svc := setupService(t, layoutAPIHandler())
	ctx := context.Background()

	expired := expiredAuthCookie()
	_, err := svc.AuthenticateFromCookie(ctx, 8, do.AuthFromCookieRequest{
		Cookie:   expired,
		Remember: true,
	})
	require.NoError(t, err)

	_, err = svc.HelperGetLibraryLayout(ctx, 8, "", 0, 211)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "已过期")

	layout, err := svc.HelperGetLibraryLayout(ctx, 8, "Authorization=fresh-cookie", 0, 211)
	require.NoError(t, err)
	require.NotNil(t, layout)
	assert.Equal(t, "A1", layout.Seats[0].SeatKey)
}
