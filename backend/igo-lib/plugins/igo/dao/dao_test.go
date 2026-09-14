// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package dao_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/pkg/idgen"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testDB struct{ db *gorm.DB }

func (s testDB) GORM() *gorm.DB                  { return s.db }
func (s testDB) DB(ctx context.Context) *gorm.DB { return s.db.WithContext(ctx) }
func (s testDB) Named(string) *gorm.DB           { return s.db }

func TestMain(m *testing.M) {
	if err := idgen.Init(1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func openMigratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "igo.db")), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := gdb.DB()
	require.NoError(t, err)
	sub, err := fs.Sub(igo.MigrationsFS, "migrations/sqlite")
	require.NoError(t, err)

	provider, err := goose.NewProvider(goose.DialectSQLite3, sqlDB, sub)
	require.NoError(t, err)
	_, err = provider.Up(context.Background())
	require.NoError(t, err)

	dao.SetDBService(testDB{db: gdb})
	t.Cleanup(func() { dao.SetDBService((contracts.DBService)(nil)) })
	return gdb
}

func TestSQLFilesCoverOwnedTables(t *testing.T) {
	roots := []string{
		filepath.Join("..", "migrations", "postgres", "00001_initial.sql"),
		filepath.Join("..", "migrations", "sqlite", "00001_initial.sql"),
	}
	for _, path := range roots {
		raw, err := os.ReadFile(path)
		require.NoError(t, err, path)
		body := string(raw)
		for _, table := range consts.OwnedTables {
			assert.Contains(t, body, "CREATE TABLE IF NOT EXISTS "+table, path)
			assert.Contains(t, body, "DROP TABLE IF EXISTS "+table, path)
		}
		assert.NotContains(t, strings.ToUpper(body), "FOREIGN KEY")
		assert.NotContains(t, strings.ToUpper(body), "REFERENCES ")
	}
}

func TestMigrateCreatesOwnedTables(t *testing.T) {
	gdb := openMigratedDB(t)
	for _, table := range consts.OwnedTables {
		var n int
		err := gdb.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&n).Error
		require.NoError(t, err)
		assert.Equal(t, 1, n, "table %s missing after migrate", table)
	}
}

func TestSessionIsolatedByUser(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	require.NoError(t, dao.UpsertSession(ctx, &entity.Session{
		UserID: 1, Cookie: "cookie-a", Source: "code", SavedAt: now, CanAutoRestore: true,
	}))
	require.NoError(t, dao.UpsertSession(ctx, &entity.Session{
		UserID: 2, Cookie: "cookie-b", Source: "cookie", SavedAt: now, CanAutoRestore: true,
	}))

	a, err := dao.GetSession(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, a)
	assert.Equal(t, "cookie-a", a.Cookie)

	b, err := dao.GetSession(ctx, 2)
	require.NoError(t, err)
	require.NotNil(t, b)
	assert.Equal(t, "cookie-b", b.Cookie)

	require.NoError(t, dao.UpsertSession(ctx, &entity.Session{
		UserID: 1, Cookie: "cookie-a-refresh", Source: "refresh", SavedAt: now, CanAutoRestore: true,
	}))
	a, err = dao.GetSession(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "cookie-a-refresh", a.Cookie)
	b, err = dao.GetSession(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, "cookie-b", b.Cookie)

	require.NoError(t, dao.DeleteSession(ctx, 1))
	gone, err := dao.GetSession(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, gone)
	still, err := dao.GetSession(ctx, 2)
	require.NoError(t, err)
	require.NotNil(t, still)
}

