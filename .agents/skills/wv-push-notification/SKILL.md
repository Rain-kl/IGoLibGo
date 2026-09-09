---
name: "wv-push-notification"
description: "Wavelet 项目专用：当需要开发或接入新的系统通知推送事件、修改消息推送底层设计、调用统一触发器投递消息、或开发带消息推送功能的业务功能时必须使用。本技能指导元数据声明、触发流程、解耦防线和动态同步机制。"
metadata:
  origin: Wavelet
---

# 新增消息推送与通知事件开发规范 (Cordis 架构)

本技能指导在 Wavelet 的 Cordis 微内核架构下，如何开发系统通知推送事件、扩展推送渠道以及跨插件触发通知。

---

## 1. 消息推送分层架构 (`backend/plugins/domain/msg_gateway`)

消息推送域以独立 Cordis 插件形式实现于 `backend/plugins/domain/msg_gateway`，严禁在业务插件中硬编码推送逻辑。

| 子包路径 | 职责定位 | 核心细节与规范 |
| :--- | :--- | :--- |
| **`push/`** | 底层推送渠道抽象 | 纯基础设施。定义 `Pusher` 接口与具体渠道实现（Email、Lark、Webhook、Bark 等），不依赖数据库。 |
| **`channels/`** | 双向 Bot 适配 | 平台 Bot 交互引擎（Telegram、QQ 等），处理指令配对与双向消息流。 |
| **`controller/`** | HTTP 接口层 | 挂载于 `/api/v1/admin/push` 与 `/api/v1/admin/channels`，提供管理台配置、测试发送与历史查询。 |
| **`service/`** | 业务逻辑与派发引擎 | `push_trigger.go`（统一触发器）、`push_worker.go`（Asynq 异步任务消费）、`push_event.go`（事件注册中心）。 |
| **`dao/` & `model/`** | 数据访问与实体 | 管理 `w_push_events`（通知事件元数据与开关）、`w_push_histories`（推送审计日志）、`w_message_channels` 等。 |

---

## 2. 跨插件推送通信模式（防线与解耦）

根据 Cordis 微内核防线，**业务插件严禁直接 import `msg_gateway` 的内部实现**。跨插件触发通知必须采用以下两种受管模式之一：

### 模式 A：事件总线异步解耦触发（推荐，强解耦）

业务插件仅依赖 `Wavelet/core`，通过微内核事件总线广播：

```go
// 业务插件在用户注册成功、订单状态变更等处广播通知事件
ctx.Events().Emit("notification:push", do.PushNotificationEvent{
    UserID:   user.ID,
    Channel:  "user_registered",
    Title:    "新用户注册提醒",
    Content:  fmt.Sprintf("用户 %s (%s) 已成功加入系统", user.Username, user.Email),
    Metadata: map[string]any{"user_id": user.ID, "role": user.Role},
})
```

`msg_gateway` 插件在 `Apply` 中已内置监听 `notification:push` 并自动驱动异步派发。

### 模式 B：通过契约注册内置事件 (`contracts.PushRegistry`)

若插件需要在系统启动时声明自己的内置通知模板元数据（以便在管理后台展示并允许管理员自定义开关与渠道）：

1. 在 `Apply` 中通过微内核获取 `contracts.PushRegistry`：

```go
import "Wavelet/core/contracts"

func (p *Plugin) Apply(ctx *core.Context) error {
    core.Bind[contracts.PushRegistry](ctx, func(registry contracts.PushRegistry) {
        registry.RegisterBuiltInEvent(contracts.PushEventMeta{
            Key:         "order_paid",
            Name:        "订单支付成功通知",
            Description: "当用户成功完成订单支付时向管理员或用户发送提醒",
            DefaultTemplate: contracts.PushNotificationTemplate{
                Title:   "订单支付通知",
                Content: "订单 {{order.id}} 已支付成功，金额: ￥{{order.amount}}。",
                Level:   "INFO",
            },
        })
    })
    return nil
}
```

2. 触发通知时通过统一触发器或 EventBus 投递，系统将自动从 `w_push_events` 读取管理员定制的渠道与模板并渲染推送。

---

## 3. 异步任务与历史审计

1. **异步派发保护**：所有推送操作必须走 Asynq 队列（任务名为 `consts.TaskPushNotification`），禁止在 HTTP 请求生命周期内进行阻塞式远程 HTTP/SMTP 推送。
2. **审计留痕**：每一次推送尝试无论成功或失败，均自动记录到 `w_push_histories` 表，包含耗时、HTTP 状态码、错误信息与渲染后的实际内容，供管理后台排查。
