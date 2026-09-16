// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao_test

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckInInfoCRUDAndIsolation(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()

	info := &entity.CheckInInfo{
		UserID:     1,
		Name:       "主馆",
		BeaconUUID: "FDA50693-A4E2-4FB1-AFCF-C6EB07647825",
		Major:      10001,
		Minor:      1980,
		Latitude:   "31.2304",
		Longitude:  "121.4737",
	}
	require.NoError(t, dao.CreateCheckInInfo(ctx, info))
	require.NotZero(t, info.ID)

	got, err := dao.GetCheckInInfoByUser(ctx, info.ID, 1)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "主馆", got.Name)

	cross, err := dao.GetCheckInInfoByUser(ctx, info.ID, 2)
	require.NoError(t, err)
	assert.Nil(t, cross)

	got.Name = "主馆西"
	require.NoError(t, dao.UpdateCheckInInfo(ctx, got))
	updated, err := dao.GetCheckInInfoByUser(ctx, info.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "主馆西", updated.Name)

	list, err := dao.ListCheckInInfosByUser(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	require.NoError(t, dao.CreatePipelineConfig(ctx, &entity.PipelineConfig{
		ID:            "card1",
		UserID:        1,
		Name:          "n",
		Cookie:        "c",
		LibraryID:     1,
		SeatKey:       "s",
		CheckinInfoID: info.ID,
	}))
	refs, err := dao.ListPipelineIDsByCheckInInfo(ctx, 1, info.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"card1"}, refs)

	require.NoError(t, dao.DeleteCheckInInfo(ctx, info.ID, 1))
	gone, err := dao.GetCheckInInfoByUser(ctx, info.ID, 1)
	require.NoError(t, err)
	assert.Nil(t, gone)
}
