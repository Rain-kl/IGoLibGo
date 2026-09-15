// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

import "time"

// PipelineConfigDTO is the data transfer object for a pipeline configuration card.
type PipelineConfigDTO struct {
	ID               string     `json:"id"`
	UserID           uint64     `json:"user_id,string"`
	Name             string     `json:"name"`
	HasCookie        bool       `json:"has_cookie"`
	CookieMasked     string     `json:"cookie_masked"`
	CookieExpiresAt  *time.Time `json:"cookie_expires_at,omitempty"`
	LibraryID        int        `json:"library_id"`
	LibraryName      string     `json:"library_name"`
	Floor            string     `json:"floor"`
	SeatKey          string     `json:"seat_key"`
	SeatName         string     `json:"seat_name"`
	AutoCheckin      bool       `json:"auto_checkin"`
	HasCheckinToken  bool       `json:"has_checkin_token"`
	CheckinExpiresAt *time.Time `json:"checkin_expires_at,omitempty"`
	BeaconUUID       string     `json:"beacon_uuid"`
	Major            int        `json:"major"`
	Minor            int        `json:"minor"`
	Latitude         string     `json:"latitude"`
	Longitude        string     `json:"longitude"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreatePipelineConfigRequest represents the payload for creating a pipeline config card.
type CreatePipelineConfigRequest struct {
	ID           string `json:"id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Cookie       string `json:"cookie" binding:"required"`
	LibraryID    int    `json:"library_id" binding:"required"`
	LibraryName  string `json:"library_name"`
	Floor        string `json:"floor"`
	SeatKey      string `json:"seat_key" binding:"required"`
	SeatName     string `json:"seat_name"`
	AutoCheckin  bool   `json:"auto_checkin"`
	CheckinToken string `json:"checkin_token"`
	BeaconUUID   string `json:"beacon_uuid"`
	Major        int    `json:"major"`
	Minor        int    `json:"minor"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
}

// UpdatePipelineConfigRequest represents the payload for updating a pipeline config card.
type UpdatePipelineConfigRequest struct {
	Name         string `json:"name"`
	Cookie       string `json:"cookie"`
	LibraryID    int    `json:"library_id"`
	LibraryName  string `json:"library_name"`
	Floor        string `json:"floor"`
	SeatKey      string `json:"seat_key"`
	SeatName     string `json:"seat_name"`
	AutoCheckin  bool   `json:"auto_checkin"`
	CheckinToken string `json:"checkin_token"`
	BeaconUUID   string `json:"beacon_uuid"`
	Major        int    `json:"major"`
	Minor        int    `json:"minor"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
}

// RunPipelineRequest represents optional override credentials for execution.
type RunPipelineRequest struct {
	Cookie       string `json:"cookie,omitempty"`
	CheckinToken string `json:"checkin_token,omitempty"`
}

// PipelineRunResult represents the outcome of executing a pipeline config card.
type PipelineRunResult struct {
	Success           bool   `json:"success"`
	ConfigID          string `json:"config_id"`
	Name              string `json:"name"`
	NeedAuth          string `json:"need_auth,omitempty"` // "LOGIN", "CHECKIN", or ""
	AuthURL           string `json:"auth_url,omitempty"`
	Message           string `json:"message"`
	ReservationStatus string `json:"reservation_status,omitempty"`
	CheckinStatus     string `json:"checkin_status,omitempty"`
	ExecutedAt        string `json:"executed_at"`
}

// HelperVerifySessionRequest is the payload to verify a raw cookie and list libraries.
type HelperVerifySessionRequest struct {
	Cookie string `json:"cookie" binding:"required"`
}

// HelperLibraryLayoutRequest is the payload to fetch layout for a specific library with a cookie.
type HelperLibraryLayoutRequest struct {
	Cookie    string `json:"cookie" binding:"required"`
	LibraryID int    `json:"library_id" binding:"required"`
}

// HelperVerifyCheckinRequest is the payload to test a checkin code/token.
type HelperVerifyCheckinRequest struct {
	TokenOrCode string `json:"token_or_code" binding:"required"`
}
