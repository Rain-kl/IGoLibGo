// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"

	"gorm.io/gorm"
)

// ListGlobalLeakTargets returns scan-priority venues for a user.
func ListGlobalLeakTargets(ctx context.Context, userID uint64) ([]entity.GlobalLeakTarget, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entity.GlobalLeakTarget
	err = gdb.Where("user_id = ?", userID).Order("scan_priority ASC").Find(&rows).Error
	return rows, err
}

// ReplaceGlobalLeakTargets atomically replaces selected leak venues.
func ReplaceGlobalLeakTargets(ctx context.Context, userID uint64, targets []entity.GlobalLeakTarget) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&entity.GlobalLeakTarget{}).Error; err != nil {
			return err
		}
		if len(targets) == 0 {
			return nil
		}
		for i := range targets {
			targets[i].UserID = userID
			ensureID(&targets[i].ID)
		}
		return tx.Create(&targets).Error
	})
}

// ListGlobalLeakBlacklist returns blacklisted seats, optionally filtered by library IDs.
func ListGlobalLeakBlacklist(ctx context.Context, userID uint64, libraryIDs []int) ([]entity.GlobalLeakBlacklistSeat, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	q := gdb.Where("user_id = ?", userID)
	if len(libraryIDs) > 0 {
		q = q.Where("library_id IN ?", libraryIDs)
	}
	var rows []entity.GlobalLeakBlacklistSeat
	err = q.Order("library_id, seat_name").Find(&rows).Error
	return rows, err
}

// ReplaceGlobalLeakBlacklistForLibraries replaces blacklist rows for the submitted libraries only.
func ReplaceGlobalLeakBlacklistForLibraries(ctx context.Context, userID uint64, libraryIDs []int, seats []entity.GlobalLeakBlacklistSeat) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	if len(libraryIDs) == 0 {
		return nil
	}
	return gdb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND library_id IN ?", userID, libraryIDs).
			Delete(&entity.GlobalLeakBlacklistSeat{}).Error; err != nil {
			return err
		}
		if len(seats) == 0 {
			return nil
		}
		for i := range seats {
			seats[i].UserID = userID
			ensureID(&seats[i].ID)
		}
		return tx.Create(&seats).Error
	})
}
