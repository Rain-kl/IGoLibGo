// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// ProtocolOverride stores per-user TraceInt template overrides as JSON.
type ProtocolOverride struct {
	ID        uint64    `gorm:"primaryKey" json:"id,string"`
	UserID    uint64    `gorm:"uniqueIndex:uq_igo_protocol_overrides_user;not null" json:"user_id,string"`
	Overrides string    `gorm:"type:text;not null" json:"overrides"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (ProtocolOverride) TableName() string { return consts.TableProtocolOverrides }
