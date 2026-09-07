// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/plugins/infra/cache"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachePluginOperations(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer func() { _ = rdb.Close() }()

	p := cache.New(
		cache.WithRedis(rdb),
		cache.WithKeyPrefix("app:"),
		cache.WithRAMCapacity(500),
	)
	ctx := core.NewContext(context.Background())
	ctx.Config().SetSource(core.NewMapSource(map[string]any{
		"redis.enabled": true,
	}))
	require.NoError(t, ctx.Config().Resolve())
	require.NoError(t, p.Apply(ctx))

	svc, err := core.Inject[contracts.CacheService](ctx)
	require.NoError(t, err)

	type Data struct {
		Value string `json:"value"`
	}

	testCtx := context.Background()

	// 1. ErrCacheMiss
	var out Data
	err = svc.Get(testCtx, "missing", &out)
	assert.ErrorIs(t, err, contracts.ErrCacheMiss)

	// 2. Set & Get
	in := Data{Value: "hello"}
	require.NoError(t, svc.Set(testCtx, "key1", in, 5*time.Minute))

	require.NoError(t, svc.Get(testCtx, "key1", &out))
	assert.Equal(t, "hello", out.Value)

	// 3. GetOrSet
	var target Data
	err = svc.GetOrSet(testCtx, "key1", &target, time.Minute, func() (any, error) {
		return Data{Value: "from_loader"}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", target.Value)

	err = svc.GetOrSet(testCtx, "key2", &target, time.Minute, func() (any, error) {
		return Data{Value: "from_loader"}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, "from_loader", target.Value)

	// 4. Invalidate / Delete
	require.NoError(t, svc.Invalidate(testCtx, "key1"))
	err = svc.Get(testCtx, "key1", &out)
	assert.ErrorIs(t, err, contracts.ErrCacheMiss)

	// 5. Verify redis.UniversalClient is provided in core context
	injectedRDB, err := core.Inject[redis.UniversalClient](ctx)
	require.NoError(t, err)
	assert.Equal(t, rdb, injectedRDB)

	require.NoError(t, ctx.Dispose())
}

func TestCachePluginL1BackfillExpires(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer func() { _ = rdb.Close() }()

	p := cache.New(
		cache.WithRedis(rdb),
		cache.WithKeyPrefix("app:"),
		cache.WithRAMCapacity(500),
	)
	ctx := core.NewContext(context.Background())
	ctx.Config().SetSource(core.NewMapSource(map[string]any{
		"redis.enabled": true,
	}))
	require.NoError(t, ctx.Config().Resolve())
	require.NoError(t, p.Apply(ctx))
	defer func() { _ = ctx.Dispose() }()

	svc, err := core.Inject[contracts.CacheService](ctx)
	require.NoError(t, err)

	testCtx := context.Background()

	type Data struct {
		Value string `json:"value"`
	}

	// Direct write to Redis L2 with 50-millisecond TTL
	rawJSON := `{"value":"expiring"}`
	require.NoError(t, rdb.Set(testCtx, "app:ttl_key", rawJSON, 50*time.Millisecond).Err())

	// First Get: misses L1, hits L2, and backfills L1 with remaining PTTL
	var out Data
	require.NoError(t, svc.Get(testCtx, "ttl_key", &out))
	assert.Equal(t, "expiring", out.Value)

	// Wait for real time to elapse beyond the 50ms TTL and advance miniredis clock
	time.Sleep(80 * time.Millisecond)
	mr.FastForward(100 * time.Millisecond)

	// Second Get: L1 entry should have expired (expireAt was set from PTTL) and L2 also expired.
	// Therefore, it must return ErrCacheMiss rather than immortal cached value.
	var expiredOut Data
	err = svc.Get(testCtx, "ttl_key", &expiredOut)
	assert.ErrorIs(t, err, contracts.ErrCacheMiss, "L1 cache backfilled from Redis must expire instead of becoming immortal")
}

func TestCachePluginL1BackfillNoExpiryKeyHasSafeTTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer func() { _ = rdb.Close() }()

	p := cache.New(
		cache.WithRedis(rdb),
		cache.WithKeyPrefix("app:"),
		cache.WithRAMCapacity(500),
	)
	ctx := core.NewContext(context.Background())
	ctx.Config().SetSource(core.NewMapSource(map[string]any{
		"redis.enabled": true,
	}))
	require.NoError(t, ctx.Config().Resolve())
	require.NoError(t, p.Apply(ctx))
	defer func() { _ = ctx.Dispose() }()

	svc, err := core.Inject[contracts.CacheService](ctx)
	require.NoError(t, err)

	testCtx := context.Background()

	type Data struct {
		Value string `json:"value"`
	}

	// Key with no TTL in Redis
	rawJSON := `{"value":"persistent"}`
	require.NoError(t, rdb.Set(testCtx, "app:persist_key", rawJSON, 0).Err())

	var out Data
	require.NoError(t, svc.Get(testCtx, "persist_key", &out))
	assert.Equal(t, "persistent", out.Value)
}
