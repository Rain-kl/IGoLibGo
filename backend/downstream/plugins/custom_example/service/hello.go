// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package service implements business logic for the custom_example plugin.
package service

import (
	"Wavelet/downstream/plugins/custom_example/dao"
	"Wavelet/downstream/plugins/custom_example/model/do"
	"Wavelet/downstream/plugins/custom_example/model/entity"
	"Wavelet/pkg/logger"
	"context"
	"time"
)

// HelloService defines business operations for the custom_example plugin.
type HelloService struct{}

// NewHelloService creates a new HelloService instance.
func NewHelloService() *HelloService {
	return &HelloService{}
}

// CreateGreeting creates a new greeting message.
func (s *HelloService) CreateGreeting(ctx context.Context, req do.CreateGreetingRequest) (*do.GreetingResponse, error) {
	item := &entity.CustomGreeting{
		Recipient: req.Recipient,
		Message:   req.Message,
	}
	if err := dao.CreateGreeting(ctx, item); err != nil {
		logger.ErrorF(ctx, "failed to create custom greeting: %v", err)
		return nil, err
	}

	return &do.GreetingResponse{
		ID:        item.ID,
		Recipient: item.Recipient,
		Message:   item.Message,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// GetGreeting retrieves a greeting by ID.
func (s *HelloService) GetGreeting(ctx context.Context, id int64) (*do.GreetingResponse, error) {
	item, err := dao.GetGreetingByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return &do.GreetingResponse{
		ID:        item.ID,
		Recipient: item.Recipient,
		Message:   item.Message,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}, nil
}
