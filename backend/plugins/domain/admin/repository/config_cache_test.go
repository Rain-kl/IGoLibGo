// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"Wavelet/pkg/cache/ram"
	"Wavelet/plugins/domain/admin/repository"
)

func setupConfigCacheTest(t *testing.T) (*miniredis.Miniredis, redis.UniversalClient, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repository.SetRedisClient(rdb)
	repository.ResetSystemConfigRAMCacheForTest()

	cleanup := func() {
		repository.StopSystemConfigCacheListener()
		repository.ResetSystemConfigRAMCacheForTest()
		repository.SetRedisClient(nil)
		_ = rdb.Close()
		mr.Close()
	}
	return mr, rdb, cleanup
}

func TestSystemConfigCacheListener_InvalidateSingleKey(t *testing.T) {
	_, rdb, cleanup := setupConfigCacheTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. 本节点写入本地 RAM 缓存
	const testKey = "system_title"
	ram.Set(ram.CacheItem{
		Key:   testKey,
		Value: `{"key":"system_title","value":"Wavelet Admin"}`,
		Type:  repository.ConfigCacheType,
		TTL:   -1,
	})

	_, found := ram.Get(repository.ConfigCacheType, testKey)
	require.True(t, found, "RAM 缓存写入应成功")

	// 2. 启动系统配置监听器
	repository.StartSystemConfigCacheListener(ctx)

	// 3. 模拟远端节点广播失效单 key
	msgPayload, err := json.Marshal(map[string]string{
		"type": repository.ConfigCacheType,
		"key":  testKey,
	})
	require.NoError(t, err)
	require.NoError(t, rdb.Publish(ctx, repository.SystemConfigBroadcastChannel, msgPayload).Err())

	// 4. 断言本节点接收到广播后清理本地 RAM 缓存
	assert.Eventually(t, func() bool {
		_, ok := ram.Get(repository.ConfigCacheType, testKey)
		return !ok
	}, 2*time.Second, 20*time.Millisecond, "收到广播后对应 key 的本地 RAM 缓存应被自动失效")
}

func TestSystemConfigCacheListener_InvalidateAll(t *testing.T) {
	_, rdb, cleanup := setupConfigCacheTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. 本节点写入多条本地 RAM 缓存
	ram.Set(ram.CacheItem{Key: "cfg_1", Value: "val1", Type: repository.ConfigCacheType, TTL: -1})
	ram.Set(ram.CacheItem{Key: "cfg_2", Value: "val2", Type: repository.ConfigCacheType, TTL: -1})

	require.Len(t, ram.GetTypeItems(repository.ConfigCacheType), 2)

	repository.StartSystemConfigCacheListener(ctx)

	// 2. 模拟远端节点广播全量失效
	msgPayload, err := json.Marshal(map[string]string{
		"type": repository.ConfigCacheType,
		"key":  "*",
	})
	require.NoError(t, err)
	require.NoError(t, rdb.Publish(ctx, repository.SystemConfigBroadcastChannel, msgPayload).Err())

	// 3. 断言全量清理
	assert.Eventually(t, func() bool {
		return len(ram.GetTypeItems(repository.ConfigCacheType)) == 0
	}, 2*time.Second, 20*time.Millisecond, "全量广播应清空该类型的所有本地 RAM 缓存")
}

func TestSystemConfigCache_BroadcastOnInvalidate(t *testing.T) {
	_, rdb, cleanup := setupConfigCacheTest(t)
	defer cleanup()

	ctx := context.Background()

	// 订阅广播频道验证 Invalidate 是否真正发送了 Pub/Sub 广播
	sub := rdb.Subscribe(ctx, repository.SystemConfigBroadcastChannel)
	defer func() { _ = sub.Close() }()

	// 等待订阅就绪
	_, err := sub.Receive(ctx)
	require.NoError(t, err)

	ch := sub.Channel()

	// 1. 测试单 key 失效广播
	require.NoError(t, repository.InvalidateSystemConfigCache(ctx, "site_name"))

	select {
	case msg := <-ch:
		require.NotNil(t, msg)
		var payload map[string]string
		require.NoError(t, json.Unmarshal([]byte(msg.Payload), &payload))
		assert.Equal(t, repository.ConfigCacheType, payload["type"])
		assert.Equal(t, "site_name", payload["key"])
	case <-time.After(2 * time.Second):
		t.Fatal("超时未收到单 key 失效广播消息")
	}

	// 2. 测试全量失效广播
	require.NoError(t, repository.InvalidateAllSystemConfigCaches(ctx))

	select {
	case msg := <-ch:
		require.NotNil(t, msg)
		var payload map[string]string
		require.NoError(t, json.Unmarshal([]byte(msg.Payload), &payload))
		assert.Equal(t, repository.ConfigCacheType, payload["type"])
		assert.Equal(t, "*", payload["key"])
	case <-time.After(2 * time.Second):
		t.Fatal("超时未收到全量失效广播消息")
	}
}

func TestSystemConfigCacheListener_LifecycleSafety(t *testing.T) {
	_, _, cleanup := setupConfigCacheTest(t)
	defer cleanup()

	ctx := context.Background()

	// 多次调用启动和停止必须幂等且不 panic / 死锁
	repository.StartSystemConfigCacheListener(ctx)
	repository.StartSystemConfigCacheListener(ctx)

	repository.StopSystemConfigCacheListener()
	repository.StopSystemConfigCacheListener()

	// 再次启动也能正常工作
	repository.StartSystemConfigCacheListener(ctx)
	repository.StopSystemConfigCacheListener()
}
