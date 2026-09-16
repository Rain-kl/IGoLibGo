// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

import "time"

// CheckInInfoDTO is reusable Beacon simulation parameters with no credentials.
type CheckInInfoDTO struct {
	ID         uint64    `json:"id,string"`
	Name       string    `json:"name"`
	BeaconUUID string    `json:"beacon_uuid"`
	Major      int       `json:"major"`
	Minor      int       `json:"minor"`
	Latitude   string    `json:"latitude"`
	Longitude  string    `json:"longitude"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCheckInInfoRequest creates a check-in info. Beacon fields may be empty.
type CreateCheckInInfoRequest struct {
	Name       string `json:"name" binding:"required"`
	BeaconUUID string `json:"beacon_uuid"`
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}

// UpdateCheckInInfoRequest updates a check-in info.
type UpdateCheckInInfoRequest struct {
	Name       string `json:"name"`
	BeaconUUID string `json:"beacon_uuid"`
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
}

// SignCheckInInfoRequest signs using an account's token and this info's Beacon.
type SignCheckInInfoRequest struct {
	AccountID           uint64 `json:"account_id,string" binding:"required"`
	ExpectedLibraryID   int    `json:"expected_library_id" binding:"required"`
	ExpectedLibraryName string `json:"expected_library_name"`
}
