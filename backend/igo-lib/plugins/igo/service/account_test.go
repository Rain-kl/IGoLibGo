// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountServiceCRUDAndIsolation(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "devices") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试用户","user_sch":"测试大学","user_student_name":"张三","user_student_no":"2023001"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()

	_, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "  "})
	require.Error(t, err)
	var coded *consts.CodedError
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, consts.CodeValidationError, coded.Code)

	created, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "甲"})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "甲", created.Name)
	assert.False(t, created.HasCookie)
	assert.False(t, created.HasCheckinToken)

	_, err = svc.GetAccount(ctx, 2, created.ID)
	require.Error(t, err)
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, consts.CodeNotFound, coded.Code)

	got, err := svc.GetAccount(ctx, 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "甲", got.Name)

	updated, err := svc.UpdateAccount(ctx, 1, created.ID, do.UpdateAccountRequest{Name: "甲改"})
	require.NoError(t, err)
	assert.Equal(t, "甲改", updated.Name)

	list, err := svc.ListAccounts(ctx, 1)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, dao.CreatePipelineConfig(ctx, &entity.PipelineConfig{
		ID:        "card1",
		UserID:    1,
		Name:      "卡",
		Cookie:    "c",
		LibraryID: 1,
		SeatKey:   "s",
		AccountID: created.ID,
	}))
	err = svc.DeleteAccount(ctx, 1, created.ID)
	require.Error(t, err)
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, http.StatusConflict, coded.Status)
	assert.Equal(t, consts.CodeConflict, coded.Code)
	assert.Contains(t, coded.Msg, "card1")

	require.NoError(t, dao.DeletePipelineConfig(ctx, "card1", 1))
	require.NoError(t, svc.DeleteAccount(ctx, 1, created.ID))
}

func TestAuthorizeAccountCheckinFillsProfile(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "devices") {
			_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试用户","user_sch":"测试大学","user_student_name":"张三","user_student_no":"2023001"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	created, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "乙"})
	require.NoError(t, err)

	out, err := svc.AuthorizeAccountCheckin(ctx, 1, created.ID, do.AccountCheckinAuthRequest{Code: "mock_token_abcdefghijklmnopqrstuvwxyz12"})
	require.NoError(t, err)
	require.NotNil(t, out.Device)
	assert.Equal(t, "张三", out.Account.StudentName)
	assert.True(t, out.Account.HasCheckinToken)
	assert.Contains(t, out.Device.BeaconUUIDs, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825")
}
