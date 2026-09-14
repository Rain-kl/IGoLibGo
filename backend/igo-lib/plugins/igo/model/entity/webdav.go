// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// WebDAV is the remote sync endpoint for one user. Password encryption is a later stage.
type WebDAV struct {
	ID              uint64    `gorm:"primaryKey" json:"id,string"`
	UserID          uint64    `gorm:"uniqueIndex:uq_igo_webdav_user;not null" json:"user_id,string"`
	Endpoint        string    `gorm:"size:1024;not null;default:''" json:"endpoint"`
	RemoteDirectory string    `gorm:"size:512;not null;default:''" json:"remote_directory"`
	Username        string    `gorm:"size:255;not null;default:''" json:"username"`
	Password        string    `gorm:"type:text;not null;default:''" json:"-"`
	TLSVerifyMode   string    `gorm:"size:32;not null;default:'default'" json:"tls_verify_mode"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (WebDAV) TableName() string { return consts.TableWebDAV }
