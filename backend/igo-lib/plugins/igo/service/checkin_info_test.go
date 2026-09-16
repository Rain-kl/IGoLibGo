// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"errors"
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

func checkinInfoHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case strings.Contains(r.URL.Path, "devices"):
		_, _ = io.WriteString(w, `{"code":0,"msg":"ok","data":{"user":{"user_nick":"测试用户"},"devices":["FDA50693-A4E2-4FB1-AFCF-C6EB07647825"]}}`)
	case strings.Contains(r.URL.Path, "getTime"):
		_, _ = io.WriteString(w, "1710000000")
	case strings.Contains(r.URL.Path, "sign.html"):
		_, _ = io.WriteString(w, `{"code":0,"msg":"验证成功","data":{"status":2,"lib_id":101,"lib_name":"主馆"}}`)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func TestCheckInInfoServiceCRUDAndSign(t *testing.T) {
	svc := setupService(t, checkinInfoHTTP)
	ctx := context.Background()

	created, err := svc.CreateCheckInInfo(ctx, 1, do.CreateCheckInInfoRequest{Name: "主馆"})
	require.NoError(t, err)
	assert.Equal(t, "主馆", created.Name)
	assert.Empty(t, created.BeaconUUID)

	_, err = svc.GetCheckInInfo(ctx, 2, created.ID)
	require.Error(t, err)
	var coded *consts.CodedError
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, consts.CodeNotFound, coded.Code)

	updated, err := svc.UpdateCheckInInfo(ctx, 1, created.ID, do.UpdateCheckInInfoRequest{
		Name:       "主馆西",
		BeaconUUID: "fda50693-a4e2-4fb1-afcf-c6eb07647825",
		Major:      10001,
		Minor:      1980,
		Latitude:   "31.2304",
		Longitude:  "121.4737",
	})
	require.NoError(t, err)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", updated.BeaconUUID)

	list, err := svc.ListCheckInInfos(ctx, 1)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, dao.CreatePipelineConfig(ctx, &entity.PipelineConfig{
		ID:            "card1",
		UserID:        1,
		Name:          "n",
		Cookie:        "c",
		LibraryID:     1,
		SeatKey:       "s",
		CheckinInfoID: created.ID,
	}))
	err = svc.DeleteCheckInInfo(ctx, 1, created.ID)
	require.Error(t, err)
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, http.StatusConflict, coded.Status)
	require.NoError(t, dao.DeletePipelineConfig(ctx, "card1", 1))

	acc, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{
		Name:         "乙",
		CheckinToken: "mock_token_abcdefghijklmnopqrstuvwxyz12",
	})
	require.NoError(t, err)

	signed, err := svc.SignCheckInInfo(ctx, 1, created.ID, do.SignCheckInInfoRequest{
		AccountID:         acc.ID,
		ExpectedLibraryID: 101,
	})
	require.NoError(t, err)
	assert.Equal(t, "验证成功", signed.Message)

	emptyTok, err := svc.CreateAccount(ctx, 1, do.CreateAccountRequest{Name: "空凭证"})
	require.NoError(t, err)
	_, err = svc.SignCheckInInfo(ctx, 1, created.ID, do.SignCheckInInfoRequest{
		AccountID:         emptyTok.ID,
		ExpectedLibraryID: 101,
	})
	require.Error(t, err)
	var na *consts.NeedAuthError
	require.True(t, errors.As(err, &na))
	assert.Equal(t, "CHECKIN", na.Kind)
	assert.Equal(t, emptyTok.ID, na.AccountID)

	bare, err := svc.CreateCheckInInfo(ctx, 1, do.CreateCheckInInfoRequest{Name: "空Beacon"})
	require.NoError(t, err)
	_, err = svc.SignCheckInInfo(ctx, 1, bare.ID, do.SignCheckInInfoRequest{
		AccountID:         acc.ID,
		ExpectedLibraryID: 101,
	})
	require.Error(t, err)
	require.ErrorAs(t, err, &coded)
	assert.Equal(t, consts.CodeValidationError, coded.Code)

	_, err = svc.SignCheckInInfo(ctx, 2, created.ID, do.SignCheckInInfoRequest{
		AccountID:         acc.ID,
		ExpectedLibraryID: 101,
	})
	require.Error(t, err)

	require.NoError(t, svc.DeleteCheckInInfo(ctx, 1, created.ID))
}
