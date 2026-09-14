// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package traceint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractCode(t *testing.T) {
	code, ok := ExtractCode("https://web.traceint.com/web/index.html?code=abcdefghijklmnopqrstuvwxyz123456&state=1")
	require.True(t, ok)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz123456", code)

	code, ok = ExtractCode("abcdefghijklmnopqrstuvwxyz123456")
	require.True(t, ok)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz123456", code)

	_, ok = ExtractCode("short")
	assert.False(t, ok)
}

func TestMapLibrariesSkipsFloorZero(t *testing.T) {
	raw := []byte(`{"data":{"userAuth":{"reserve":{"libs":[
		{"lib_id":1,"lib_name":"一楼","lib_floor":"1","is_open":true,"lib_rt":{"seats_total":10,"seats_used":2,"seats_booking":1}},
		{"lib_id":2,"lib_name":"隐藏","lib_floor":"0","is_open":true}
	]}}}}`)
	libs, err := mapLibraries(raw)
	require.NoError(t, err)
	require.Len(t, libs, 1)
	assert.Equal(t, 1, libs[0].LibraryID)
	assert.Equal(t, 10, libs[0].TotalSeats)
}

func TestMapReservationEmpty(t *testing.T) {
	raw := []byte(`{"data":{"userAuth":{"reserve":{"reserve":null,"getSToken":""}}}}`)
	info, err := mapReservation(raw)
	require.NoError(t, err)
	assert.False(t, info.HasReservation)
}

func TestGraphQLError(t *testing.T) {
	_, err := mapLibraries([]byte(`{"errors":[{"message":"未登录","code":40001}]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "未登录")
}

func TestTomorrowSeatKey(t *testing.T) {
	assert.Equal(t, "A1.", TomorrowSeatKey("A1"))
	assert.Equal(t, "A1.", TomorrowSeatKey("A1."))
}

func TestBuildAuthorizationURL(t *testing.T) {
	u := BuildAuthorizationURL(
		"http://example.com/auth?r=ReplaceMeByReturnUrl&code=ReplaceMeByCode",
		"abc",
		"https://web.traceint.com/web/index.html",
	)
	assert.Contains(t, u, "code=abc")
	assert.Contains(t, u, "r=https")
}
