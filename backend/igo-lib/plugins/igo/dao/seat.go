// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListFavorites returns favorite seats for a user and library.
func ListFavorites(ctx context.Context, userID uint64, libraryID int) ([]entity.Favorite, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entity.Favorite
	err = gdb.Where("user_id = ? AND library_id = ?", userID, libraryID).Order("seat_name").Find(&rows).Error
	return rows, err
}

// ReplaceFavorites atomically replaces favorites for a user and library.
func ReplaceFavorites(ctx context.Context, userID uint64, libraryID int, seats []entity.Favorite) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND library_id = ?", userID, libraryID).Delete(&entity.Favorite{}).Error; err != nil {
			return err
		}
		if len(seats) == 0 {
			return nil
		}
		for i := range seats {
			seats[i].UserID = userID
			seats[i].LibraryID = libraryID
			ensureID(&seats[i].ID)
		}
		return tx.Create(&seats).Error
	})
}

// ListSeatLabels returns labels for a user and library.
func ListSeatLabels(ctx context.Context, userID uint64, libraryID int) ([]entity.SeatLabel, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entity.SeatLabel
	err = gdb.Where("user_id = ? AND library_id = ?", userID, libraryID).Order("seat_name").Find(&rows).Error
	return rows, err
}

// UpsertSeatLabels writes labels for the given seats (does not clear others).
func UpsertSeatLabels(ctx context.Context, userID uint64, libraryID int, labels []entity.SeatLabel) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	if len(labels) == 0 {
		return nil
	}
	for i := range labels {
		labels[i].UserID = userID
		labels[i].LibraryID = libraryID
		ensureID(&labels[i].ID)
	}
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: colUserID}, {Name: "library_id"}, {Name: "seat_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"seat_name", "label_text", colUpdatedAt}),
	}).Create(&labels).Error
}

// DeleteSeatLabels removes labels by seat key.
func DeleteSeatLabels(ctx context.Context, userID uint64, libraryID int, seatKeys []string) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	if len(seatKeys) == 0 {
		return nil
	}
	return gdb.Where("user_id = ? AND library_id = ? AND seat_key IN ?", userID, libraryID, seatKeys).
		Delete(&entity.SeatLabel{}).Error
}
