// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// DashboardResponse is the home-page snapshot.
type DashboardResponse struct {
	Authorized             bool                `json:"authorized"`
	HeroStatus             string              `json:"hero_status"`
	HeroStatusDetail       string              `json:"hero_status_detail"`
	HistoricalSuccessCount int                 `json:"historical_success_count"`
	TotalGuardSeconds      int64               `json:"total_guard_seconds"`
	EngineSummary          string              `json:"engine_summary"`
	BoundLibrary           *LibrarySummary     `json:"bound_library,omitempty"`
	Reservation            ReservationResponse `json:"reservation"`
	Tasks                  []CoordinatorStatus `json:"tasks"`
}

// TaskListResponse lists all coordinators.
type TaskListResponse struct {
	Tasks []CoordinatorStatus `json:"tasks"`
}

// GrabStartRequest starts the grab-seat coordinator.
type GrabStartRequest struct {
	LibraryID           int       `json:"library_id" binding:"required"`
	LibraryName         string    `json:"library_name"`
	Seats               []SeatRef `json:"seats" binding:"required,min=1"`
	PollingMode         string    `json:"polling_mode"`
	ReservationStrategy string    `json:"reservation_strategy"`
	ScheduledStart      string    `json:"scheduled_start,omitempty"`
	PollingMinDelayMS   int       `json:"polling_min_delay_ms"`
	PollingMaxDelayMS   int       `json:"polling_max_delay_ms"`
}

// OccupyStartRequest starts the occupy-seat coordinator.
type OccupyStartRequest struct {
	ReReserveDelaySeconds int    `json:"re_reserve_delay_seconds" binding:"required"`
	CheckIntervalMode     string `json:"check_interval_mode"`
}

// GlobalLeakStartRequest starts the global-leak coordinator.
type GlobalLeakStartRequest struct {
	Libraries           []GlobalLeakLibraryTarget `json:"libraries" binding:"required,min=1"`
	ScanIntervalSeconds int                       `json:"scan_interval_seconds" binding:"required"`
}

// TomorrowStartRequest starts the tomorrow-reservation coordinator.
type TomorrowStartRequest struct {
	LibraryID          int     `json:"library_id" binding:"required"`
	LibraryName        string  `json:"library_name"`
	Seat               SeatRef `json:"seat" binding:"required"`
	ScheduledStart     string  `json:"scheduled_start" binding:"required"`
	ExecuteImmediately bool    `json:"execute_immediately"`
}

// GlobalLeakLibraryTarget is a venue in scan-priority order.
type GlobalLeakLibraryTarget struct {
	LibraryID   int    `json:"library_id" binding:"required"`
	LibraryName string `json:"library_name"`
	Floor       string `json:"floor"`
}

// TaskLaunchRecord is a recent grab or global-leak launch.
type TaskLaunchRecord struct {
	RecordID            string                    `json:"record_id"`
	Kind                string                    `json:"kind"`
	RecordedAt          string                    `json:"recorded_at"`
	LibraryID           int                       `json:"library_id,omitempty"`
	LibraryName         string                    `json:"library_name,omitempty"`
	Seats               []SeatRef                 `json:"seats,omitempty"`
	Libraries           []GlobalLeakLibraryTarget `json:"libraries,omitempty"`
	ScanIntervalSeconds int                       `json:"scan_interval_seconds,omitempty"`
	PollingMode         string                    `json:"polling_mode,omitempty"`
	ReservationStrategy string                    `json:"reservation_strategy,omitempty"`
}

// GlobalLeakBlacklistResponse maps library id to blacklisted seats.
type GlobalLeakBlacklistResponse struct {
	Items map[int][]SeatRef `json:"items"`
}

// SaveGlobalLeakBlacklistRequest atomically replaces submitted venue blacklists.
type SaveGlobalLeakBlacklistRequest struct {
	Items map[int][]SeatRef `json:"items" binding:"required"`
}

// SaveGlobalLeakSelectedLibrariesRequest persists scan-priority venues.
type SaveGlobalLeakSelectedLibrariesRequest struct {
	Libraries []GlobalLeakLibraryTarget `json:"libraries" binding:"required"`
}

// ActivityLogEntry is a user-visible task/session event.
type ActivityLogEntry struct {
	ID        int64  `json:"id,string"`
	Level     string `json:"level"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}
