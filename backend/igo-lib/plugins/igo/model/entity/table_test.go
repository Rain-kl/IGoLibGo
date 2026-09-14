// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity_test

import (
	"testing"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/entity"

	"github.com/stretchr/testify/assert"
)

func TestTableNamesMatchOwnedTables(t *testing.T) {
	got := map[string]struct{}{
		entity.Session{}.TableName():                 {},
		entity.Venue{}.TableName():                   {},
		entity.Favorite{}.TableName():                {},
		entity.SeatLabel{}.TableName():               {},
		entity.ProtocolOverride{}.TableName():        {},
		entity.Settings{}.TableName():                {},
		entity.TaskRun{}.TableName():                 {},
		entity.TaskLaunchHistory{}.TableName():       {},
		entity.GlobalLeakTarget{}.TableName():        {},
		entity.GlobalLeakBlacklistSeat{}.TableName(): {},
		entity.CheckInSession{}.TableName():          {},
		entity.DashboardMetrics{}.TableName():        {},
		entity.WebDAV{}.TableName():                  {},
	}
	assert.Len(t, got, len(consts.OwnedTables))
	for _, name := range consts.OwnedTables {
		_, ok := got[name]
		assert.True(t, ok, "missing entity TableName for %s", name)
		assert.Contains(t, name, "w_igo_")
	}
}
