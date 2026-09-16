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

// ErrPipelineConfigNotFound is returned when a pipeline config card is not found.
var ErrPipelineConfigNotFound = errors.New("pipeline config not found")

// CreatePipelineConfig inserts a new pipeline configuration.
func CreatePipelineConfig(ctx context.Context, row *entity.PipelineConfig) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = now
	}
	return gdb.Create(row).Error
}

// GetPipelineConfig retrieves a pipeline configuration by primary key ID.
func GetPipelineConfig(ctx context.Context, id string) (*entity.PipelineConfig, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.PipelineConfig
	if err := gdb.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetPipelineConfigByUser retrieves a pipeline config by ID and owner userID.
func GetPipelineConfigByUser(ctx context.Context, id string, userID uint64) (*entity.PipelineConfig, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.PipelineConfig
	if err := gdb.Where("id = ? AND user_id = ?", id, userID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListPipelineConfigsByUser lists all pipeline configs owned by userID ordered by created_at DESC.
func ListPipelineConfigsByUser(ctx context.Context, userID uint64) ([]entity.PipelineConfig, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var list []entity.PipelineConfig
	if err := gdb.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListAllPipelineConfigs lists all pipeline configs across the platform.
func ListAllPipelineConfigs(ctx context.Context) ([]entity.PipelineConfig, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var list []entity.PipelineConfig
	if err := gdb.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdatePipelineConfig updates an existing pipeline config.
func UpdatePipelineConfig(ctx context.Context, row *entity.PipelineConfig) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	row.UpdatedAt = time.Now().UTC()
	res := gdb.Model(row).Where("id = ? AND user_id = ?", row.ID, row.UserID).Updates(map[string]any{
		"name":               row.Name,
		colCookie:            row.Cookie,
		"cookie_expires_at":  row.CookieExpiresAt,
		colLibraryID:         row.LibraryID,
		"library_name":       row.LibraryName,
		"floor":              row.Floor,
		"seat_key":           row.SeatKey,
		"seat_name":          row.SeatName,
		"auto_checkin":       row.AutoCheckin,
		"checkin_token":      row.CheckinToken,
		"checkin_expires_at": row.CheckinExpiresAt,
		"beacon_uuid":        row.BeaconUUID,
		"major":              row.Major,
		"minor":              row.Minor,
		"latitude":           row.Latitude,
		"longitude":          row.Longitude,
		"account_id":         row.AccountID,
		"checkin_account_id": row.CheckinAccountID,
		"checkin_info_id":    row.CheckinInfoID,
		colUpdatedAt:         row.UpdatedAt,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPipelineConfigNotFound
	}
	return nil
}

// DeletePipelineConfig removes a pipeline config by ID and userID.
func DeletePipelineConfig(ctx context.Context, id string, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	res := gdb.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.PipelineConfig{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPipelineConfigNotFound
	}
	return nil
}

// UpdatePipelineCookie updates the cookie and expiry for a specific pipeline card.
func UpdatePipelineCookie(ctx context.Context, id string, cookie string, exp *time.Time) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Model(&entity.PipelineConfig{}).Where("id = ?", id).Updates(map[string]any{
		"cookie":            cookie,
		"cookie_expires_at": exp,
		"updated_at":        time.Now().UTC(),
	}).Error
}

// UpdatePipelineCheckinToken updates the checkin token and expiry for a specific pipeline card.
func UpdatePipelineCheckinToken(ctx context.Context, id string, token string, exp *time.Time) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Model(&entity.PipelineConfig{}).Where("id = ?", id).Updates(map[string]any{
		"checkin_token":      token,
		"checkin_expires_at": exp,
		"updated_at":         time.Now().UTC(),
	}).Error
}
