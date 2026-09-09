// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/admin/model"
	"Wavelet/plugins/domain/admin/repository"
	cacheplugin "Wavelet/plugins/infra/cache"
)

// stubDBService 用内存 SQLite 满足 DBService 契约，隔离外部依赖。
type stubDBService struct{ db *gorm.DB }

func (s stubDBService) GORM() *gorm.DB { return s.db }

func (s stubDBService) DB(context.Context) *gorm.DB { return s.db }

func (s stubDBService) Named(string) *gorm.DB { return s.db }

// newFlushLogTestCache 构建真实多层缓存服务并注入 admin 插件上下文。
func newFlushLogTestCache(t *testing.T) (contracts.CacheService, *miniredis.Miniredis, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaintNotificationsConfig: &maintnotifications.Config{Mode: maintnotifications.ModeDisabled}})

	p := cacheplugin.New(cacheplugin.WithRedis(rdb), cacheplugin.WithRAMCapacity(64))
	ctx := core.NewContext(context.Background())
	ctx.Config().SetSource(core.NewMapSource(map[string]any{
		"redis": map[string]any{
			"enabled": true,
			"addrs":   []string{mr.Addr()},
		},
	}))
	require.NoError(t, ctx.Config().Resolve())
	require.NoError(t, p.Apply(ctx))
	svc, err := ctx.Inject[contracts.CacheService]()
	require.NoError(t, err)

	repository.SetCacheService(svc)
	repository.SetRedisClient(rdb)
	cleanup := func() {
		repository.SetRedisClient(nil)
		repository.SetCacheService(nil)
		_ = rdb.Close()
		mr.Close()
	}
	return svc, mr, cleanup
}

// TestFlushTaskExecutionLogPropagatesCacheError 回归：缓存读取失败（非未命中）时，
// FlushTaskExecutionLog 必须返回错误而不是静默吞掉日志并误报成功（nilerr 修复）。
func TestFlushTaskExecutionLogPropagatesCacheError(t *testing.T) {
	_, mr, cleanup := newFlushLogTestCache(t)
	defer cleanup()

	ctx := context.Background()
	const taskID = "flush-err-task"

	// 先缓冲一行日志
	require.NoError(t, repository.AppendTaskExecutionLog(ctx, taskID, "step-1 ok"))

	// 关闭 miniredis 模拟缓存基础设施故障（读取出错而非未命中）
	mr.Close()

	err := repository.FlushTaskExecutionLog(ctx, taskID)
	assert.Error(t, err, "缓存故障时必须返回错误，防止缓冲日志被静默丢弃")
}

// TestFlushTaskExecutionLogCacheMissIsNoop 回归：任务无缓冲日志（未命中）时应为空操作成功。
func TestFlushTaskExecutionLogCacheMissIsNoop(t *testing.T) {
	_, _, cleanup := newFlushLogTestCache(t)
	defer cleanup()

	ctx := context.Background()
	assert.NoError(t, repository.FlushTaskExecutionLog(ctx, "missing-task"))
}

// TestFlushTaskExecutionLogPersistsAndClears 验证正常路径：缓冲日志写入执行记录后清理缓存。
func TestFlushTaskExecutionLogPersistsAndClears(t *testing.T) {
	_, _, cleanup := newFlushLogTestCache(t)
	defer cleanup()

	ctx := context.Background()
	const taskID = "flush-ok-task"
	require.NoError(t, repository.AppendTaskExecutionLog(ctx, taskID, "done"))

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(&model.TaskExecution{}))
	repository.SetDBService(stubDBService{db: sqliteDB})
	defer repository.SetDBService(nil)
	gormDB := sqliteDB
	exec := &model.TaskExecution{TaskID: taskID, TaskType: "upload:test", TaskName: "t", Status: model.TaskExecutionStatusSucceeded}
	require.NoError(t, gormDB.Create(exec).Error)

	require.NoError(t, repository.FlushTaskExecutionLog(ctx, taskID))

	var got model.TaskExecution
	require.NoError(t, gormDB.First(&got, exec.ID).Error)
	assert.Contains(t, got.Log, "done")

	// 缓存中的缓冲日志应已被清理
	rdb := repository.GetRedisClient(ctx)
	require.NotNil(t, rdb)
	logLines, err := rdb.LRange(ctx, repository.TaskExecutionLogRedisKey(taskID), 0, -1).Result()
	require.NoError(t, err)
	assert.Empty(t, logLines, "flush 后缓存应清空")
}

