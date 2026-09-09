// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package dao provides data access objects for the custom_example plugin.
package dao

import (
	"Wavelet/core/contracts"
	"Wavelet/downstream/plugins/custom_example/model/entity"
	"Wavelet/pkg/util"
	"context"
)

var dbSvc contracts.DBService

// SetDBService sets the database service for the DAO layer.
func SetDBService(svc contracts.DBService) {
	dbSvc = svc
}

// CreateGreeting inserts a new CustomGreeting record.
func CreateGreeting(ctx context.Context, greeting *entity.CustomGreeting) error {
	if dbSvc == nil {
		return nil
	}
	return dbSvc.DB(ctx).Create(greeting).Error
}

// GetGreetingByID retrieves a CustomGreeting by ID.
func GetGreetingByID(ctx context.Context, id int64) (*entity.CustomGreeting, error) {
	if dbSvc == nil {
		return nil, nil
	}
	var greeting entity.CustomGreeting
	if err := dbSvc.DB(ctx).First(&greeting, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &greeting, nil
}

// ListGreetings lists greetings by recipient prefix (using EscapeLike for injection prevention).
func ListGreetings(ctx context.Context, recipientKeyword string) ([]entity.CustomGreeting, error) {
	if dbSvc == nil {
		return nil, nil
	}
	var list []entity.CustomGreeting
	query := dbSvc.DB(ctx)
	if recipientKeyword != "" {
		escaped := util.EscapeLike(recipientKeyword) + "%"
		query = query.Where("recipient LIKE ? ESCAPE '\\'", escaped)
	}
	if err := query.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
