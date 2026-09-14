// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package do defines domain objects and API request/response DTOs for the custom_example plugin.
package do

// CreateGreetingRequest defines the input payload for creating a custom greeting.
type CreateGreetingRequest struct {
	Recipient string `json:"recipient" binding:"required,min=1,max=64"`
	Message   string `json:"message" binding:"required,min=1,max=255"`
}

// GreetingResponse defines the output DTO for greeting API endpoints.
type GreetingResponse struct {
	ID        int64  `json:"id"`
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}
