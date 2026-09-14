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
)

// ErrNotImplemented is returned by service stubs before business logic is migrated.
var ErrNotImplemented = errors.New("not_implemented")

// SupportedTaskKinds lists task kinds accepted by /tasks/:kind/*.
var SupportedTaskKinds = map[string]struct{}{
	TaskKindGrab:       {},
	TaskKindOccupy:     {},
	TaskKindGlobalLeak: {},
	TaskKindTomorrow:   {},
}
