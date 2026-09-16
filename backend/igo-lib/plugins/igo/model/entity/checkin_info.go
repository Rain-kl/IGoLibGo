// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// CheckInInfo stores reusable Beacon simulation parameters with no credentials.
type CheckInInfo struct {
	ID         uint64    `gorm:"primaryKey" json:"id,string"`
	UserID     uint64    `gorm:"index:idx_igo_checkin_infos_user;not null" json:"user_id,string"`
	Name       string    `gorm:"size:128;not null" json:"name"`
	BeaconUUID string    `gorm:"size:64;not null;default:''" json:"beacon_uuid"`
	Major      int       `gorm:"not null;default:0" json:"major"`
	Minor      int       `gorm:"not null;default:0" json:"minor"`
	Latitude   string    `gorm:"size:32;not null;default:''" json:"latitude"`
	Longitude  string    `gorm:"size:32;not null;default:''" json:"longitude"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (CheckInInfo) TableName() string { return consts.TableCheckInInfos }
