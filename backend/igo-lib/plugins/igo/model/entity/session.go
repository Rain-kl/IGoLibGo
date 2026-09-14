// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package entity defines GORM table mappings for the igo plugin.
package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// Session is the TraceInt cookie session for one Wavelet user.
type Session struct {
	ID             uint64     `gorm:"primaryKey" json:"id,string"`
	UserID         uint64     `gorm:"uniqueIndex:uq_igo_sessions_user;not null" json:"user_id,string"`
	Cookie         string     `gorm:"type:text;not null" json:"-"`
	Source         string     `gorm:"size:32;not null;default:''" json:"source"`
	SavedAt        time.Time  `gorm:"not null" json:"saved_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CanAutoRestore bool       `gorm:"not null;default:true" json:"can_auto_restore"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (Session) TableName() string { return consts.TableSessions }
