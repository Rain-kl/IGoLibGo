---
name: "wv-new-async-task"
description: "Wavelet 项目专用：新增或修改基于 Cordis 插件的 Asynq 异步任务、后台 Worker 消费处理器、Cron 定时调度任务与任务执行追踪时必须使用。"
metadata:
  origin: Wavelet
---

# 异步任务与定时调度开发规范 (Cordis 插件化架构)

本技能是 Wavelet 在 Cordis 微内核与插件化架构下，进行 Asynq 异步后台任务与 Cron 定时调度开发的唯一指导规范。

---

## 1. 核心架构：插件内自包含任务声明

在 Cordis 架构中，后台 Worker 消费与定时调度**不依赖全局硬编码注册表**，而是由各个业务插件在自身的 `Apply` 方法中通过微内核扩展点直接声明。

### 扩展点矩阵

| 扩展点方法 | 说明 | 适用场景 |
| :--- | :--- | :--- |
| `ctx.Task().Register(pattern, handler, opts...)` | 注册 Asynq 任务类型与消费处理器 | 异步耗时计算、队列任务、通知外发 |
| `ctx.Schedule().RegisterCron(spec, taskType, payload)` | 注册 Cron 表达式定时调度任务 | 周期统计、定时清理、健康检查 |

---

## 2. 异步任务开发全流程

### 步骤 1：在插件内定义任务 Payload 与常量

根据物理子包分层规范，任务定义应放在 `consts/` 与 `service/` 或 `model/dto/` 中（严禁在插件根目录下平铺）：

```go
// backend/plugins/domain/order/consts/tasks.go
package consts

const (
	TaskTypeOrderTimeoutCancel = "order:timeout_cancel"
)

// backend/plugins/domain/order/model/dto/task.go
package dto

type OrderTimeoutPayload struct {
	OrderID   string `json:"order_id"`
	Reason    string `json:"reason"`
	CreatedAt int64  `json:"created_at"`
}
```

### 步骤 2：在 Service 中实现任务执行处理器 (Handler)

处理器可以是一个结构体方法或普通函数，接受 `context.Context` 和 `[]byte` payload：

```go
// backend/plugins/domain/order/service/order_task.go
package service

import (
	"context"
	"encoding/json"
	"Wavelet/pkg/logger"
	"Wavelet/plugins/domain/order/model/dto"
)

type OrderTaskHandler struct {
	svc *OrderService
}

func NewOrderTaskHandler(svc *OrderService) *OrderTaskHandler {
	return &OrderTaskHandler{svc: svc}
}

func (h *OrderTaskHandler) Execute(ctx context.Context, payload []byte) error {
	var p dto.OrderTimeoutPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.ErrorF(ctx, "failed to unmarshal task payload: %v", err)
		return err
	}

	return h.svc.CancelTimeoutOrder(ctx, p.OrderID, p.Reason)
}
```

### 步骤 3：在插件 `Apply` 中注册任务元数据

在 `plugin.go` 中通过 `ctx.Task().Register` 声明消费，并附加元数据：

```go
func (p *Plugin) Apply(ctx *core.Context) error {
	taskHandler := service.NewOrderTaskHandler(p.svc)

	ctx.Task().Register(
		consts.TaskTypeOrderTimeoutCancel,
		taskHandler.Execute,
		extpoints.WithTaskType("order_timeout_cancel"),
		extpoints.WithTaskName("订单超时自动取消"),
		extpoints.WithTaskDescription("定时检测未支付订单并释放库存"),
		extpoints.WithTaskCategory("order"),
		extpoints.WithTaskRetry(3),
		extpoints.WithTaskQueue("default"),
		extpoints.WithTaskRetryable(true),
	)

	return nil
}
```

---

## 3. 定时调度 (Cron Schedule)

若需要周期性执行（如每小时检查一次超时、每天凌晨统计数据）：

```go
func (p *Plugin) Apply(ctx *core.Context) error {
	// 每 10 分钟触发一次清理任务
	ctx.Schedule().RegisterCron(
		"*/10 * * * *",
		consts.TaskTypeOrderTimeoutCancel,
		map[string]any{"trigger": "cron_scheduler"},
	)
	return nil
}
```

---

## 4. 派发异步任务

在业务逻辑中，通过依赖注入获取 `contracts.TaskService`，向队列投递任务：

```go
import "Wavelet/core/contracts"

type OrderService struct {
	taskSvc contracts.TaskService
}

func (s *OrderService) EnqueueOrderTimeout(ctx context.Context, orderID string) error {
	payload, _ := json.Marshal(dto.OrderTimeoutPayload{
		OrderID:   orderID,
		Reason:    "auto_timeout_15m",
		CreatedAt: time.Now().Unix(),
	})

	return s.taskSvc.Enqueue(ctx, consts.TaskTypeOrderTimeoutCancel, payload, contracts.TaskOptions{
		ProcessIn: 15 * time.Minute,
	})
}
```
