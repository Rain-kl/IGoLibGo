// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package logstore

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeDBService struct {
	db *gorm.DB
}

func (f *fakeDBService) GORM() *gorm.DB                  { return f.db }
func (f *fakeDBService) DB(ctx context.Context) *gorm.DB { return f.db.WithContext(ctx) }
func (f *fakeDBService) Named(name string) *gorm.DB      { return f.db }

func newTestUserAccessStore(t *testing.T) *userAccessLogGormStore {
	t.Helper()
	_ = idgen.Init(1)
	gdb, err := gorm.Open(sqlite.Open("file:logstore-"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gdb.AutoMigrate(&UserAccessLog{}))
	return newUserAccessLogGormStore(gdb)
}

func TestGormUserAccessLogCountList(t *testing.T) {
	ua := newTestUserAccessStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, ua.BatchInsert(ctx, []UserAccessLog{
		{UserID: 10, Path: "/api/v1/users", Method: "GET", Status: 200, CreatedAt: now},
		{UserID: 20, Path: "/api/v1/admin", Method: "GET", Status: 200, CreatedAt: now},
		{UserID: 10, Path: "/api/v1/other", Method: "POST", Status: 201, CreatedAt: now},
	}))

	count, err := ua.Count(ctx, AccessLogFilter{UserIDs: []uint64{10}, Path: "users"})
	require.NoError(t, err)
	require.Equal(t, uint64(1), count)

	rows, total, err := ua.List(ctx, AccessLogFilter{UserIDs: []uint64{10}, Path: "users"}, 1, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(1), total)
	require.Len(t, rows, 1)
	require.Equal(t, "/api/v1/users", rows[0].Path)
	require.NotZero(t, rows[0].ID)
}

func TestGormUserAccessLogBrowserDistribution(t *testing.T) {
	ua := newTestUserAccessStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, ua.BatchInsert(ctx, []UserAccessLog{
		{UserID: 10, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36", CreatedAt: now},
		{UserID: 20, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/121.0.0.0 Safari/537.36", CreatedAt: now},
		{UserID: 30, UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/119.0", CreatedAt: now},
	}))

	shares, err := ua.GetBrowserDistribution(ctx, now.Add(-time.Hour))
	require.NoError(t, err)
	require.Len(t, shares, 2)
	require.Equal(t, "Chrome", shares[0].Browser)
	require.Equal(t, uint64(2), shares[0].Count)
	require.Equal(t, "Firefox", shares[1].Browser)
	require.Equal(t, uint64(1), shares[1].Count)
}

func TestDBHelperAppContextInjection(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:logstore-dbhelper?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	svc := &fakeDBService{db: gdb}
	appCtx := core.NewContext(context.Background())
	appCtx.Provide[contracts.DBService](svc)

	// 1. Direct *core.Context passed as context.Context
	db := getDB(appCtx)
	require.NotNil(t, db)

	ch := getChDB(appCtx)
	require.NotNil(t, ch)

	// 2. Standard context wrapped via core.WithAppContext
	wrappedCtx := core.WithAppContext(context.Background(), appCtx)
	dbWrapped := getDB(wrappedCtx)
	require.NotNil(t, dbWrapped)

	chWrapped := getChDB(wrappedCtx)
	require.NotNil(t, chWrapped)
}

func TestGormUserAccessLogFreeze(t *testing.T) {
	ua := newTestUserAccessStore(t)
	SetConfigReader(func(_ context.Context, key string) (string, error) {
		if key == logMigrationKey {
			return "migrating", nil
		}
		return "", nil
	})
	t.Cleanup(ResetForTest)

	err := ua.BatchInsert(context.Background(), []UserAccessLog{{UserID: 1, CreatedAt: time.Now()}})
	require.ErrorIs(t, err, ErrMigrating)
}
