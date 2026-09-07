// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package user_test

import (
	"Wavelet/core/contracts"
	"Wavelet/pkg/idgen"
	"Wavelet/plugins/domain/user"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type taskTestDB struct{ db *gorm.DB }

func (m *taskTestDB) GORM() *gorm.DB                  { return m.db }
func (m *taskTestDB) DB(ctx context.Context) *gorm.DB { return m.db.WithContext(ctx) }
func (m *taskTestDB) Named(_ string) *gorm.DB         { return m.db }

type sysConfigRow struct {
	Key   string `gorm:"primaryKey;size:64"`
	Value string `gorm:"type:text"`
}

func (sysConfigRow) TableName() string { return "w_system_configs" }

func setupUserTaskDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = idgen.Init(1)
	testDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "user_task.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(&user.User{}, &user.AccessToken{}, &sysConfigRow{}))
	user.SetDBService(&taskTestDB{db: testDB})
	t.Cleanup(func() { user.SetDBService(nil) })
	return testDB
}

func TestSendEmailCodeValidatePayload(t *testing.T) {
	h := &user.SendEmailCodeHandler{}
	_, err := h.ValidatePayload([]byte(`{"email":"not-an-email"}`))
	require.Error(t, err)

	out, err := h.ValidatePayload([]byte(`{"email":"User@Example.com"}`))
	require.NoError(t, err)
	assert.Contains(t, string(out), `"user@example.com"`)
}

func TestSendMailValidatePayload(t *testing.T) {
	h := &user.SendMailHandler{}
	_, err := h.ValidatePayload([]byte(`{"to":"a@b.com","subject":"","body":"x"}`))
	require.Error(t, err)
	_, err = h.ValidatePayload([]byte(`{"to":"a@b.com","subject":"Hi","body":"<p>ok</p>"}`))
	require.NoError(t, err)
}

func TestSendMailRequiresSMTP(t *testing.T) {
	setupUserTaskDB(t)
	h := &user.SendMailHandler{}
	_, err := h.Execute(context.Background(), []byte(`{"to":"a@b.com","subject":"Hi","body":"<p>ok</p>"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP")
}

func TestCleanupInactiveNeverLoggedInUsers(t *testing.T) {
	db := setupUserTaskDB(t)
	old := time.Now().Add(-40 * 24 * time.Hour)
	stale := user.User{ID: 42, Username: "stale", Password: "x", IsActive: true, CreatedAt: old}
	require.NoError(t, db.Create(&stale).Error)
	require.NoError(t, db.Model(&stale).Updates(map[string]any{
		"created_at":    old,
		"last_login_at": time.Time{},
	}).Error)

	fresh := user.User{ID: 43, Username: "fresh", Password: "x", IsActive: true, LastLoginAt: time.Now()}
	require.NoError(t, db.Create(&fresh).Error)

	admin := user.User{ID: 1, Username: "admin", Password: "x", IsAdmin: true, CreatedAt: old}
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Model(&admin).Updates(map[string]any{
		"created_at":    old,
		"last_login_at": time.Time{},
	}).Error)

	h := &user.CleanupInactiveHandler{}
	res, err := h.Execute(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.Message, "1")

	_, err = user.GetUserByID(context.Background(), 42)
	assert.Error(t, err)
	_, err = user.GetUserByID(context.Background(), 43)
	require.NoError(t, err)
	_, err = user.GetUserByID(context.Background(), 1)
	require.NoError(t, err)
}

func TestSendEmailCodeMetaExported(t *testing.T) {
	assert.Equal(t, "send_email_code", user.SendEmailCodeMeta.Type)
	assert.Equal(t, "user:send_email_code", user.SendEmailCodeMeta.AsynqTask)
	_ = contracts.TaskHandler(&user.SendEmailCodeHandler{})
}

type stubTaskCacheService struct {
	contracts.CacheService
	data map[string]string
}

func (s *stubTaskCacheService) Get(_ context.Context, key string, target any) error {
	v, ok := s.data[key]
	if !ok {
		return contracts.ErrCacheMiss
	}
	if strPtr, ok := target.(*string); ok {
		*strPtr = v
		return nil
	}
	return nil
}

func (s *stubTaskCacheService) Set(_ context.Context, key string, value any, _ time.Duration) error {
	if s.data == nil {
		s.data = make(map[string]string)
	}
	if str, ok := value.(string); ok {
		s.data[key] = str
	}
	return nil
}

func TestSendEmailCodeExecute_CacheBehavior(t *testing.T) {
	setupUserTaskDB(t)

	// Case 1: Cache unavailable
	h := &user.SendEmailCodeHandler{}
	_, err := h.Execute(context.Background(), []byte(`{"email":"test@example.com"}`))
	require.Error(t, err)

	// Case 2: Cache available, with existing code (idempotent reuse)
	cache := &stubTaskCacheService{
		data: map[string]string{
			"user:email_code:test@example.com": "999888",
		},
	}
	user.SetCacheService(cache)
	defer user.SetCacheService(nil)

	// When executing without SMTP config, it will fail at loadSMTPConfig, but after resolveAndCacheEmailCode
	_, err = h.Execute(context.Background(), []byte(`{"email":"test@example.com"}`))
	require.Error(t, err)
	// Cache should preserve existing code
	assert.Equal(t, "999888", cache.data["user:email_code:test@example.com"])

	// Case 3: Cache available, code empty -> auto generated
	cache2 := &stubTaskCacheService{data: make(map[string]string)}
	user.SetCacheService(cache2)
	_, _ = h.Execute(context.Background(), []byte(`{"email":"auto@example.com"}`))
	assert.Len(t, cache2.data["user:email_code:auto@example.com"], 6)

	// Case 4: Cache available, code explicitly passed
	cache3 := &stubTaskCacheService{data: make(map[string]string)}
	user.SetCacheService(cache3)
	_, _ = h.Execute(context.Background(), []byte(`{"email":"explicit@example.com","code":"123456"}`))
	assert.Equal(t, "123456", cache3.data["user:email_code:explicit@example.com"])
}
