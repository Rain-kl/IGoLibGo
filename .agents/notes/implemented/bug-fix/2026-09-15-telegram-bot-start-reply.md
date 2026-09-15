# Agent Note: 消息网关未鉴权私聊消息自动回复配对码与 Bot Runner 启动修正

Status: implemented

## Problem

之前消息网关 (msg_gateway) 存在两个严重缺陷：
1. `bot_runner.go` 中的 `service.Start(ctx)` 为空桩实现，只打印了日志，未从数据库加载已启用的 Bot 频道并开启 Connect/轮询，导致后台 Telegram Bot 从未真正上线。
2. 缺少通用 Inbound 消息处理逻辑，当私聊用户发送 `/start` 或任意文本时，无法识别用户鉴权/绑定状态并自动回复配对码。

## Decision

1. **通用未鉴权私聊自动回复**：不注册单独的 `/start` 命令，对所有来自未绑定平台身份（`PlatformUserID`）的私聊消息，自动生成/复用 15 分钟内有效的 8 位配对码（如 `ABCD-EFGH`）并回复指导信息。对于已绑定的身份回复“账号已成功绑定”。
2. **Bot Runner 生命周期管理**：在 `bot_runner.go` 中实现全量 Channel 加载、连接与平滑重启 (`Reload`)；默认在 `msg_gateway` 插件加载时启动 Runner。Channel CRUD 提交后触发 `ReloadAsync`，Connect 绑定 Runner lifetime 而非 HTTP 请求 context（见 [bot-runner-request-ctx-and-nested-config](2026-09-15-bot-runner-request-ctx-and-nested-config.md)）。

3. **`MessagePairingCode` 实体与 Goose SQL 迁移 Schema 严格对齐**：修复 Go 模型 `entity.MessagePairingCode` 错配包含不存在的 `id` 与 `user_id` 字段的问题，使其严格与 Goose SQL 迁移 `00001_initial.sql` 中 `code VARCHAR(16) PRIMARY KEY` 保持一致，消除 `no such column: w_message_pairing_codes.id` 数据库报错。

## Alternatives considered

- **只硬编码处理 `/start` 命令**：否决。用户可能发送 `help`、`hello` 或点击 Telegram 内置按钮，统一对所有未鉴权/未绑定私聊消息回复配对码操作更加直观且不容易因指令拼写不同产生漏洞。
- **由前端/定时器轮询拉取**：否决。通过 WebSocket / Telegram LongPoller 在收到 Inbound 消息时实时同步回复性能与即时性最佳。

## Consequences

- 任何未绑定用户向 Bot 发送任何私聊消息都能立即拿到配对码进行绑定。
- 后台 Bot 能够在服务启动及频道配置更新时平滑开启/热加载长轮询通道。
- `w_message_pairing_codes` 表操作完全符合 Goose DDL 定义。
