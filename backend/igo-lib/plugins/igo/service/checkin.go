// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

// GetCheckInSession returns the remote-check-in session.
func (s *Service) GetCheckInSession(ctx context.Context, userID uint64) (*do.CheckInSessionResponse, error) {
	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toCheckInSession(row), nil
}

// GetCheckInAuthQRCode returns the independent WeChat QR entry for check-in.
func (s *Service) GetCheckInAuthQRCode(ctx context.Context, userID uint64) (*do.QRCodeResponse, error) {
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &do.QRCodeResponse{
		ImageDataURL: defaultQRCodeDataURL(),
		AuthURL:      tpl.RemoteCheckInAuthorizationReturnURL,
	}, nil
}

// AuthorizeCheckInFromCode exchanges a WeChat code for a check-in session.
func (s *Service) AuthorizeCheckInFromCode(ctx context.Context, userID uint64, req do.CheckInAuthFromCodeRequest) (*do.CheckInAuthorizationResponse, error) {
	code, ok := traceint.ExtractCode(req.Code)
	if !ok {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "签到授权链接中未找到 32 位 code")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	token, exp, err := s.api(ctx, userID).ExchangeCheckInCode(ctx, tpl, code)
	if err != nil {
		return nil, wrapTrace(err)
	}
	now := time.Now().UTC()
	row := &entity.CheckInSession{
		UserID:         userID,
		Token:          token,
		SavedAt:        now,
		ExpiresAt:      exp,
		CanAutoRestore: req.Remember,
	}
	if err := dao.UpsertCheckInSession(ctx, row); err != nil {
		return nil, err
	}
	device, devErr := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, token)
	out := &do.CheckInAuthorizationResponse{Session: *toCheckInSession(row), Device: device}
	if devErr != nil {
		out.DeviceRefreshWarning = devErr.Error()
	}
	return out, nil
}

// GetCheckInDevices returns profile and beacon UUIDs.
func (s *Service) GetCheckInDevices(ctx context.Context, userID uint64) (*do.CheckInDeviceResponse, error) {
	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeSessionRequired, "请先完成签到授权")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	info, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, row.Token)
	return info, wrapTrace(err)
}

// SignCheckIn submits a Bluetooth beacon check-in.
func (s *Service) SignCheckIn(ctx context.Context, userID uint64, req do.CheckInSignRequest) (*do.CheckInSignResponse, error) {
	if req.ExpectedLibraryID <= 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "场馆编号无效")
	}
	normUUID, ok := traceint.NormalizeUUID(req.BeaconUUID)
	if !ok {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请填写有效的 32/36 位 Beacon UUID")
	}
	req.BeaconUUID = normUUID
	if req.Major < 0 || req.Major > 65535 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Major 必须介于 0 和 65535 之间")
	}
	if req.Minor < 0 || req.Minor > 65535 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Minor 必须介于 0 和 65535 之间")
	}
	if req.Latitude < -90.0 || req.Latitude > 90.0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "纬度必须介于 -90 和 90 之间")
	}
	if req.Longitude < -180.0 || req.Longitude > 180.0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "经度必须介于 -180 和 180 之间")
	}

	row, err := dao.GetCheckInSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeSessionRequired, "请先完成签到授权")
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	ts, err := s.api(ctx, userID).GetCheckInServerTime(ctx, tpl)
	if err != nil {
		return nil, wrapTrace(err)
	}
	res, err := s.api(ctx, userID).SignCheckIn(ctx, tpl, row.Token, req, ts)
	if err != nil {
		return nil, wrapTrace(err)
	}
	// Auto-persist successfully used profile for this venue
	_, _ = s.saveProfileInternal(ctx, userID, req.ExpectedLibraryID, req.ExpectedLibraryName, req.BeaconUUID, req.Major, req.Minor, req.Latitude, req.Longitude)
	return res, nil
}

