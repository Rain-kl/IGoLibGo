---
name: wv-logging
description: Wavelet 项目专用：结构化日志与链路追踪规范。指导使用 backend/pkg/logger（基于 Zap + otelzap + 5000 行 GlobalRingBuffer 环形缓冲区，支持 Admin WebSocket 实时日志流）及 contracts.LoggerService，禁止使用裸 log 或未经受管的第三方日志库。
metadata:
  origin: Wavelet
---

# Wavelet 日志与链路追踪规范 (wv-logging)

本技能指导 Wavelet 项目中统一的日志方案设计、日志调用、链路追踪（OpenTelemetry）与管理后台日志流保障。

---

## 1. 核心架构与设计决策

Wavelet 项目**不使用纯裸 `log/slog`**，而是采用经过深度整合的 `backend/pkg/logger`：

```
                    +--------------------------------+
                    |       业务调用 / 插件代码       |
                    | logger.InfoF(ctx, "msg: %s", v)|
                    +---------------+----------------+
                                    |
                                    v
                    +--------------------------------+
                    |      backend/pkg/logger        |
                    |   (自动提取 TraceID / SpanID)   |
                    +---------------+----------------+
                                    |
                    +---------------+----------------+
                    |  Zap Core + MultiWriteSyncer   |
                    +-------+----------------+-------+
                            |                |
             [输出 1: 终端 / 文件]     [输出 2: GlobalRingBuffer]
             stdout / app.log        容量 5000 行内存环形缓冲区
                                             |
                                             v
                                  [Admin 控制台 WebSocket]
                                  管理员实时查看与流式推送
```

### 为什么必须统一使用 `backend/pkg/logger`？
1. **链路追踪绑定**：内置 `otelzap`，自动从 `ctx` 中提取 OpenTelemetry Trace ID 和 Span ID，注入结构化字段。
2. **管理后台实时控制台**：内置 `GlobalRingBuffer`（容量 5000 行），管理后台系统监控面板通过 WebSocket 实时监听该缓冲区以流式展示系统运行日志。若使用裸 `slog` 或第三方日志，会导致后台无法采集到日志。
3. **性能与 Goroutine 安全**：底层为 `uber-go/zap` 高性能零分配架构，配置多写器（MultiWriteSyncer）。

---

## 2. 代码调用标准

### 2.1 基础调用范式

导入包名：`Wavelet/pkg/logger`。
所有日志方法**必须将 `context.Context` 作为第一个参数传入**：

```go
import "Wavelet/pkg/logger"

// 1. 信息日志 (Info)
logger.InfoF(ctx, "user login success: uid=%d, ip=%s", user.ID, clientIP)

// 2. 错误日志 (Error) - 必须记录错误详情
if err != nil {
    logger.ErrorF(ctx, "failed to query order: order_id=%s, err=%v", orderID, err)
}

// 3. 警告日志 (Warn)
logger.WarnF(ctx, "cache fallback triggered: key=%s, reason=%s", cacheKey, reason)

// 4. 调试日志 (Debug)
logger.DebugF(ctx, "payload decoded: raw_len=%d", len(payload))
```

### 2.2 跨插件契约调用 (`contracts.LoggerService`)

在 Cordis 架构中，若通过服务契约容器使用日志服务，可注入 `contracts.LoggerService`：

```go
import "Wavelet/core/contracts"

type MyService struct {
    log contracts.LoggerService
}

func (s *MyService) DoWork(ctx context.Context) {
    s.log.Infof(ctx, "processing item %s", itemID)
    // 或键值风格
    s.log.Info(ctx, "processing item", "item_id", itemID)
}
```

---

## 3. 严格禁止项 (Anti-Patterns)

- ❌ **严禁使用裸标准库 `log.Println` / `log.Printf`**（丢失上下文、Trace ID 及环形缓冲区同步）。
- ❌ **严禁使用 `fmt.Println` 打印运行时日志**。
- ❌ **严禁在子包中重新实例化 `zap.New` 或 `slog.New`**（会脱离全局 RingBuffer）。
- ❌ **严禁静默吞掉关键错误**（`_ = err`），底层错误在 Logic/Handler 边界必须用 `logger.ErrorF` 记录。
- ❌ **严禁在日志中输出明文敏感信息**（密码、密钥、用户明文 Token、支付凭据；必须脱敏处理）。

---

## 4. 日志级别选用指南

| 级别 | 何时使用 | 示例 |
| :--- | :--- | :--- |
| **Debug** | 深入排查问题时的冗余诊断信息，生产环境通常关闭 | 报文原始内容、缓存未命中键名、SQL 参数 |
| **Info** | 系统的关键生命周期或重要业务状态变迁 | 插件加载完成、用户登录登出、异步任务完成、发布触发 |
| **Warn** | 预期外但系统已安全降级/自愈的事件 | 第三方 API 超时重试、缓存回落查数据库、限流触发拦截 |
| **Error** | 导致当前操作中断或需要人工排查的错误 | 数据库写入失败、外部服务彻底不可用、文件存储写入异常 |
