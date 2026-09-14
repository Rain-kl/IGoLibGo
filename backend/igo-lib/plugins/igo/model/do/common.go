// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package do defines API request/response DTOs for the igo plugin.
package do

// SeatRef is a seat identity used by favorites, labels, and task targets.
type SeatRef struct {
	SeatKey  string `json:"seat_key" binding:"required"`
	SeatName string `json:"seat_name"`
}

// SeatSnapshot is a live seat on a library layout.
type SeatSnapshot struct {
	SeatKey    string `json:"seat_key"`
	SeatName   string `json:"seat_name"`
	IsOccupied bool   `json:"is_occupied"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	SeatStatus *int   `json:"seat_status,omitempty"`
}

// SeatLabel is a user-defined label on a seat.
type SeatLabel struct {
	SeatKey  string `json:"seat_key"`
	SeatName string `json:"seat_name"`
	Text     string `json:"text"`
}

// LibrarySummary is a venue row from TraceInt.
type LibrarySummary struct {
	LibraryID   int    `json:"library_id"`
	Name        string `json:"name"`
	Floor       string `json:"floor"`
	IsOpen      bool   `json:"is_open"`
	TotalSeats  int    `json:"total_seats"`
	UsedSeats   int    `json:"used_seats"`
	BookedSeats int    `json:"booked_seats"`
}

// CoordinatorStatus is a task engine snapshot.
type CoordinatorStatus struct {
	Kind          string `json:"kind"`
	State         string `json:"state"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	StartedAt     string `json:"started_at,omitempty"`
	LastUpdatedAt string `json:"last_updated_at,omitempty"`
	PollCount     int    `json:"poll_count"`
	RequestCount  int    `json:"request_count"`
	LastRequestAt string `json:"last_request_at,omitempty"`
	Reason        string `json:"reason,omitempty"`
	IsActive      bool   `json:"is_active"`
}

// QRCodeResponse is a WeChat authorization QR payload.
type QRCodeResponse struct {
	ImageDataURL string `json:"image_data_url"`
	AuthURL      string `json:"auth_url"`
	ExpiresAt    string `json:"expires_at,omitempty"`
}
