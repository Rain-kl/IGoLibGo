// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// Settings stores migratable per-user settings as JSON.
type Settings struct {
	ID        uint64    `gorm:"primaryKey" json:"id,string"`
	UserID    uint64    `gorm:"uniqueIndex:uq_igo_settings_user;not null" json:"user_id,string"`
	Payload   string    `gorm:"type:text;not null" json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (Settings) TableName() string { return consts.TableSettings }
