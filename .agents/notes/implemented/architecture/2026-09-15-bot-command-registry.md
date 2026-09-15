# Agent Note: msg_gateway 命令接入契约与业务指令外置

Status: implemented

## Problem

`msg_gateway` 是通道、配对、推送的底层插件，但下游 IGoLibGo 把 `/show`、`/run` 和 TraceInt 补录写进了网关内部，并 import `igo` 的 dao/service。框架插件依赖下游业务，命令也无法被其它插件注册。已绑定用户还没有查看身份的指令，只能吃到写死的「绑定成功」回复。

## Decision

网关对外提供 `contracts.BotCommandRegistry`（与 `PushRegistry` 同级）。业务插件在自己的 `bot/` 包实现 `BotCommand` / `BotConversation`，`Apply` 里 `Register`。网关独占入站分发、配对、对话占位和已连接通道上的 Reply。

公开契约在 `backend/core/contracts/bot.go`。`msg_gateway.Apply` 创建注册表、注册自带指令后 `Provide[contracts.BotCommandRegistry]`。

自带指令放在 `msg_gateway/bot/`：`/help`（别名 `/start`）、`/me`、`/cancel`。`/me` 展示绑定与 Wavelet 用户身份，不展示业务数据。命令名与别名大小写不敏感全局唯一，`Register` 冲突则 `panic`/启动失败。

多轮输入（补录）用网关对话占位：key 为 `(channelID, platformUserID)`，state 对网关透明，存储走 `CacheService`。igo 只实现 `igo.login_auth` / `igo.checkin_auth`。下游 IGoLibGo 尚未迁命令。

完整接口、分发表与命令开发指南见 [设计文档](../../../../docs/superpowers/specs/2026-09-15-bot-command-registry-design.md) 与 [wv-new-bot-command](../../../skills/wv-new-bot-command/SKILL.md)。

## Alternatives considered

- **业务插件自管 Cache session + 非命令 fallback：** 每个插件重做占位、TTL、取消；网关要轮询谁要这条文本。占位是通道身份上的排他资源，应留在网关。
- **`ctx.BotCommands()` 做成微内核扩展点：** 注册写法更短，但会把消息域能力塞进 `core/extpoints`。命令属于 `msg_gateway`，不是 Router/Task 那种内核机制。
- **只广播入站事件、不建命令表：** 网关最瘦，但 `/help` 聚合、启动查重、配对优先权全靠约定，没有「命令接入契约」。

## Consequences

- **收益**：网关只承载通道、配对、分发与自带指令；业务命令落在各自插件的 `bot/` 包，业务插件不 import `msg_gateway` 内部包。已绑定用户用 `/me` 查看身份，静默等待命令而不是写死的「绑定成功」文案。
- **代价**：已绑定非命令文本只有已注册对话才能进入 `OnMessage`，否则丢弃。`ctx.Bind` 不能返回 error，注册冲突靠 `panic` 停启动（本期业务插件尚未接入，冲突只在将来接入时出现）。对话占位无跨节点锁，同一用户连发可能交叉；真出现串话再加 per-key 锁。下游尚未迁命令；igo 对注册表是软依赖，没有 `msg_gateway` 时命令不出现、HTTP 仍启动。
