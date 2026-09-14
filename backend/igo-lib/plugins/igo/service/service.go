// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package service implements IGoLibrary use cases. Stage 1 methods return ErrNotImplemented.
package service

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"context"
)

// Service is the igo business facade. Stage 1 is a contract-only stub.
type Service struct{}

// New creates a stub Service.
func New() *Service {
	return &Service{}
}

func stub[T any]() (T, error) {
	var zero T
	return zero, consts.ErrNotImplemented
}

func stubErr() error { return consts.ErrNotImplemented }

// GetDashboard returns the home snapshot.
func (s *Service) GetDashboard(ctx context.Context, userID uint64) (*do.DashboardResponse, error) {
	return stub[*do.DashboardResponse]()
}

// GetStatus returns the compact status payload (original /api/status).
func (s *Service) GetStatus(ctx context.Context, userID uint64) (*do.DashboardResponse, error) {
	return stub[*do.DashboardResponse]()
}

// ListActivityLogs returns recent activity entries.
func (s *Service) ListActivityLogs(ctx context.Context, userID uint64, page, perPage int) ([]do.ActivityLogEntry, int64, error) {
	return nil, 0, consts.ErrNotImplemented
}

// GetSession returns the TraceInt cookie session.
func (s *Service) GetSession(ctx context.Context, userID uint64) (*do.SessionResponse, error) {
	return stub[*do.SessionResponse]()
}

// GetAuthQRCode returns the WeChat authorization QR.
func (s *Service) GetAuthQRCode(ctx context.Context, userID uint64) (*do.QRCodeResponse, error) {
	return stub[*do.QRCodeResponse]()
}

// AuthenticateFromCode exchanges a WeChat code/link for a cookie session.
func (s *Service) AuthenticateFromCode(ctx context.Context, userID uint64, req do.AuthFromCodeRequest) (*do.SessionWorkflowResponse, error) {
	return stub[*do.SessionWorkflowResponse]()
}

// AuthenticateFromCookie stores and validates a raw cookie.
func (s *Service) AuthenticateFromCookie(ctx context.Context, userID uint64, req do.AuthFromCookieRequest) (*do.SessionWorkflowResponse, error) {
	return stub[*do.SessionWorkflowResponse]()
}

// RefreshCookie refreshes the current TraceInt cookie.
func (s *Service) RefreshCookie(ctx context.Context, userID uint64) (*do.SessionWorkflowResponse, error) {
	return stub[*do.SessionWorkflowResponse]()
}

// SignOut clears the TraceInt session.
func (s *Service) SignOut(ctx context.Context, userID uint64) error {
	return stubErr()
}

// ListLibraries loads venues for the authorized account.
func (s *Service) ListLibraries(ctx context.Context, userID uint64) ([]do.LibrarySummary, error) {
	return stub[[]do.LibrarySummary]()
}

// GetBoundLibrary returns the locked venue.
func (s *Service) GetBoundLibrary(ctx context.Context, userID uint64) (*do.BoundLibraryResponse, error) {
	return stub[*do.BoundLibraryResponse]()
}

// RefreshBoundLibrary reloads the locked venue layout.
func (s *Service) RefreshBoundLibrary(ctx context.Context, userID uint64) (*do.BoundLibraryResponse, error) {
	return stub[*do.BoundLibraryResponse]()
}

// GetLibrary returns one venue summary.
func (s *Service) GetLibrary(ctx context.Context, userID uint64, libraryID int) (*do.LibrarySummary, error) {
	return stub[*do.LibrarySummary]()
}

// GetLibraryLayout returns the seat map.
func (s *Service) GetLibraryLayout(ctx context.Context, userID uint64, libraryID int) (*do.LibraryLayoutResponse, error) {
	return stub[*do.LibraryLayoutResponse]()
}

// GetLibraryRule returns opening/booking rules.
func (s *Service) GetLibraryRule(ctx context.Context, userID uint64, libraryID int) (*do.LibraryRuleResponse, error) {
	return stub[*do.LibraryRuleResponse]()
}

// BindLibrary locks a venue.
func (s *Service) BindLibrary(ctx context.Context, userID uint64, libraryID int) (*do.BoundLibraryResponse, error) {
	return stub[*do.BoundLibraryResponse]()
}

