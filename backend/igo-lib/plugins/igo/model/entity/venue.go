// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// Venue is the currently bound library for one Wavelet user.
type Venue struct {
	ID          uint64    `gorm:"primaryKey" json:"id,string"`
	UserID      uint64    `gorm:"uniqueIndex:uq_w_igo_venues_user;not null" json:"user_id,string"`
	LibraryID   int       `gorm:"not null" json:"library_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Floor       string    `gorm:"size:64;not null;default:''" json:"floor"`
	IsOpen      bool      `gorm:"not null;default:false" json:"is_open"`
	TotalSeats  int       `gorm:"not null;default:0" json:"total_seats"`
	UsedSeats   int       `gorm:"not null;default:0" json:"used_seats"`
	BookedSeats int       `gorm:"not null;default:0" json:"booked_seats"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the plugin-prefixed table name.
func (Venue) TableName() string { return consts.TableVenues }
