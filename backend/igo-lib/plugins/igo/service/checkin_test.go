// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"net/http"
	"testing"

	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckInVenueProfileWorkflow(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	userID := uint64(42)

	// 1. Initially no profile for library 101
	p, err := svc.GetCheckInVenueProfile(ctx, userID, 101)
	require.NoError(t, err)
	assert.Nil(t, p)

	// 2. Validation failures
	_, err = svc.SaveCheckInVenueProfile(ctx, userID, 0, do.SaveCheckInVenueProfileRequest{
		BeaconUUID: "fda50693-a4e2-4fb1-afcf-c6eb07647825",
	})
	require.Error(t, err)

	_, err = svc.SaveCheckInVenueProfile(ctx, userID, 101, do.SaveCheckInVenueProfileRequest{
		BeaconUUID: "invalid-uuid",
		Major:      10001,
		Minor:      1980,
		Latitude:   31.2304,
		Longitude:  121.4737,
	})
	require.Error(t, err)

	_, err = svc.SaveCheckInVenueProfile(ctx, userID, 101, do.SaveCheckInVenueProfileRequest{
		BeaconUUID: "fda50693-a4e2-4fb1-afcf-c6eb07647825",
		Major:      70000,
	})
	require.Error(t, err)

	_, err = svc.SaveCheckInVenueProfile(ctx, userID, 101, do.SaveCheckInVenueProfileRequest{
		BeaconUUID: "fda50693-a4e2-4fb1-afcf-c6eb07647825",
		Latitude:   95.0,
	})
	require.Error(t, err)

	// 3. Save valid profile for Library 101
	saved101, err := svc.SaveCheckInVenueProfile(ctx, userID, 101, do.SaveCheckInVenueProfileRequest{
		LibraryName: "第三电子阅览室",
		BeaconUUID:  "fda50693-a4e2-4fb1-afcf-c6eb07647825",
		Major:       10001,
		Minor:       1980,
		Latitude:    31.2304,
		Longitude:   121.4737,
	})
	require.NoError(t, err)
	assert.Equal(t, 101, saved101.LibraryID)
	assert.Equal(t, "第三电子阅览室", saved101.LibraryName)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", saved101.BeaconUUID)
	assert.Equal(t, 10001, saved101.Major)
	assert.Equal(t, 1980, saved101.Minor)
	assert.Equal(t, 31.2304, saved101.Latitude)
	assert.Equal(t, 121.4737, saved101.Longitude)

	// 4. Save valid profile for Library 102
	saved102, err := svc.SaveCheckInVenueProfile(ctx, userID, 102, do.SaveCheckInVenueProfileRequest{
		LibraryName: "自修二室",
		BeaconUUID:  "e2c56db5-dffb-48d2-b060-d0f5a71096e0",
		Major:       20002,
		Minor:       2980,
		Latitude:    39.9042,
		Longitude:   116.4074,
	})
	require.NoError(t, err)
	assert.Equal(t, 102, saved102.LibraryID)
	assert.Equal(t, "E2C56DB5-DFFB-48D2-B060-D0F5A71096E0", saved102.BeaconUUID)

	// 5. Get individually
	p101, err := svc.GetCheckInVenueProfile(ctx, userID, 101)
	require.NoError(t, err)
	require.NotNil(t, p101)
	assert.Equal(t, "FDA50693-A4E2-4FB1-AFCF-C6EB07647825", p101.BeaconUUID)

	p102, err := svc.GetCheckInVenueProfile(ctx, userID, 102)
	require.NoError(t, err)
	require.NotNil(t, p102)
	assert.Equal(t, "E2C56DB5-DFFB-48D2-B060-D0F5A71096E0", p102.BeaconUUID)

	// 6. List profiles
	resp, err := svc.ListCheckInVenueProfiles(ctx, userID)
	require.NoError(t, err)
	require.Len(t, resp.Profiles, 2)
	assert.Equal(t, 101, resp.Profiles[0].LibraryID)
	assert.Equal(t, 102, resp.Profiles[1].LibraryID)
}