// PreviewLibrary loads a venue without locking it.
func (s *Service) PreviewLibrary(ctx context.Context, userID uint64, libraryID int) (*do.LibraryLayoutResponse, error) {
	return stub[*do.LibraryLayoutResponse]()
}

// GetFavorites returns favorite seats for a venue.
func (s *Service) GetFavorites(ctx context.Context, userID uint64, libraryID int) ([]do.SeatRef, error) {
	return stub[[]do.SeatRef]()
}

// SaveFavorites replaces favorite seats for a venue.
func (s *Service) SaveFavorites(ctx context.Context, userID uint64, libraryID int, seats []do.SeatRef) error {
	return stubErr()
}

// SetSeatLabels writes labels onto selected seats.
func (s *Service) SetSeatLabels(ctx context.Context, userID uint64, libraryID int, req do.SetSeatLabelsRequest) ([]do.SeatLabel, error) {
	return stub[[]do.SeatLabel]()
}

// DeleteSeatLabels removes labels by seat key.
func (s *Service) DeleteSeatLabels(ctx context.Context, userID uint64, libraryID int, seatKeys []string) error {
	return stubErr()
}

// GetReservation returns the current reservation.
func (s *Service) GetReservation(ctx context.Context, userID uint64) (*do.ReservationResponse, error) {
	return stub[*do.ReservationResponse]()
}

// RefreshReservation reloads reservation info from TraceInt.
func (s *Service) RefreshReservation(ctx context.Context, userID uint64) (*do.ReservationOperationResponse, error) {
	return stub[*do.ReservationOperationResponse]()
}

// CancelReservation cancels the current reservation.
func (s *Service) CancelReservation(ctx context.Context, userID uint64, req do.CancelReservationRequest) (*do.ReservationOperationResponse, error) {
	return stub[*do.ReservationOperationResponse]()
}

// ListTasks returns coordinator snapshots.
func (s *Service) ListTasks(ctx context.Context, userID uint64) (*do.TaskListResponse, error) {
	return stub[*do.TaskListResponse]()
}

// ListTaskRecords returns recent launch history.
func (s *Service) ListTaskRecords(ctx context.Context, userID uint64) ([]do.TaskLaunchRecord, error) {
	return stub[[]do.TaskLaunchRecord]()
}

// StartGrab starts the grab coordinator.
func (s *Service) StartGrab(ctx context.Context, userID uint64, req do.GrabStartRequest) error {
	return stubErr()
}

// StartOccupy starts the occupy coordinator.
func (s *Service) StartOccupy(ctx context.Context, userID uint64, req do.OccupyStartRequest) error {
	return stubErr()
}

// StartGlobalLeak starts the global-leak coordinator.
func (s *Service) StartGlobalLeak(ctx context.Context, userID uint64, req do.GlobalLeakStartRequest) error {
	return stubErr()
}

// StartTomorrow starts the tomorrow-reservation coordinator.
func (s *Service) StartTomorrow(ctx context.Context, userID uint64, req do.TomorrowStartRequest) error {
	return stubErr()
}

// CancelTask stops a coordinator by kind.
func (s *Service) CancelTask(ctx context.Context, userID uint64, kind string) error {
	return stubErr()
}

// RunTomorrowNow executes one tomorrow-reservation attempt immediately.
func (s *Service) RunTomorrowNow(ctx context.Context, userID uint64, req do.TomorrowStartRequest) error {
	return stubErr()
}

// GetGlobalLeakBlacklist returns blacklisted seats keyed by library id.
func (s *Service) GetGlobalLeakBlacklist(ctx context.Context, userID uint64) (*do.GlobalLeakBlacklistResponse, error) {
	return stub[*do.GlobalLeakBlacklistResponse]()
}

// SaveGlobalLeakBlacklist replaces submitted venue blacklists.
func (s *Service) SaveGlobalLeakBlacklist(ctx context.Context, userID uint64, req do.SaveGlobalLeakBlacklistRequest) error {
	return stubErr()
}

// GetGlobalLeakSelectedLibraries returns scan-priority venues.
func (s *Service) GetGlobalLeakSelectedLibraries(ctx context.Context, userID uint64) ([]do.GlobalLeakLibraryTarget, error) {
	return stub[[]do.GlobalLeakLibraryTarget]()
}

// SaveGlobalLeakSelectedLibraries persists scan-priority venues.
func (s *Service) SaveGlobalLeakSelectedLibraries(ctx context.Context, userID uint64, req do.SaveGlobalLeakSelectedLibrariesRequest) error {
	return stubErr()
}

