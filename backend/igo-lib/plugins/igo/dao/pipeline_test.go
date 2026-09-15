// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao_test

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineConfigCRUDAndIsolation(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// 1. Create config for user 1
	cfg1 := &entity.PipelineConfig{
		ID:          "myseat01",
		UserID:      1,
		Name:        "考研专座",
		Cookie:      "cookie-u1",
		LibraryID:   101,
		LibraryName: "总馆三楼",
		Floor:       "3",
		SeatKey:     "SK-301",
		SeatName:    "301号",
		AutoCheckin: true,
		BeaconUUID:  "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major:       10001,
		Minor:       1984,
		Latitude:    "39.9042",
		Longitude:   "116.4074",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, dao.CreatePipelineConfig(ctx, cfg1))

	// Duplicate primary key should fail
	dupCfg := &entity.PipelineConfig{
		ID:          "myseat01",
		UserID:      2,
		Name:        "重复ID",
		Cookie:      "cookie-u2",
		LibraryID:   102,
		LibraryName: "分馆",
		SeatKey:     "SK-202",
	}
	require.Error(t, dao.CreatePipelineConfig(ctx, dupCfg))

	// Create config for user 2
	cfg2 := &entity.PipelineConfig{
		ID:          "room202",
		UserID:      2,
		Name:        "自习室二楼",
		Cookie:      "cookie-u2",
		LibraryID:   102,
		LibraryName: "分馆二楼",
		Floor:       "2",
		SeatKey:     "SK-202",
		SeatName:    "202号",
		AutoCheckin: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, dao.CreatePipelineConfig(ctx, cfg2))

	// 2. Query by ID
	got1, err := dao.GetPipelineConfig(ctx, "myseat01")
	require.NoError(t, err)
	require.NotNil(t, got1)
	assert.Equal(t, uint64(1), got1.UserID)
	assert.Equal(t, "考研专座", got1.Name)
	assert.Equal(t, "cookie-u1", got1.Cookie)
	assert.True(t, got1.AutoCheckin)

	// 3. User isolation check
	gotU1, err := dao.GetPipelineConfigByUser(ctx, "myseat01", 1)
	require.NoError(t, err)
	require.NotNil(t, gotU1)

	gotU2Unauthorized, err := dao.GetPipelineConfigByUser(ctx, "myseat01", 2)
	require.NoError(t, err)
	assert.Nil(t, gotU2Unauthorized)

	// 4. List by user
	list1, err := dao.ListPipelineConfigsByUser(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, list1, 1)
	assert.Equal(t, "myseat01", list1[0].ID)

	list2, err := dao.ListPipelineConfigsByUser(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Equal(t, "room202", list2[0].ID)

	// 5. Update config
	got1.Name = "考研专座(改)"
	got1.SeatKey = "SK-302"
	got1.SeatName = "302号"
	require.NoError(t, dao.UpdatePipelineConfig(ctx, got1))

	updated1, err := dao.GetPipelineConfig(ctx, "myseat01")
	require.NoError(t, err)
	require.NotNil(t, updated1)
	assert.Equal(t, "考研专座(改)", updated1.Name)
	assert.Equal(t, "SK-302", updated1.SeatKey)

	// Update credentials
	expTime := now.Add(24 * time.Hour)
	require.NoError(t, dao.UpdatePipelineCookie(ctx, "myseat01", "new-cookie-u1", &expTime))
	require.NoError(t, dao.UpdatePipelineCheckinToken(ctx, "myseat01", "sess-token-u1", &expTime))

	updatedCreds, err := dao.GetPipelineConfig(ctx, "myseat01")
	require.NoError(t, err)
	assert.Equal(t, "new-cookie-u1", updatedCreds.Cookie)
	assert.Equal(t, "sess-token-u1", updatedCreds.CheckinToken)

	// 6. Delete config (cross-user delete should fail with ErrPipelineConfigNotFound)
	require.ErrorIs(t, dao.DeletePipelineConfig(ctx, "myseat01", 2), dao.ErrPipelineConfigNotFound)

	// Owner delete succeeds
	require.NoError(t, dao.DeletePipelineConfig(ctx, "myseat01", 1))
	deleted, err := dao.GetPipelineConfig(ctx, "myseat01")
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