// TestAppendTaskExecutionLogKeepsBufferOnCacheReadError 回归：Redis 故障或未初始化时，
// AppendTaskExecutionLog 必须报错上抛，且采用 LIST RPush 不会因读失败而覆盖已有缓冲日志。
func TestAppendTaskExecutionLogKeepsBufferOnCacheReadError(t *testing.T) {
	// 1. 未初始化 Redis 时必须返回错误
	repository.SetRedisClient(nil)
	err := repository.AppendTaskExecutionLog(context.Background(), "append-err-task", "step-1")
	assert.Error(t, err, "未初始化 Redis 时必须上抛错误")

	// 2. Redis 故障时也必须返回错误
	_, mr, cleanup := newFlushLogTestCache(t)
	defer cleanup()

	ctx := context.Background()
	const taskID = "append-fail-task"
	require.NoError(t, repository.AppendTaskExecutionLog(ctx, taskID, "line-1 ok"))

	mr.Close()
	err = repository.AppendTaskExecutionLog(ctx, taskID, "line-2 fail")
	assert.Error(t, err, "Redis 不可用时必须返回错误，而不是静默丢失或覆盖")
}

// TestTaskExecutionLogUsesRedisList 验证任务日志底层严格采用 LIST 数据结构，杜绝与 Worker 的 WRONGTYPE 冲突。
func TestTaskExecutionLogUsesRedisList(t *testing.T) {
	_, _, cleanup := newFlushLogTestCache(t)
	defer cleanup()

	ctx := context.Background()
	const taskID = "task-list-data-type-check"
	key := repository.TaskExecutionLogRedisKey(taskID)
	rdb := repository.GetRedisClient(ctx)
	require.NotNil(t, rdb)

	// 追加两条日志
	require.NoError(t, repository.AppendTaskExecutionLog(ctx, taskID, "first log line"))
	require.NoError(t, repository.AppendTaskExecutionLog(ctx, taskID, "second log line"))

	// 验证 Redis key 的类型严格为 list
	keyType, err := rdb.Type(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, "list", keyType, "任务日志必须存储为 Redis LIST")

	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, sqliteDB.AutoMigrate(&model.TaskExecution{}))
	repository.SetDBService(stubDBService{db: sqliteDB})
	defer repository.SetDBService(nil)

	execDirect := &model.TaskExecution{TaskID: taskID}
	execDirect.ID = 1001
	execDirect.TaskType = "test"
	require.NoError(t, sqliteDB.Create(execDirect).Error)

	gotExec, err := repository.GetTaskExecutionByTaskID(ctx, taskID)
	require.NoError(t, err)
	require.NotNil(t, gotExec)
	assert.Contains(t, gotExec.Log, "first log line")
	assert.Contains(t, gotExec.Log, "second log line")

	// 验证批量加载 loadTaskExecutionLogs
	batchList, total, err := repository.ListTaskExecutionRecords(ctx, model.ListTaskExecutionsRequest{
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Contains(t, batchList[0].Log, "first log line")

	// 验证 FlushTaskExecutionLog 持久化并清理
	require.NoError(t, repository.FlushTaskExecutionLog(ctx, taskID))
	var flushed model.TaskExecution
	require.NoError(t, sqliteDB.First(&flushed, execDirect.ID).Error)
	assert.Contains(t, flushed.Log, "first log line")
	assert.Contains(t, flushed.Log, "second log line")

	// 验证 Redis key 已被清理
	postFlushLines, err := rdb.LRange(ctx, key, 0, -1).Result()
	require.NoError(t, err)
	assert.Empty(t, postFlushLines)
}
