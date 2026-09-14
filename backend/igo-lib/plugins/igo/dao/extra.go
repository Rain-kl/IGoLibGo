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

func getByUser[T any](ctx context.Context, userID uint64, dest *T) (*T, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	err = gdb.Where("user_id = ?", userID).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return dest, nil
}

func upsertByUser(ctx context.Context, id *uint64, row any, updateCols []string) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(id)
	return gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns(updateCols),
	}).Create(row).Error
}

// GetProtocolOverride returns protocol overrides JSON for a user.
func GetProtocolOverride(ctx context.Context, userID uint64) (*entity.ProtocolOverride, error) {
	var row entity.ProtocolOverride
	return getByUser(ctx, userID, &row)
}

// UpsertProtocolOverride stores protocol overrides for a user.
func UpsertProtocolOverride(ctx context.Context, row *entity.ProtocolOverride) error {
	return upsertByUser(ctx, &row.ID, row, []string{"overrides", "updated_at"})
}

// GetSettings returns settings JSON for a user.
func GetSettings(ctx context.Context, userID uint64) (*entity.Settings, error) {
	var row entity.Settings
	return getByUser(ctx, userID, &row)
}

// UpsertSettings stores settings JSON for a user.
func UpsertSettings(ctx context.Context, row *entity.Settings) error {
	return upsertByUser(ctx, &row.ID, row, []string{"payload", "updated_at"})
}

// GetCheckInSession returns the remote-check-in session for a user.
func GetCheckInSession(ctx context.Context, userID uint64) (*entity.CheckInSession, error) {
	var row entity.CheckInSession
	return getByUser(ctx, userID, &row)
}

// UpsertCheckInSession stores the remote-check-in session for a user.
func UpsertCheckInSession(ctx context.Context, row *entity.CheckInSession) error {
	return upsertByUser(ctx, &row.ID, row, []string{"token", "saved_at", "expires_at", "can_auto_restore", "updated_at"})
}

// DeleteCheckInSession removes the remote-check-in session for a user.
func DeleteCheckInSession(ctx context.Context, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Where("user_id = ?", userID).Delete(&entity.CheckInSession{}).Error
}

// GetDashboardMetrics returns home counters for a user.
func GetDashboardMetrics(ctx context.Context, userID uint64) (*entity.DashboardMetrics, error) {
	var row entity.DashboardMetrics
	return getByUser(ctx, userID, &row)
}

// UpsertDashboardMetrics stores home counters for a user.
func UpsertDashboardMetrics(ctx context.Context, row *entity.DashboardMetrics) error {
	return upsertByUser(ctx, &row.ID, row, []string{"historical_success_count", "total_guard_seconds", "updated_at"})
}

// GetWebDAV returns WebDAV settings for a user.
func GetWebDAV(ctx context.Context, userID uint64) (*entity.WebDAV, error) {
	var row entity.WebDAV
	return getByUser(ctx, userID, &row)
}

// UpsertWebDAV stores WebDAV settings for a user.
func UpsertWebDAV(ctx context.Context, row *entity.WebDAV) error {
	return upsertByUser(ctx, &row.ID, row, []string{"endpoint", "remote_directory", "username", "password", "tls_verify_mode", "updated_at"})
}
