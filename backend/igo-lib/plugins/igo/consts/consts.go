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

	// MaxTaskLaunchHistory is the maximum number of recent launch history records retained per task kind.
	MaxTaskLaunchHistory = 5
)

// Database table names owned by the igo plugin.
const (
	TableSessions            = "igo_sessions"
	TableVenues              = "igo_venues"
	TableFavorites           = "igo_favorites"
	TableSeatLabels          = "igo_seat_labels"
	TableProtocolOverrides   = "igo_protocol_overrides"
	TableSettings            = "igo_settings"
	TableTaskRuns            = "igo_task_runs"
	TableTaskLaunchHistory   = "igo_task_launch_history"
	TableGlobalLeakTargets   = "igo_global_leak_targets"
	TableGlobalLeakBlacklist = "igo_global_leak_blacklist"
	TableCheckInSessions     = "igo_checkin_sessions"
	TableDashboardMetrics    = "igo_dashboard_metrics"
	TablePipelineConfigs     = "igo_pipeline_configs"
	TableAccounts            = "igo_accounts"
	TableCheckInInfos        = "igo_checkin_infos"
)

// API error codes.
const (
	CodeSessionRequired = "session_required"
	CodeNotFound        = "not_found"
	CodeTraceInt        = "traceint_error"
	CodeConflict        = "conflict"
	CodeNeedAuth        = "need_auth"

	TaskTypeTick        = "igo:tick"
	TaskTypeCookieWatch = "igo:cookie_watch"

	PushGrabSucceeded       = "igo.grab_succeeded"
	PushOccupySucceeded     = "igo.occupy_succeeded"
	PushGlobalLeakSucceeded = "igo.global_leak_succeeded"
	PushTomorrowSucceeded   = "igo.tomorrow_succeeded"
	PushTaskFailed          = "igo.task_failed"
	PushCookieExpiring      = "igo.cookie_expiring"
	PushSessionInvalid      = "igo.session_invalid"

	CookieExpiringLead = 10 * 60 // seconds

	WeChatLoginAuthURL   = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2Fgraphql%3FoperationName%3Dindex%26query%3Dquery%257BuserAuth%257BtongJi%257Brank%257D%257D%257D&response_type=code&scope=snsapi_userinfo&state=1&connect_redirect=1#wechat_redirect"
	WeChatCheckinAuthURL = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A//wechat.v2.traceint.com/index.php/graphql%3FoperationName%3Dindex%26query%3Dquery%257BuserAuth%257BtongJi%257Brank%257D%257D%257D&response_type=code&scope=snsapi_userinfo&state=1&connect_redirect=1"

	PlaceholderCode      = "ReplaceMeByCode"
	PlaceholderReturnURL = "ReplaceMeByReturnUrl"
	PlaceholderSeatKey   = "ReplaceMeBySeatKey"
	PlaceholderLibID     = "ReplaceMeByLibID"
	PlaceholderGeneric   = "ReplaceMe"
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
	TablePipelineConfigs,
	TableAccounts,
	TableCheckInInfos,
}

// ErrNotImplemented is returned by service stubs before business logic is migrated.
var ErrNotImplemented = errors.New("not_implemented")

// ErrNoSession means the Wavelet user has no TraceInt cookie yet.
var ErrNoSession = errors.New("session_required")

// CodedError is a service-layer error mapped to the wavelet envelope.
type CodedError struct {
	Status int
	Code   string
	Msg    string
	Err    error
}

func (e *CodedError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *CodedError) Unwrap() error { return e.Err }

// NewError builds a CodedError.
func NewError(status int, code, msg string) *CodedError {
	return &CodedError{Status: status, Code: code, Msg: msg}
}

// NeedAuthError tells the HTTP layer to return 409 with need_auth payload.
type NeedAuthError struct {
	Kind      string
	AccountID uint64
	AuthURL   string
	Msg       string
}

// Error implements the error interface.
func (e *NeedAuthError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

// SupportedTaskKinds lists task kinds accepted by /tasks/:kind/*.
var SupportedTaskKinds = map[string]struct{}{
	TaskKindGrab:       {},
	TaskKindOccupy:     {},
	TaskKindGlobalLeak: {},
	TaskKindTomorrow:   {},
}