// GetCheckInSession returns the remote-check-in session.
func (s *Service) GetCheckInSession(ctx context.Context, userID uint64) (*do.CheckInSessionResponse, error) {
	return stub[*do.CheckInSessionResponse]()
}

// GetCheckInAuthQRCode returns the independent WeChat QR for check-in.
func (s *Service) GetCheckInAuthQRCode(ctx context.Context, userID uint64) (*do.QRCodeResponse, error) {
	return stub[*do.QRCodeResponse]()
}

// AuthorizeCheckInFromCode exchanges a WeChat code for a check-in session.
func (s *Service) AuthorizeCheckInFromCode(ctx context.Context, userID uint64, req do.CheckInAuthFromCodeRequest) (*do.CheckInAuthorizationResponse, error) {
	return stub[*do.CheckInAuthorizationResponse]()
}

// GetCheckInDevices returns profile and beacon UUIDs.
func (s *Service) GetCheckInDevices(ctx context.Context, userID uint64) (*do.CheckInDeviceResponse, error) {
	return stub[*do.CheckInDeviceResponse]()
}

// SignCheckIn submits a Bluetooth beacon check-in.
func (s *Service) SignCheckIn(ctx context.Context, userID uint64, req do.CheckInSignRequest) (*do.CheckInSignResponse, error) {
	return stub[*do.CheckInSignResponse]()
}

// ClearCheckInSession drops the remote-check-in session.
func (s *Service) ClearCheckInSession(ctx context.Context, userID uint64) error {
	return stubErr()
}

// GetProtocolTemplates returns effective (default + override) templates.
func (s *Service) GetProtocolTemplates(ctx context.Context, userID uint64) (*do.ProtocolTemplatesResponse, error) {
	return stub[*do.ProtocolTemplatesResponse]()
}

// GetDefaultProtocolTemplates returns built-in templates.
func (s *Service) GetDefaultProtocolTemplates(ctx context.Context, userID uint64) (*do.ProtocolTemplatesResponse, error) {
	return stub[*do.ProtocolTemplatesResponse]()
}

// SaveProtocolTemplates stores per-user overrides.
func (s *Service) SaveProtocolTemplates(ctx context.Context, userID uint64, req do.SaveProtocolTemplatesRequest) (*do.ProtocolTemplatesResponse, error) {
	return stub[*do.ProtocolTemplatesResponse]()
}

// ResetProtocolTemplates clears overrides.
func (s *Service) ResetProtocolTemplates(ctx context.Context, userID uint64) (*do.ProtocolTemplatesResponse, error) {
	return stub[*do.ProtocolTemplatesResponse]()
}

// GetSettings returns migratable settings.
func (s *Service) GetSettings(ctx context.Context, userID uint64) (*do.SettingsResponse, error) {
	return stub[*do.SettingsResponse]()
}

// SaveSettings updates migratable settings.
func (s *Service) SaveSettings(ctx context.Context, userID uint64, req do.SaveSettingsRequest) (*do.SettingsResponse, error) {
	return stub[*do.SettingsResponse]()
}

// ExportBackup produces an encrypted backup blob.
func (s *Service) ExportBackup(ctx context.Context, userID uint64, req do.BackupExportRequest) (*do.BackupExportResponse, error) {
	return stub[*do.BackupExportResponse]()
}

// ImportBackup restores from an encrypted backup blob.
func (s *Service) ImportBackup(ctx context.Context, userID uint64, req do.BackupImportRequest) error {
	return stubErr()
}

// GetWebDAV returns WebDAV settings with password redacted.
func (s *Service) GetWebDAV(ctx context.Context, userID uint64) (*do.WebDAVSettings, error) {
	return stub[*do.WebDAVSettings]()
}

// SaveWebDAV writes WebDAV settings.
func (s *Service) SaveWebDAV(ctx context.Context, userID uint64, req do.SaveWebDAVRequest) (*do.WebDAVSettings, error) {
	return stub[*do.WebDAVSettings]()
}

// SyncWebDAV runs a manual remote sync.
func (s *Service) SyncWebDAV(ctx context.Context, userID uint64) (*do.WebDAVSyncResponse, error) {
	return stub[*do.WebDAVSyncResponse]()
}
