// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"

	"gorm.io/gorm/clause"
)

// GetSession returns the TraceInt session for userID, or nil if missing.
func GetSession(ctx context.Context, userID uint64) (*entity.Session, error) {
	var row entity.Session
	return getByUser(ctx, userID, &row)
}

// UpsertSession inserts or replaces the session keyed by user_id.
func UpsertSession(ctx context.Context, row *entity.Session) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: colUserID}},
		DoUpdates: clause.AssignmentColumns([]string{"cookie", "source", "saved_at", "expires_at", "can_auto_restore", colUpdatedAt}),
	}).Create(row).Error
}

// ListSessions returns all stored TraceInt sessions.
func ListSessions(ctx context.Context) ([]entity.Session, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entity.Session
	err = gdb.Find(&rows).Error
	return rows, err
}

// DeleteSession removes the session for userID.
func DeleteSession(ctx context.Context, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Where("user_id = ?", userID).Delete(&entity.Session{}).Error
}
