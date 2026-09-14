// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
)

// GlobalLeakTarget is a venue in scan-priority order for one user.
type GlobalLeakTarget struct {
	ID           uint64 `gorm:"primaryKey" json:"id,string"`
	UserID       uint64 `gorm:"uniqueIndex:uq_igo_leak_targets_user_lib,priority:1;index:idx_igo_leak_targets_user_prio,priority:1;not null" json:"user_id,string"`
	LibraryID    int    `gorm:"uniqueIndex:uq_igo_leak_targets_user_lib,priority:2;not null" json:"library_id"`
	LibraryName  string `gorm:"size:255;not null;default:''" json:"library_name"`
	Floor        string `gorm:"size:64;not null;default:''" json:"floor"`
	ScanPriority int    `gorm:"index:idx_igo_leak_targets_user_prio,priority:2;not null;default:0" json:"scan_priority"`
}

// TableName returns the plugin-prefixed table name.
func (GlobalLeakTarget) TableName() string { return consts.TableGlobalLeakTargets }

// GlobalLeakBlacklistSeat is a seat excluded from global-leak scanning.
type GlobalLeakBlacklistSeat struct {
	ID        uint64 `gorm:"primaryKey" json:"id,string"`
	UserID    uint64 `gorm:"uniqueIndex:uq_igo_leak_blacklist_user_lib_seat,priority:1;index:idx_igo_leak_blacklist_user_lib,priority:1;not null" json:"user_id,string"`
	LibraryID int    `gorm:"uniqueIndex:uq_igo_leak_blacklist_user_lib_seat,priority:2;index:idx_igo_leak_blacklist_user_lib,priority:2;not null" json:"library_id"`
	SeatKey   string `gorm:"uniqueIndex:uq_igo_leak_blacklist_user_lib_seat,priority:3;size:64;not null" json:"seat_key"`
	SeatName  string `gorm:"size:128;not null;default:''" json:"seat_name"`
}

// TableName returns the plugin-prefixed table name.
func (GlobalLeakBlacklistSeat) TableName() string { return consts.TableGlobalLeakBlacklist }
