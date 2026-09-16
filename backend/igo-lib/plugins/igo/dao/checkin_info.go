// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// CreateCheckInInfo inserts a check-in info row and assigns a snowflake ID when missing.
func CreateCheckInInfo(ctx context.Context, row *entity.CheckInInfo) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	ensureID(&row.ID)
	now := time.Now().UTC()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = now
	}
	return gdb.Create(row).Error
}

// GetCheckInInfoByUser returns the check-in info owned by userID, or nil if missing.
func GetCheckInInfoByUser(ctx context.Context, id, userID uint64) (*entity.CheckInInfo, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.CheckInInfo
	if err := gdb.Where("id = ? AND user_id = ?", id, userID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListCheckInInfosByUser lists check-in infos owned by userID ordered by created_at ASC.
func ListCheckInInfosByUser(ctx context.Context, userID uint64) ([]entity.CheckInInfo, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var list []entity.CheckInInfo
	if err := gdb.Where("user_id = ?", userID).Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateCheckInInfo persists an existing check-in info row.
func UpdateCheckInInfo(ctx context.Context, row *entity.CheckInInfo) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	row.UpdatedAt = time.Now().UTC()
	return gdb.Save(row).Error
}

// DeleteCheckInInfo removes a check-in info owned by userID.
func DeleteCheckInInfo(ctx context.Context, id, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.CheckInInfo{}).Error
}

// ListPipelineIDsByCheckInInfo returns pipeline card IDs that reference the info.
func ListPipelineIDsByCheckInInfo(ctx context.Context, userID, infoID uint64) ([]string, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var ids []string
	err = gdb.Model(&entity.PipelineConfig{}).
		Where("user_id = ? AND checkin_info_id = ?", userID, infoID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}
