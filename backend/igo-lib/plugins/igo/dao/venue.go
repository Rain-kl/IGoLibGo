// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"

	"gorm.io/gorm/clause"
)

// GetVenue returns the locked venue for userID, or nil if none.
func GetVenue(ctx context.Context, userID uint64) (*entity.Venue, error) {
	var row entity.Venue
	return getByUser(ctx, userID, &row)
}

// UpsertVenue inserts or replaces the bound venue keyed by user_id.
func UpsertVenue(ctx context.Context, row *entity.Venue) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: colUserID}},
		DoUpdates: clause.AssignmentColumns([]string{colLibraryID, "name", "floor", "is_open", "total_seats", "used_seats", "booked_seats", colUpdatedAt}),
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