func TestFavoritesReplaceDoesNotCrossUserOrLibrary(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()

	require.NoError(t, dao.ReplaceFavorites(ctx, 1, 10, []entity.Favorite{
		{SeatKey: "A-1", SeatName: "A1"},
		{SeatKey: "A-2", SeatName: "A2"},
	}))
	require.NoError(t, dao.ReplaceFavorites(ctx, 1, 11, []entity.Favorite{
		{SeatKey: "B-1", SeatName: "B1"},
	}))
	require.NoError(t, dao.ReplaceFavorites(ctx, 2, 10, []entity.Favorite{
		{SeatKey: "C-1", SeatName: "C1"},
	}))

	require.NoError(t, dao.ReplaceFavorites(ctx, 1, 10, []entity.Favorite{
		{SeatKey: "A-9", SeatName: "A9"},
	}))

	u1l10, err := dao.ListFavorites(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, u1l10, 1)
	assert.Equal(t, "A-9", u1l10[0].SeatKey)

	u1l11, err := dao.ListFavorites(ctx, 1, 11)
	require.NoError(t, err)
	require.Len(t, u1l11, 1)
	assert.Equal(t, "B-1", u1l11[0].SeatKey)

	u2l10, err := dao.ListFavorites(ctx, 2, 10)
	require.NoError(t, err)
	require.Len(t, u2l10, 1)
	assert.Equal(t, "C-1", u2l10[0].SeatKey)
}

func TestTaskLaunchHistoryFingerprintUniquePerUser(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()
	now := time.Now().UTC()

	require.NoError(t, dao.UpsertTaskLaunchHistory(ctx, &entity.TaskLaunchHistory{
		UserID: 1, RecordID: "r1", Kind: consts.TaskKindGrab, Fingerprint: "fp",
		RecordedAt: now, PayloadJSON: `{"seats":1}`,
	}))
	require.NoError(t, dao.UpsertTaskLaunchHistory(ctx, &entity.TaskLaunchHistory{
		UserID: 2, RecordID: "r2", Kind: consts.TaskKindGrab, Fingerprint: "fp",
		RecordedAt: now, PayloadJSON: `{"seats":2}`,
	}))
	require.NoError(t, dao.UpsertTaskLaunchHistory(ctx, &entity.TaskLaunchHistory{
		UserID: 1, RecordID: "r1b", Kind: consts.TaskKindGrab, Fingerprint: "fp",
		RecordedAt: now, PayloadJSON: `{"seats":9}`,
	}))

	u1, err := dao.ListTaskLaunchHistory(ctx, 1, consts.TaskKindGrab, 10)
	require.NoError(t, err)
	require.Len(t, u1, 1)
	assert.Equal(t, "r1b", u1[0].RecordID)
	assert.Equal(t, `{"seats":9}`, u1[0].PayloadJSON)

	u2, err := dao.ListTaskLaunchHistory(ctx, 2, consts.TaskKindGrab, 10)
	require.NoError(t, err)
	require.Len(t, u2, 1)
	assert.Equal(t, "r2", u2[0].RecordID)
}

func TestTaskLaunchHistoryPrunesToFive(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for i := 0; i < 7; i++ {
		require.NoError(t, dao.UpsertTaskLaunchHistory(ctx, &entity.TaskLaunchHistory{
			UserID: 3, RecordID: "r" + string(rune('a'+i)), Kind: consts.TaskKindGrab,
			Fingerprint: "fp" + string(rune('a'+i)), RecordedAt: now, PayloadJSON: `{}`,
		}))
	}
	rows, err := dao.ListTaskLaunchHistory(ctx, 3, consts.TaskKindGrab, 20)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(rows), 5)
}

func TestVenueAndSettingsUpsertByUser(t *testing.T) {
	_ = openMigratedDB(t)
	ctx := context.Background()

	require.NoError(t, dao.UpsertVenue(ctx, &entity.Venue{UserID: 7, LibraryID: 3, Name: "三楼"}))
	require.NoError(t, dao.UpsertVenue(ctx, &entity.Venue{UserID: 7, LibraryID: 4, Name: "四楼"}))
	row, err := dao.GetVenue(ctx, 7)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, 4, row.LibraryID)
	assert.Equal(t, "四楼", row.Name)

	require.NoError(t, dao.UpsertSettings(ctx, &entity.Settings{UserID: 7, Payload: `{"network_max_retries":3}`}))
	s, err := dao.GetSettings(ctx, 7)
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Contains(t, s.Payload, "network_max_retries")
}

func TestDBNotReady(t *testing.T) {
	dao.SetDBService((contracts.DBService)(nil))
	_, err := dao.GetSession(context.Background(), 1)
	require.ErrorIs(t, err, dao.ErrDBNotReady)
}
