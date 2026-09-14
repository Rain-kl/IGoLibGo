// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// Favorite is a starred seat in a library.
type Favorite struct {
	ID        uint64    `gorm:"primaryKey" json:"id,string"`
	UserID    uint64    `gorm:"uniqueIndex:uq_igo_favorites_user_lib_seat,priority:1;index:idx_igo_favorites_user_lib,priority:1;not null" json:"user_id,string"`
	LibraryID int       `gorm:"uniqueIndex:uq_igo_favorites_user_lib_seat,priority:2;index:idx_igo_favorites_user_lib,priority:2;not null" json:"library_id"`
	SeatKey   string    `gorm:"uniqueIndex:uq_igo_favorites_user_lib_seat,priority:3;size:64;not null" json:"seat_key"`
	SeatName  string    `gorm:"size:128;not null;default:''" json:"seat_name"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the plugin-prefixed table name.
func (Favorite) TableName() string { return consts.TableFavorites }

// SeatLabel is a user-defined label on a seat.
type SeatLabel struct {
	ID        uint64    `gorm:"primaryKey" json:"id,string"`
	UserID    uint64    `gorm:"uniqueIndex:uq_igo_seat_labels_user_lib_seat,priority:1;index:idx_igo_seat_labels_user_lib,priority:1;not null" json:"user_id,string"`
	LibraryID int       `gorm:"uniqueIndex:uq_igo_seat_labels_user_lib_seat,priority:2;index:idx_igo_seat_labels_user_lib,priority:2;not null" json:"library_id"`
	SeatKey   string    `gorm:"uniqueIndex:uq_igo_seat_labels_user_lib_seat,priority:3;size:64;not null" json:"seat_key"`
	SeatName  string    `gorm:"size:128;not null;default:''" json:"seat_name"`
	LabelText string    `gorm:"size:64;not null" json:"label_text"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (SeatLabel) TableName() string { return consts.TableSeatLabels }
