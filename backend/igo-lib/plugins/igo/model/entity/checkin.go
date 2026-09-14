// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// CheckInSession is the independent remote-check-in WeChat session.
type CheckInSession struct {
	ID             uint64     `gorm:"primaryKey" json:"id,string"`
	UserID         uint64     `gorm:"uniqueIndex:uq_igo_checkin_sessions_user;not null" json:"user_id,string"`
	Token          string     `gorm:"type:text;not null" json:"-"`
	SavedAt        time.Time  `gorm:"not null" json:"saved_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CanAutoRestore bool       `gorm:"not null;default:true" json:"can_auto_restore"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (CheckInSession) TableName() string { return consts.TableCheckInSessions }
