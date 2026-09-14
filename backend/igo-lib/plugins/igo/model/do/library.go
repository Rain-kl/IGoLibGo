// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// LibraryLayoutResponse is a venue map plus seat snapshots.
type LibraryLayoutResponse struct {
	LibrarySummary
	Seats                  []SeatSnapshot `json:"seats"`
	MaxX                   *int           `json:"max_x,omitempty"`
	MaxY                   *int           `json:"max_y,omitempty"`
	AvailableSeats         int            `json:"available_seats"`
	InvalidLayoutItemCount int            `json:"invalid_layout_item_count"`
}

// LibraryRuleResponse is TraceInt opening/booking rules for a venue.
type LibraryRuleResponse struct {
	LibraryID        int    `json:"library_id"`
	AdvanceBooking   string `json:"advance_booking"`
	SeatTTLMinutes   string `json:"seat_ttl_minutes"`
	HoldTTLMinutes   string `json:"hold_ttl_minutes"`
	RenewTimeMinutes string `json:"renew_time_minutes"`
	HoldReasonJSON   string `json:"hold_reason_json"`
	CloseStartDate   string `json:"close_start_date,omitempty"`
	CloseEndDate     string `json:"close_end_date,omitempty"`
	OpenTime         int64  `json:"open_time"`
	OpenTimeText     string `json:"open_time_text"`
	CloseTime        int64  `json:"close_time"`
	CloseTimeText    string `json:"close_time_text"`
	ValidateTime     int    `json:"validate_time"`
}

// BoundLibraryResponse is the currently locked venue.
type BoundLibraryResponse struct {
	Bound   bool                   `json:"bound"`
	Library *LibrarySummary        `json:"library,omitempty"`
	Layout  *LibraryLayoutResponse `json:"layout,omitempty"`
}

// SaveFavoritesRequest replaces favorites for a venue.
type SaveFavoritesRequest struct {
	Seats []SeatRef `json:"seats" binding:"required"`
}

// SetSeatLabelsRequest writes the same label text onto selected seats.
type SetSeatLabelsRequest struct {
	Seats []SeatRef `json:"seats" binding:"required,min=1"`
	Text  string    `json:"text" binding:"required"`
}

// DeleteSeatLabelsRequest removes labels by seat key.
type DeleteSeatLabelsRequest struct {
	SeatKeys []string `json:"seat_keys" binding:"required,min=1"`
}
