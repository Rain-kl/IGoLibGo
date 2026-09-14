// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package entity defines database models for the custom_example plugin.
package entity

import "time"

// CustomGreeting represents a greeting message record in the database.
type CustomGreeting struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Recipient string    `gorm:"size:64;not null;index" json:"recipient"`
	Message   string    `gorm:"size:255;not null" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name with the plugin-specific prefix.
func (CustomGreeting) TableName() string {
	return "w_custom_greetings"
}
