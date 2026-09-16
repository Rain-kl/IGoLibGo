// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// Account holds TraceInt occupy credentials and WeChat check-in credentials
// for one student identity owned by a Wavelet user.
//
// Note: 账户与签到信息拆表，pipeline 只存 ID — 见 .agents/notes/implemented/architecture/2026-09-16-checkin-info-and-accounts.md
type Account struct {
	ID               uint64     `gorm:"primaryKey" json:"id,string"`
	UserID           uint64     `gorm:"index:idx_igo_accounts_user;not null" json:"user_id,string"`
	Name             string     `gorm:"size:128;not null" json:"name"`
	Cookie           string     `gorm:"type:text;not null;default:''" json:"-"`
	CookieExpiresAt  *time.Time `json:"cookie_expires_at,omitempty"`
	CookieSource     string     `gorm:"size:32;not null;default:''" json:"cookie_source"`
	CheckinToken     string     `gorm:"type:text;not null;default:''" json:"-"`
	CheckinExpiresAt *time.Time `json:"checkin_expires_at,omitempty"`
	Nickname         string     `gorm:"size:128;not null;default:''" json:"nickname"`
	School           string     `gorm:"size:128;not null;default:''" json:"school"`
	StudentName      string     `gorm:"size:128;not null;default:''" json:"student_name"`
	StudentNumber    string     `gorm:"size:64;not null;default:''" json:"student_number"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (Account) TableName() string { return consts.TableAccounts }
