// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package traceint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeUUID(t *testing.T) {
	// Standard with hyphens
	u1, ok := NormalizeUUID("fda50693-a4e2-4fb1-afcf-c6eb07647825")
	require.True(t, ok)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", u1)

	// Uppercase without hyphens
	u2, ok := NormalizeUUID("FDA50693A4E24FB1AFCFC6EB07647825")
	require.True(t, ok)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", u2)

	// With whitespace
	u3, ok := NormalizeUUID("  fda50693-a4e2-4fb1-afcf-c6eb07647825  ")
	require.True(t, ok)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", u3)

	// Invalid
	_, ok = NormalizeUUID("not-a-valid-uuid")
	assert.False(t, ok)
}

func TestMapSignWithTimestamps(t *testing.T) {
	raw := []byte(`{
		"code": 0,
		"msg": "验证成功",
		"data": {
			"status": 2,
			"lib_id": 101,
			"lib_name": "第三电子阅览室",
			"lib_floor": "3楼",
			"seat_key": "12,34",
			"seat_name": "042",
			"date": 1782346716,
			"exp_date": 1782348516
		}
	}`)
	res, err := mapSign(raw)
	require.NoError(t, err)
	assert.Equal(t, "验证成功", res.Message)
	require.NotNil(t, res.Status)
	assert.Equal(t, 2, *res.Status)
	require.NotNil(t, res.LibraryID)
	assert.Equal(t, 101, *res.LibraryID)
	assert.Equal(t, "第三电子阅览室", res.LibraryName)
	assert.Equal(t, "3楼", res.LibraryFloor)
	assert.Equal(t, "042", res.SeatName)
	assert.NotEmpty(t, res.SignedAt)
	assert.NotEmpty(t, res.ExpirationTime)
}
