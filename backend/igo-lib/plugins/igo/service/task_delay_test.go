// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"encoding/json"
	"testing"
	"time"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTickDelayOccupyFixedTenSeconds(t *testing.T) {
	raw, _ := json.Marshal(do.OccupyStartRequest{CheckIntervalMode: "fixed_ten_seconds"})
	assert.Equal(t, 10*time.Second, tickDelay(consts.TaskKindOccupy, string(raw)))
}

func TestTickDelayLeakUsesScanInterval(t *testing.T) {
	raw, _ := json.Marshal(do.GlobalLeakStartRequest{ScanIntervalSeconds: 7})
	assert.Equal(t, 7*time.Second, tickDelay(consts.TaskKindGlobalLeak, string(raw)))
}

func TestTickDelayGrabAggressive(t *testing.T) {
	raw, _ := json.Marshal(do.GrabStartRequest{PollingMode: "aggressive"})
	assert.Equal(t, time.Second, tickDelay(consts.TaskKindGrab, string(raw)))
}

func TestScheduledWaitAcceptsHHMM(t *testing.T) {
	wait, _ := scheduledWait("23:59")
	// 23:59 may or may not be in the future depending on local clock; parse must succeed.
	fire, ok := parseClock("23:59")
	require.True(t, ok)
	assert.Equal(t, 23, fire.Hour())
	assert.Equal(t, 59, fire.Minute())
	_ = wait
}

func TestParseClockPastTimeOfDayRollsToTomorrow(t *testing.T) {
	now := time.Now()
	past := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if now.Hour() == 0 && now.Minute() == 0 {
		t.Skip("midnight")
	}
	fire, ok := parseClock("00:00:00")
	require.True(t, ok)
	assert.True(t, fire.After(now), "got %s vs now %s past=%s", fire, now, past)
}

func TestParseClockRFC3339(t *testing.T) {
	fire, ok := parseClock("2099-01-01T00:00:00Z")
	require.True(t, ok)
	wait, msg := scheduledWait("2099-01-01T00:00:00Z")
	assert.True(t, wait)
	assert.Contains(t, msg, "等待定时启动")
	assert.True(t, fire.After(time.Now()))
}
