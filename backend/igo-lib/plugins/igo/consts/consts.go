// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package consts defines constants and error codes for the igo plugin.
package consts

import "errors"

const (
	// PluginName is the Cordis plugin identifier.
	PluginName = "igo"

	// APIPrefix is the HTTP group prefix for all igo endpoints.
	APIPrefix = "/api/v1/igo"

	// CodeNotImplemented is the api-design error code for stub handlers.
	CodeNotImplemented = "not_implemented"

	// CodeValidationError is returned when request binding fails.
	CodeValidationError = "validation_error"

	// CodeInvalidID is returned when a path id cannot be parsed.
	CodeInvalidID = "invalid_id"

	// CodeInvalidTaskKind is returned when tasks/:kind is not supported.
	CodeInvalidTaskKind = "invalid_task_kind"

	// TaskKindGrab is the grab-seat coordinator.
	TaskKindGrab = "grab"
	// TaskKindOccupy is the occupy-seat coordinator.
	TaskKindOccupy = "occupy"
	// TaskKindGlobalLeak is the global-leak coordinator.
	TaskKindGlobalLeak = "global-leak"
	// TaskKindTomorrow is the tomorrow-reservation coordinator.
	TaskKindTomorrow = "tomorrow"

	TableSessions            = "w_igo_sessions"
	TableVenues              = "w_igo_venues"
	TableFavorites           = "w_igo_favorites"
	TableSeatLabels          = "w_igo_seat_labels"
	TableProtocolOverrides   = "w_igo_protocol_overrides"
	TableSettings            = "w_igo_settings"
	TableTaskRuns            = "w_igo_task_runs"
	TableTaskLaunchHistory   = "w_igo_task_launch_history"
	TableGlobalLeakTargets   = "w_igo_global_leak_targets"
	TableGlobalLeakBlacklist = "w_igo_global_leak_blacklist"
	TableCheckInSessions     = "w_igo_checkin_sessions"
	TableDashboardMetrics    = "w_igo_dashboard_metrics"
	TableWebDAV              = "w_igo_webdav"
)

// OwnedTables is the igo plugin's single-owner table list (must match Goose SQL).
var OwnedTables = []string{
	TableSessions,
	TableVenues,
	TableFavorites,
	TableSeatLabels,
	TableProtocolOverrides,
	TableSettings,
	TableTaskRuns,
	TableTaskLaunchHistory,
	TableGlobalLeakTargets,
	TableGlobalLeakBlacklist,
	TableCheckInSessions,
	TableDashboardMetrics,
	TableWebDAV,
}

// ErrNotImplemented is returned by service stubs before business logic is migrated.
var ErrNotImplemented = errors.New("not_implemented")

// SupportedTaskKinds lists task kinds accepted by /tasks/:kind/*.
var SupportedTaskKinds = map[string]struct{}{
	TaskKindGrab:       {},
	TaskKindOccupy:     {},
	TaskKindGlobalLeak: {},
	TaskKindTomorrow:   {},
}
