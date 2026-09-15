# Agent Note: Bot Runner 不得绑定 HTTP 请求 context，嵌套 config 必须展开到叶子

Status: implemented

## Problem

Docker 开启 Redis 后，管理端创建/删除消息通道会在约 15s 被前端 axios 掐断，随后 Telegram 无法收发。关掉 Redis 并重启后又恢复。根因不是 Redis 协议本身：

1. Channel CRUD 同步 `Reload(request.Context())`，Telegram `Connect` 把 `bot.Stop()` 绑在该 ctx 上。`tele.NewBot`（getMe）在服务器上约 22s，超过前端 15s 超时。客户端一断，Go 取消请求 ctx，long poll 立刻停掉。库里有通道，进程里没有 poller。
2. `syncChannels` 在 `Connect` 期间持有 runner 写锁，`SendText` 被卡住。
3. Asynq `BotDispatchHandler` 为发一条下行再 `Connect` 一条 getUpdates，Telegram 只允许一个 poller，主通道被踢掉。
4. `walkConfigFields` 把 `hostConfig.Log` 这种带 `config:"log"` 的嵌套结构体当成叶子，不展开 `log.level`。Docker 镜像没有 `config.yaml`，`LOG_LEVEL` 和 default 都绑不上，日志出现 `invalid log level: , defaulting to info`。

Redis 打开后只是让 CRUD 更容易撞上这条路径（Asynq + 管理端热加载）；进程重启会用 app ctx 再 `Start`，看起来像「去掉 Redis 就好了」。

## Decision

Bot 长连接的生命周期跟随 Runner，不跟随 HTTP 请求。

- `Runner` 在 `Start` 时保存 `life` context；`Connect` 始终传入 `life`。请求取消不能 `bot.Stop()`。
- `reloadMu` 串行化 Start/Reload/Stop；`channels` 的 `mu` 只保护 map。`Connect` 不再握着写锁，`SendText` 不会死锁。
- Create/Update/Delete 只落库并 `ReloadAsync()`，HTTP 不等 `getMe`。
- 下行分发走 `GlobalRunner.SendText`，禁止为发消息再开一条 long poll。
- Telegram HTTP Client 超时 30s，且必须大于 long-poll 窗口（10s），避免 getUpdates 被截断，也避免 Bot API 永久挂死 Reload。
- `walkConfigFields` 递归带 `config` 标签的嵌套结构体，叶子带 `index`，`Bind` 用 `FieldByIndex` 写入。`LOG_LEVEL` 绑定到 `log.level`，无 env/文件时走 default `info`。

相关入口：`bot_runner.go` 的 `Start`/`Reload`/`ReloadAsync`，`bot_channel.go` 的 CRUD，`bot_dispatch.go` 的 `dispatchOnChannel`，`extpoints/config.go` 的 `walkConfigFields`。

## Alternatives considered

- **CRUD 仍同步 Reload，只给 `tele.NewBot` 加短超时** — 最强论据是调用方能立刻知道 token 是否可用。否决：10s 超时仍接近前端 15s；失败时通道已写入，响应语义混乱。凭证探测留给独立的 Probe 接口（已有 10s timeout）。
- **Connect 继续用请求 ctx，靠 `context.WithoutCancel` 只包 Stop 观察者** — 能挡住 abort，但 HTTP 仍被 getMe 堵住，接口超时还在。
- **把 hostConfig 改成扁平字段，不动 walkConfigFields** — 能修 LOG_LEVEL，但每个嵌套 `config:"app"` / `config:"log"` 都会再踩一次。框架必须展开嵌套结构体。
- **Dispatch 失败时再临时 Connect** — 看起来更稳，实际会再踢主 poller。未连接就报错，由 Runner 负责在线。

## Consequences

- **收益**：管理端增删通道不再打死 long poll；写接口不再等 Telegram；Asynq 下发不再冲突 getUpdates；无 config.yaml 时 `LOG_LEVEL` 与 default 生效。
- **代价与已知上限**：CRUD 返回时 Bot 可能尚未连上，前端会先看到通道、后才真正收发。`getMe` 失败只打日志，通道仍在库里，下次 `Start`/`Reload` 再试。若 Probe 未做、token 无效，用户会感觉「创建成功但 Bot 没反应」。

## Verification

- `go test ./core/extpoints -run TestBindNestedStructFromEnvAndDefaults`
- `go test ./plugins/domain/msg_gateway/service -run 'TestRunner_ReloadDoesNotStopChannelWhenRequestContextCancels|TestRunner_SendTextDoesNotDeadlockWhileConnectRuns|TestCreateChannel_ReturnsBeforeConnectFinishes|TestBotDispatchReusesConnectedRunnerChannel'`
- `go test ./plugins/domain/msg_gateway/channels/telegram -run TestBuildTeleSettingsLongPollWindow`
