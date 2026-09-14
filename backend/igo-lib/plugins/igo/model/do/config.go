// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// ProtocolTemplatesResponse is the TraceInt URL/GraphQL template set.
type ProtocolTemplatesResponse struct {
	GetCookieURLTemplate                string `json:"get_cookie_url_template"`
	CookieAuthorizationReturnURL        string `json:"cookie_authorization_return_url"`
	GraphQLEndpointURL                  string `json:"graphql_endpoint_url"`
	GraphQLDefaultRefererURL            string `json:"graphql_default_referer_url"`
	GraphQLDefaultOriginURL             string `json:"graphql_default_origin_url"`
	GraphQLTomorrowRefererURL           string `json:"graphql_tomorrow_referer_url"`
	GraphQLTomorrowOriginURL            string `json:"graphql_tomorrow_origin_url"`
	TomorrowReservationQueueURLTemplate string `json:"tomorrow_reservation_queue_url_template"`
	RemoteCheckInAuthURLTemplate        string `json:"remote_checkin_auth_url_template"`
	RemoteCheckInAuthorizationReturnURL string `json:"remote_checkin_authorization_return_url"`
	RemoteCheckInAuthRefererURL         string `json:"remote_checkin_auth_referer_url"`
	RemoteCheckInDevicesEndpointURL     string `json:"remote_checkin_devices_endpoint_url"`
	RemoteCheckInTimeEndpointURL        string `json:"remote_checkin_time_endpoint_url"`
	RemoteCheckInSignEndpointURL        string `json:"remote_checkin_sign_endpoint_url"`
	RemoteCheckInAPIRefererURL          string `json:"remote_checkin_api_referer_url"`
	QueryLibrariesTemplate              string `json:"query_libraries_template"`
	QueryLibraryLayoutTemplate          string `json:"query_library_layout_template"`
	QueryLibraryRuleTemplate            string `json:"query_library_rule_template"`
	QueryReservationInfoTemplate        string `json:"query_reservation_info_template"`
	ReserveSeatTemplate                 string `json:"reserve_seat_template"`
	CancelReservationTemplate           string `json:"cancel_reservation_template"`
	TomorrowReservationWarmUpTemplate   string `json:"tomorrow_reservation_warmup_template"`
	TomorrowReservationSaveTemplate     string `json:"tomorrow_reservation_save_template"`
	TomorrowReservationInfoTemplate     string `json:"tomorrow_reservation_info_template"`
}

// SaveProtocolTemplatesRequest stores per-user protocol overrides.
type SaveProtocolTemplatesRequest struct {
	Overrides ProtocolTemplatesResponse `json:"overrides" binding:"required"`
}

// SettingsResponse is the migratable subset of desktop system settings.
type SettingsResponse struct {
	RequestTimeoutSeconds              int    `json:"request_timeout_seconds"`
	NetworkMaxRetries                  int    `json:"network_max_retries"`
	TraceIntGraphQLOverridesEnabled    bool   `json:"traceint_graphql_overrides_enabled"`
	GrabReservationStrategy            string `json:"grab_reservation_strategy"`
	OptimalGrabStrategyReminderEnabled bool   `json:"optimal_grab_strategy_reminder_enabled"`
	GrabScheduledStartDefault          string `json:"grab_scheduled_start_default,omitempty"`
	TomorrowScheduledStartDefault      string `json:"tomorrow_scheduled_start_default,omitempty"`
	OccupyReReserveDelaySeconds        int    `json:"occupy_re_reserve_delay_seconds"`
	OccupyCheckIntervalMode            string `json:"occupy_check_interval_mode"`
	GlobalLeakScanIntervalSeconds      int    `json:"global_leak_scan_interval_seconds"`
	AutoReleaseEnabled                 bool   `json:"auto_release_enabled"`
	AutoReleaseLeadSeconds             int    `json:"auto_release_lead_seconds"`
	HomeReservationProgressMode        string `json:"home_reservation_progress_mode,omitempty"`
}

// SaveSettingsRequest updates migratable settings. All fields optional (PATCH-like PUT).
type SaveSettingsRequest struct {
	RequestTimeoutSeconds              *int    `json:"request_timeout_seconds"`
	NetworkMaxRetries                  *int    `json:"network_max_retries"`
	TraceIntGraphQLOverridesEnabled    *bool   `json:"traceint_graphql_overrides_enabled"`
	GrabReservationStrategy            *string `json:"grab_reservation_strategy"`
	OptimalGrabStrategyReminderEnabled *bool   `json:"optimal_grab_strategy_reminder_enabled"`
	GrabScheduledStartDefault          *string `json:"grab_scheduled_start_default"`
	TomorrowScheduledStartDefault      *string `json:"tomorrow_scheduled_start_default"`
	OccupyReReserveDelaySeconds        *int    `json:"occupy_re_reserve_delay_seconds"`
	OccupyCheckIntervalMode            *string `json:"occupy_check_interval_mode"`
	GlobalLeakScanIntervalSeconds      *int    `json:"global_leak_scan_interval_seconds"`
	AutoReleaseEnabled                 *bool   `json:"auto_release_enabled"`
	AutoReleaseLeadSeconds             *int    `json:"auto_release_lead_seconds"`
	HomeReservationProgressMode        *string `json:"home_reservation_progress_mode"`
}

// BackupExportRequest encrypts and returns a backup blob.
type BackupExportRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// BackupExportResponse is the exported archive (base64).
type BackupExportResponse struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// BackupImportRequest restores from an encrypted archive.
type BackupImportRequest struct {
	Password string `json:"password" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// WebDAVSettings is the remote sync endpoint (password never echoed back).
type WebDAVSettings struct {
	Endpoint        string `json:"endpoint"`
	RemoteDirectory string `json:"remote_directory"`
	Username        string `json:"username"`
	PasswordSet     bool   `json:"password_set"`
	TLSVerifyMode   string `json:"tls_verify_mode"`
}

// SaveWebDAVRequest writes WebDAV sync settings.
type SaveWebDAVRequest struct {
	Endpoint        string `json:"endpoint" binding:"required"`
	RemoteDirectory string `json:"remote_directory"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	TLSVerifyMode   string `json:"tls_verify_mode"`
}

// WebDAVSyncResponse is the result of a manual sync.
type WebDAVSyncResponse struct {
	Message string `json:"message"`
}
