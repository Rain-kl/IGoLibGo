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

// GetTaskRun returns the coordinator snapshot for a kind, or nil if missing.
func GetTaskRun(ctx context.Context, userID uint64, kind string) (*entity.TaskRun, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.TaskRun
	err = gdb.Where("user_id = ? AND kind = ?", userID, kind).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListTaskRuns returns all coordinator snapshots for a user.
func ListTaskRuns(ctx context.Context, userID uint64) ([]entity.TaskRun, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var rows []entity.TaskRun
	err = gdb.Where("user_id = ?", userID).Order("kind").Find(&rows).Error
	return rows, err
}

// UpsertTaskRun inserts or replaces a coordinator snapshot.
func UpsertTaskRun(ctx context.Context, row *entity.TaskRun) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	return gdb.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "kind"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"state", "title", "message", "plan_json",
			"started_at", "last_updated_at", "last_request_at",
			"poll_count", "request_count", "reason",
		}),
	}).Create(row).Error
}

// ListTaskLaunchHistory returns recent launch records for a kind.
func ListTaskLaunchHistory(ctx context.Context, userID uint64, kind string, limit int) ([]entity.TaskLaunchHistory, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}
	var rows []entity.TaskLaunchHistory
	err = gdb.Where("user_id = ? AND kind = ?", userID, kind).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// UpsertTaskLaunchHistory inserts or refreshes a launch record by fingerprint.
func UpsertTaskLaunchHistory(ctx context.Context, row *entity.TaskLaunchHistory) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	if err := gdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "kind"}, {Name: "fingerprint"}},
		DoUpdates: clause.AssignmentColumns([]string{"record_id", "recorded_at", "payload_json"}),
	}).Create(row).Error; err != nil {
		return err
	}
	return pruneTaskLaunchHistory(gdb, row.UserID, row.Kind, 5)
}

func pruneTaskLaunchHistory(gdb *gorm.DB, userID uint64, kind string, keep int) error {
	var ids []uint64
	if err := gdb.Model(&entity.TaskLaunchHistory{}).
		Where("user_id = ? AND kind = ?", userID, kind).
		Order("id DESC").
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) <= keep {
		return nil
	}
	return gdb.Where("user_id = ? AND kind = ? AND id IN ?", userID, kind, ids[keep:]).
		Delete(&entity.TaskLaunchHistory{}).Error
}
