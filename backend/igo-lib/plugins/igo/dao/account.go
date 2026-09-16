// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao

import (
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateAccount inserts a new account and assigns a snowflake ID when missing.
func CreateAccount(ctx context.Context, row *entity.Account) error {
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

// GetAccountByUser returns the account owned by userID, or nil if missing.
func GetAccountByUser(ctx context.Context, id, userID uint64) (*entity.Account, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.Account
	if err := gdb.Where("id = ? AND user_id = ?", id, userID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListAccountsByUser lists accounts owned by userID ordered by created_at ASC.
func ListAccountsByUser(ctx context.Context, userID uint64) ([]entity.Account, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var list []entity.Account
	if err := gdb.Where("user_id = ?", userID).Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateAccount persists an existing account row.
func UpdateAccount(ctx context.Context, row *entity.Account) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	row.UpdatedAt = time.Now().UTC()
	return gdb.Save(row).Error
}

// DeleteAccount removes an account owned by userID.
func DeleteAccount(ctx context.Context, id, userID uint64) error {
	gdb, err := db(ctx)
	if err != nil {
		return err
	}
	return gdb.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Account{}).Error
}

// FindAccountByCookie finds a non-empty cookie match for the user.
func FindAccountByCookie(ctx context.Context, userID uint64, cookie string) (*entity.Account, error) {
	if strings.TrimSpace(cookie) == "" {
		return nil, nil
	}
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row entity.Account
	if err := gdb.Where("user_id = ? AND cookie = ?", userID, cookie).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListPipelineIDsByAccount returns pipeline card IDs that reference the account.
func ListPipelineIDsByAccount(ctx context.Context, userID, accountID uint64) ([]string, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var ids []string
	err = gdb.Model(&entity.PipelineConfig{}).
		Where("user_id = ? AND (account_id = ? OR checkin_account_id = ?)", userID, accountID, accountID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}
