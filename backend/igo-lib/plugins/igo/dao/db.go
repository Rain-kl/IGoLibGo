// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package dao provides GORM access for igo plugin tables.
package dao

import (
	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrDBNotReady is returned when DBService has not been bound yet.
var ErrDBNotReady = errors.New("igo: database not ready")

var dbSvc contracts.DBService

// Common database column names.
const (
	colUserID    = "user_id"
	colUpdatedAt = "updated_at"
)

// SetDBService binds the platform DBService. Wired via ctx.Bind.
func SetDBService(svc contracts.DBService) {
	dbSvc = svc
}

func db(ctx context.Context) (*gorm.DB, error) {
	if dbSvc == nil {
		return nil, ErrDBNotReady
	}
	return dbSvc.DB(ctx), nil
}

func ensureID(id *uint64) {
	if id == nil || *id != 0 {
		return
	}
	*id = idgen.NextUint64ID()
}

func getByUser[T any](ctx context.Context, userID uint64, dest *T) (*T, error) {
	gdb, err := db(ctx)
	if err != nil {
		return nil, err
	}
	err = gdb.Where(colUserID+" = ?", userID).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return dest, nil
}
