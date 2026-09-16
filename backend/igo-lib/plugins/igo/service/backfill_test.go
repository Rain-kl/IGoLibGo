// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackfillAccountsAndCheckinInfosIdempotent(t *testing.T) {
	svc := setupService(t, func(http.ResponseWriter, *http.Request) {})
	ctx := context.Background()
	now := time.Now().UTC()

	require.NoError(t, dao.UpsertSession(ctx, &entity.Session{
		UserID:  1,
		Cookie:  "mock_cookie_shared",
		Source:  "test",
		SavedAt: now,
	}))
	require.NoError(t, dao.UpsertCheckInSession(ctx, &entity.CheckInSession{
		UserID:  1,
		Token:   "mock_checkin_token_aaaaaaaaaaaaaaaaaaaa",
		SavedAt: now,
	}))
	require.NoError(t, dao.CreatePipelineConfig(ctx, &entity.PipelineConfig{
		ID:          "card1",
		UserID:      1,
		Name:        "考研专座",
		Cookie:      "mock_cookie_shared",
		LibraryID:   101,
		LibraryName: "总馆",
		SeatKey:     "SK-1",
		BeaconUUID:  "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major:       10001,
		Minor:       1980,
		Latitude:    "31.2",
		Longitude:   "121.4",
		AutoCheckin: true,
	}))
	payload, err := json.Marshal(do.SettingsResponse{
		CheckInProfiles: map[int]do.CheckInVenueProfile{
			101: {LibraryID: 101, LibraryName: "总馆", BeaconUUID: "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", Major: 1, Minor: 2, Latitude: 31.2, Longitude: 121.4},
		},
	})
	require.NoError(t, err)
	require.NoError(t, dao.UpsertSettings(ctx, &entity.Settings{UserID: 1, Payload: string(payload)}))

	require.NoError(t, svc.BackfillAccountsAndCheckinInfos(ctx))
	require.NoError(t, svc.BackfillAccountsAndCheckinInfos(ctx))

	accounts, err := svc.ListAccounts(ctx, 1)
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	assert.Equal(t, "默认", accounts[0].Name)
	assert.True(t, accounts[0].HasCookie)
	assert.True(t, accounts[0].HasCheckinToken)

	infos, err := svc.ListCheckInInfos(ctx, 1)
	require.NoError(t, err)
	require.Len(t, infos, 2)
	names := []string{infos[0].Name, infos[1].Name}
	assert.Contains(t, names, "考研专座 签到")
	assert.Contains(t, names, "总馆")

	card, err := dao.GetPipelineConfigByUser(ctx, "card1", 1)
	require.NoError(t, err)
	require.NotZero(t, card.AccountID)
	require.NotZero(t, card.CheckinInfoID)
	assert.Equal(t, accounts[0].ID, card.AccountID)
}
