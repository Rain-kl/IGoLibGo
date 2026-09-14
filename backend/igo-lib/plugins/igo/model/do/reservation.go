// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// ReservationResponse is the current TraceInt reservation.
type ReservationResponse struct {
	HasReservation   bool   `json:"has_reservation"`
	ReservationToken string `json:"reservation_token,omitempty"`
	LibraryID        int    `json:"library_id,omitempty"`
	LibraryName      string `json:"library_name,omitempty"`
	SeatKey          string `json:"seat_key,omitempty"`
	SeatName         string `json:"seat_name,omitempty"`
	ExpirationTime   string `json:"expiration_time,omitempty"`
}

// CancelReservationRequest cancels the current reservation.
type CancelReservationRequest struct {
	StopOccupyFirst bool `json:"stop_occupy_first"`
}

// ReservationOperationResponse is the result of refresh/cancel.
type ReservationOperationResponse struct {
	Reservation ReservationResponse `json:"reservation"`
	Message     string              `json:"message,omitempty"`
}