// GetCheckInVenueProfile returns saved check-in profile for a library.
func (s *Service) GetCheckInVenueProfile(ctx context.Context, userID uint64, libraryID int) (*do.CheckInVenueProfile, error) {
	if libraryID <= 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "场馆编号无效")
	}
	st := s.settings(ctx, userID)
	if st.CheckInProfiles == nil {
		return nil, nil
	}
	profile, ok := st.CheckInProfiles[libraryID]
	if !ok {
		return nil, nil
	}
	return &profile, nil
}

// SaveCheckInVenueProfile saves or updates check-in profile for a library.
func (s *Service) SaveCheckInVenueProfile(ctx context.Context, userID uint64, libraryID int, req do.SaveCheckInVenueProfileRequest) (*do.CheckInVenueProfile, error) {
	return s.saveProfileInternal(ctx, userID, libraryID, req.LibraryName, req.BeaconUUID, req.Major, req.Minor, req.Latitude, req.Longitude)
}

// ListCheckInVenueProfiles returns all saved venue profiles for a user.
func (s *Service) ListCheckInVenueProfiles(ctx context.Context, userID uint64) (*do.CheckInVenueProfilesResponse, error) {
	st := s.settings(ctx, userID)
	var list []do.CheckInVenueProfile
	for _, p := range st.CheckInProfiles {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].LibraryID < list[j].LibraryID
	})
	if list == nil {
		list = []do.CheckInVenueProfile{}
	}
	return &do.CheckInVenueProfilesResponse{Profiles: list}, nil
}

func (s *Service) saveProfileInternal(ctx context.Context, userID uint64, libraryID int, libraryName, beaconUUID string, major, minor int, lat, lng float64) (*do.CheckInVenueProfile, error) {
	if libraryID <= 0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "场馆编号无效")
	}
	normUUID, ok := traceint.NormalizeUUID(beaconUUID)
	if !ok {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "请填写有效的 32/36 位 Beacon UUID")
	}
	if major < 0 || major > 65535 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Major 必须介于 0 和 65535 之间")
	}
	if minor < 0 || minor > 65535 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "Minor 必须介于 0 和 65535 之间")
	}
	if lat < -90.0 || lat > 90.0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "纬度必须介于 -90 和 90 之间")
	}
	if lng < -180.0 || lng > 180.0 {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "经度必须介于 -180 和 180 之间")
	}

	st := s.settings(ctx, userID)
	if st.CheckInProfiles == nil {
		st.CheckInProfiles = make(map[int]do.CheckInVenueProfile)
	}
	name := strings.TrimSpace(libraryName)
	if name == "" && st.CheckInProfiles[libraryID].LibraryName != "" {
		name = st.CheckInProfiles[libraryID].LibraryName
	}
	profile := do.CheckInVenueProfile{
		LibraryID:   libraryID,
		LibraryName: name,
		BeaconUUID:  normUUID,
		Major:       major,
		Minor:       minor,
		Latitude:    lat,
		Longitude:   lng,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	st.CheckInProfiles[libraryID] = profile

	raw, err := json.Marshal(st)
	if err != nil {
		return nil, err
	}
	if err := dao.UpsertSettings(ctx, &entity.Settings{UserID: userID, Payload: string(raw)}); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ClearCheckInSession drops the remote-check-in session.
func (s *Service) ClearCheckInSession(ctx context.Context, userID uint64) error {
	return dao.DeleteCheckInSession(ctx, userID)
}

func toCheckInSession(row *entity.CheckInSession) *do.CheckInSessionResponse {
	if row == nil {
		return &do.CheckInSessionResponse{Authorized: false}
	}
	out := &do.CheckInSessionResponse{
		Authorized:     true,
		SavedAt:        row.SavedAt.UTC().Format(time.RFC3339),
		CanAutoRestore: row.CanAutoRestore,
	}
	if row.ExpiresAt != nil {
		out.ExpiresAt = row.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}
