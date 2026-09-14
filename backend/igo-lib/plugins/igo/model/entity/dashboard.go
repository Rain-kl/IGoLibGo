// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// DashboardMetrics stores cumulative home-page counters per user.
type DashboardMetrics struct {
	ID                     uint64    `gorm:"primaryKey" json:"id,string"`
	UserID                 uint64    `gorm:"uniqueIndex:uq_w_igo_dashboard_metrics_user;not null" json:"user_id,string"`
	HistoricalSuccessCount int       `gorm:"not null;default:0" json:"historical_success_count"`
	TotalGuardSeconds      int64     `gorm:"not null;default:0" json:"total_guard_seconds"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (DashboardMetrics) TableName() string { return consts.TableDashboardMetrics }
