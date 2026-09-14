// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetVenue returns the bound venue for userID, or nil if missing.
func GetVenue(ctx context.Context, userID uint64) (*entity.Venue, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.Venue
	err = gdb.Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertVenue inserts or replaces the bound venue keyed by user_id.
func UpsertVenue(ctx context.Context, row *entity.Venue) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"library_id", "name", "floor", "is_open", "total_seats", "used_seats", "booked_seats", "updated_at"}),
	}).Create(row).Error
}

// DeleteVenue unbinds the venue for userID.
func DeleteVenue(ctx context.Context, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Where("user_id = ?", userID).Delete(&entity.Venue{}).Error
}
