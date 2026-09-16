// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// PipelineConfig is an All-in-One automation configuration card.
type PipelineConfig struct {
	ID               string     `gorm:"primaryKey;size:64" json:"id"`
	UserID           uint64     `gorm:"index:idx_igo_pipeline_configs_user;not null" json:"user_id,string"`
	Name             string     `gorm:"size:128;not null" json:"name"`
	Cookie           string     `gorm:"type:text;not null" json:"-"`
	CookieExpiresAt  *time.Time `json:"cookie_expires_at,omitempty"`
	LibraryID        int        `gorm:"not null" json:"library_id"`
	LibraryName      string     `gorm:"size:255;not null" json:"library_name"`
	Floor            string     `gorm:"size:64;not null;default:''" json:"floor"`
	SeatKey          string     `gorm:"size:64;not null" json:"seat_key"`
	SeatName         string     `gorm:"size:128;not null;default:''" json:"seat_name"`
	AutoCheckin      bool       `gorm:"not null;default:false" json:"auto_checkin"`
	CheckinToken     string     `gorm:"type:text" json:"-"`
	CheckinExpiresAt *time.Time `json:"checkin_expires_at,omitempty"`
	BeaconUUID       string     `gorm:"size:64;not null;default:''" json:"beacon_uuid"`
	Major            int        `gorm:"not null;default:0" json:"major"`
	Minor            int        `gorm:"not null;default:0" json:"minor"`
	Latitude         string     `gorm:"size:32;not null;default:''" json:"latitude"`
	Longitude        string     `gorm:"size:32;not null;default:''" json:"longitude"`
	AccountID        uint64     `gorm:"not null;default:0" json:"account_id,string"`
	CheckinAccountID uint64     `gorm:"not null;default:0" json:"checkin_account_id,string"`
	CheckinInfoID    uint64     `gorm:"not null;default:0" json:"checkin_info_id,string"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (PipelineConfig) TableName() string { return consts.TablePipelineConfigs }
