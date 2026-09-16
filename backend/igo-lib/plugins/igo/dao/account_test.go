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

func TestAccountCRUDAndIsolation(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()

	a := &entity.Account{UserID: 1, Name: "甲", Cookie: "mock_cookie_a"}
	require.NoError(t, dao.CreateAccount(ctx, a))
	require.NotZero(t, a.ID)

	got, err := dao.GetAccountByUser(ctx, a.ID, 1)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "甲", got.Name)
	assert.Equal(t, "mock_cookie_a", got.Cookie)

	cross, err := dao.GetAccountByUser(ctx, a.ID, 2)
	require.NoError(t, err)
	assert.Nil(t, cross)

	dup, err := dao.FindAccountByCookie(ctx, 1, "mock_cookie_a")
	require.NoError(t, err)
	require.NotNil(t, dup)
	assert.Equal(t, a.ID, dup.ID)

	empty, err := dao.FindAccountByCookie(ctx, 1, "")
	require.NoError(t, err)
	assert.Nil(t, empty)

	blank, err := dao.FindAccountByCookie(ctx, 1, "   ")
	require.NoError(t, err)
	assert.Nil(t, blank)

	got.Name = "甲改"
	require.NoError(t, dao.UpdateAccount(ctx, got))
	updated, err := dao.GetAccountByUser(ctx, a.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "甲改", updated.Name)

	list, err := dao.ListAccountsByUser(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	otherList, err := dao.ListAccountsByUser(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, otherList)

	require.NoError(t, dao.CreatePipelineConfig(ctx, &entity.PipelineConfig{
		ID:               "card1",
		UserID:           1,
		Name:             "卡",
		Cookie:           "c",
		LibraryID:        1,
		SeatKey:          "s",
		AccountID:        a.ID,
		CheckinAccountID: a.ID,
	}))
	refs, err := dao.ListPipelineIDsByAccount(ctx, 1, a.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"card1"}, refs)

	require.NoError(t, dao.DeleteAccount(ctx, a.ID, 1))
	gone, err := dao.GetAccountByUser(ctx, a.ID, 1)
	require.NoError(t, err)
	assert.Nil(t, gone)
}
