// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// CheckInSessionResponse is the independent remote-check-in WeChat session.
type CheckInSessionResponse struct {
	Authorized     bool   `json:"authorized"`
	SavedAt        string `json:"saved_at,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	CanAutoRestore bool   `json:"can_auto_restore"`
}

// CheckInAuthFromCodeRequest authorizes remote check-in from a WeChat code.
type CheckInAuthFromCodeRequest struct {
	Code     string `json:"code" binding:"required"`
	Remember bool   `json:"remember"`
}

// CheckInDeviceResponse is profile + beacon UUIDs from the sign-in API.
type CheckInDeviceResponse struct {
	Nickname      string   `json:"nickname"`
	School        string   `json:"school"`
	StudentName   string   `json:"student_name"`
	StudentNumber string   `json:"student_number"`
	BeaconUUIDs   []string `json:"beacon_uuids"`
}

// CheckInSignRequest submits a Bluetooth beacon check-in.
type CheckInSignRequest struct {
	ExpectedLibraryID   int     `json:"expected_library_id" binding:"required"`
	ExpectedLibraryName string  `json:"expected_library_name"`
	BeaconUUID          string  `json:"beacon_uuid" binding:"required"`
	Major               int     `json:"major"`
	Minor               int     `json:"minor"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
}

// CheckInSignResponse is the result of a remote sign-in.
type CheckInSignResponse struct {
	Message        string `json:"message"`
	Status         *int   `json:"status,omitempty"`
	LibraryID      *int   `json:"library_id,omitempty"`
	LibraryName    string `json:"library_name,omitempty"`
	LibraryFloor   string `json:"library_floor,omitempty"`
	SeatKey        string `json:"seat_key,omitempty"`
	SeatName       string `json:"seat_name,omitempty"`
	SignedAt       string `json:"signed_at,omitempty"`
	ExpirationTime string `json:"expiration_time,omitempty"`
}

// CheckInAuthorizationResponse is returned after exchanging a WeChat code.
type CheckInAuthorizationResponse struct {
	Session              CheckInSessionResponse `json:"session"`
	Device               *CheckInDeviceResponse `json:"device,omitempty"`
	DeviceRefreshWarning string                 `json:"device_refresh_warning,omitempty"`
}
