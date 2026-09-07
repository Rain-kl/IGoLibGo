// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	cache "Wavelet/plugins/infra/cache"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestStorageCache(t *testing.T) {
	// 1. Reset cache
	ResetCache()

	if activeConfigJSON != "" || activeDriver != "" || activeBackend != nil || !lastChecked.IsZero() {
		t.Fatal("ResetCache did not clear cache variables")
	}

	// 2. Set up cache manually
	expectedConfig := Config{
		Driver: DriverLocal,
		Local:  LocalConfig{Root: t.TempDir()},
	}
	cfgJSON, err := json.Marshal(expectedConfig)
	if err != nil {
		t.Fatalf("Marshal config failed: %v", err)
	}

	cacheMutex.Lock()
	activeConfigJSON = string(cfgJSON)
	lastChecked = time.Now()
	cacheMutex.Unlock()

	// 3. Call LoadConfig and verify it loads from cache (doesn't hit database, which would fail/panic because DB is not initialized)
	ctx := context.Background()
	loadedCfg, err := LoadConfig(ctx)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedCfg.Driver != expectedConfig.Driver || loadedCfg.Local.Root != expectedConfig.Local.Root {
		t.Errorf("Loaded config %+v, expected %+v", loadedCfg, expectedConfig)
	}

	// 4. Test Active() returns cached driver and backend
	mockBnd := &functionBackend{
		put:    func(context.Context, string, io.Reader, int64, string) error { return nil },
		get:    func(context.Context, string) (*Object, error) { return nil, nil },
		delete: func(context.Context, string) error { return nil },
	}

	cacheMutex.Lock()
	activeBackend = mockBnd
	activeDriver = DriverLocal
	cacheMutex.Unlock()

	drv, bnd, err := Active(ctx)
	if err != nil {
		t.Fatalf("Active failed: %v", err)
	}
	if drv != DriverLocal || bnd != mockBnd {
		t.Errorf("Active returned driver %v, backend %v; expected %v, %v", drv, bnd, DriverLocal, mockBnd)
	}

	// 5. Test ResetCache again
	ResetCache()
	if activeConfigJSON != "" || activeDriver != "" || activeBackend != nil || !lastChecked.IsZero() {
		t.Fatal("ResetCache did not clear cache variables after setting them")
	}
}

func TestStorageCachePubSub(t *testing.T) {
	// 1. Start miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to run miniredis: %v", err)
	}
	defer mr.Close()

	// 2. Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	// 3. Set cache.Redis to our client
	oldRedis := cache.Redis
	cache.Redis = rdb
	defer func() {
		cache.Redis = oldRedis
	}()

	// Reset cache and set some cached config
	ResetCache()
	cacheMutex.Lock()
	activeConfigJSON = "some_config"
	lastChecked = time.Now()
	cacheMutex.Unlock()

	// 4. Force trigger lazy initialization of subscription
	// Reset the once guard so it runs the listener
	pubSubOnce = sync.Once{}
	ctx := context.Background()

	// Create mock backend for Active call
	mockBnd := &functionBackend{
		put:    func(context.Context, string, io.Reader, int64, string) error { return nil },
		get:    func(context.Context, string) (*Object, error) { return nil, nil },
		delete: func(context.Context, string) error { return nil },
	}
	cacheMutex.Lock()
	activeBackend = mockBnd
	activeDriver = DriverLocal
	cacheMutex.Unlock()

	_, _, _ = Active(ctx) // This calls startPubSubListener()

	// Allow some time for subscriber connection
	time.Sleep(100 * time.Millisecond)

	// 5. Publish cache invalidation
	PublishCacheInvalidation(ctx)

	// Allow message propagation
	time.Sleep(100 * time.Millisecond)

	// 6. Verify cache was cleared
	cacheMutex.RLock()
	configJSON := activeConfigJSON
	cacheMutex.RUnlock()

	if configJSON != "" {
		t.Error("Memory cache was not cleared after Redis Pub/Sub broadcast")
	}
}

func TestUpsertSystemConfigVisibility(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Create table with visibility INTEGER as in production PG / SQLite migrations
	err = db.Exec(`CREATE TABLE w_system_configs (
		key VARCHAR(64) PRIMARY KEY,
		value TEXT NOT NULL,
		type VARCHAR(32) NOT NULL DEFAULT 'system',
		visibility INTEGER NOT NULL DEFAULT 0,
		description VARCHAR(255),
		updated_at TIMESTAMP,
		created_at TIMESTAMP
	)`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	cfg := DefaultConfig()
	if err := upsertSystemConfig(context.Background(), db, "storage_config", cfg, "desc"); err != nil {
		t.Fatalf("upsertSystemConfig failed: %v", err)
	}

	var row struct {
		Key        string
		Visibility int
	}
	if err := db.Table("w_system_configs").Where("key = ?", "storage_config").Scan(&row).Error; err != nil {
		t.Fatalf("query system config failed: %v", err)
	}
	if row.Visibility != 0 {
		t.Fatalf("expected visibility 0, got %d", row.Visibility)
	}

	// Update existing record
	if err := upsertSystemConfig(context.Background(), db, "storage_config", cfg, "updated desc"); err != nil {
		t.Fatalf("upsertSystemConfig update failed: %v", err)
	}
	if err := db.Table("w_system_configs").Where("key = ?", "storage_config").Scan(&row).Error; err != nil {
		t.Fatalf("query updated system config failed: %v", err)
	}
	if row.Visibility != 0 {
		t.Fatalf("expected visibility 0 after update, got %d", row.Visibility)
	}
}
