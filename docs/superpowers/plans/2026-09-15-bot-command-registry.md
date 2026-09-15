# Bot Command Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Wavelet 把 `msg_gateway` 收成命令底层：契约、注册表、入站分发、自带 `/help` `/me` `/cancel`，并补上 `wv-new-bot-command` 技能。

**Architecture:** 业务插件只面向 `contracts.BotCommandRegistry`。网关 Provide 注册表、解析斜杠命令、管对话占位，经 Runner 已连接通道 Reply。自带指令在 `msg_gateway/bot/`，与业务指令同一张表。

**Tech Stack:** Go、Cordis `Provide`/`Bind`、`contracts.CacheService`、现有 Telegram/QQ adapter 与 pairing。

**Spec:** `docs/superpowers/specs/2026-09-15-bot-command-registry-design.md`

## Global Constraints

- 只改 Wavelet。禁止改 IGoLibGo，禁止在网关里 import `igo`。
- 业务插件不得 import `Wavelet/plugins/domain/msg_gateway/...`；只 import `Wavelet/core/contracts`。
- 已有 `service/bot_dispatch.go` 是下行 Asynq 任务，不要覆盖。入站分发用 `service/bot_inbound.go`。
- `service` 不得 import `msg_gateway/bot`；由 `plugin.go` 组装。
- `/start` 与 `/help` 同一 Handle。`/me` 不展示 email/phone/业务数据。
- `Register` 冲突必须让 `Apply` 失败。`ctx.Bind` 回调不能返回 error，业务侧（本期不做）用 panic；本期自带指令在 `Apply` 里 `return err`。
- 回复走 Runner `SendText`，禁止为 Reply 再 Connect 适配器。
- 测试用 `t.TempDir()`，表驱动，`make format` / 相关 `go test` 通过后再按任务提交。禁止 push。
- 用户可见文案本期可用中文硬编码（机器人私聊，不是 Next.js UI）。

## File map

Create:

- `backend/core/contracts/bot.go`
- `backend/core/contracts/bot_test.go`
- `backend/plugins/domain/msg_gateway/service/bot_registry.go`
- `backend/plugins/domain/msg_gateway/service/bot_registry_test.go`
- `backend/plugins/domain/msg_gateway/service/bot_inbound.go`
- `backend/plugins/domain/msg_gateway/service/bot_inbound_test.go`
- `backend/plugins/domain/msg_gateway/service/bot_session.go`
- `backend/plugins/domain/msg_gateway/bot/help.go`
- `backend/plugins/domain/msg_gateway/bot/me.go`
- `backend/plugins/domain/msg_gateway/bot/cancel.go`
- `backend/plugins/domain/msg_gateway/bot/builtins.go`
- `backend/plugins/domain/msg_gateway/bot/help_test.go`
- `backend/plugins/domain/msg_gateway/bot/me_test.go`
- `.agents/skills/wv-new-bot-command/SKILL.md`

Modify:

- `backend/plugins/domain/msg_gateway/plugin.go` — Provide 注册表、注册自带指令
- `backend/plugins/domain/msg_gateway/service/bot_runner.go` — inbound 改 `Registry.Dispatch`
- `backend/plugins/domain/msg_gateway/service/bot_pairing.go` — 抽出配对回复；已绑定不再写死成功文案
- `backend/plugins/domain/msg_gateway/service/bot_pairing_test.go` — 已绑定走分发器，不再断言「绑定成功」
- `backend/plugins/domain/msg_gateway/plugin_registry_test.go` — 断言 `BotCommandRegistry` 可 Inject
- `backend/plugins/domain/msg_gateway/consts/bot.go` — 自带命令名常量
- `AGENTS.md` — 增加 `wv-new-bot-command` 行
- `.agents/skills/README.md` — 自研技能 10→11，表里加一行
- `.agents/notes/proposed/architecture/2026-09-15-bot-command-registry.md` → `implemented/`（最后一任务）

Out of scope: `igo/`、`bot_command_handler.go`（只存在于 IGoLibGo）。

契约补丁（相对 spec）：`BotCommandRequest` 增加 `HasConversation() bool` 与 `CancelActive() error`，供 `/cancel` 使用（cancel 是命令不是对话，必须能结束占位）。

---

### Task 1: contracts.BotCommand*

**Files:**

- Create: `backend/core/contracts/bot.go`
- Create: `backend/core/contracts/bot_test.go`

**Interfaces:**

- Produces: `BotInbound`, `BotCommand`, `BotCommandRequest`, `BotConversation`, `BotConversationRequest`, `BotCommandRegistry`

- [ ] **Step 1: Write the failing test**

```go
package contracts_test

func TestBotCommandRegistry_isInterface(t *testing.T) {
    var _ contracts.BotCommandRegistry = stubRegistry{}
    var _ contracts.BotCommand = stubCmd{}
    var _ contracts.BotConversation = stubConv{}
}
```

`stubCmd` 必须实现 `Name/Aliases/Description/Usage/Handle`。`stubConv` 实现 `Name/OnMessage/OnCancel`。编译失败即 RED。

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./core/contracts/ -run TestBotCommandRegistry_isInterface -count=1`

Expected: FAIL compile — `BotCommandRegistry` undefined

- [ ] **Step 3: Write `bot.go`**

```go
package contracts

import (
    "context"
    "time"
)

type BotInbound struct {
    ChannelID      uint64
    ChannelType    string
    PlatformUserID string
    ChatID         string
    MessageID      string
    Text           string
    UserID         uint64
}

type BotCommand interface {
    Name() string
    Aliases() []string
    Description() string
    Usage() string
    Handle(ctx context.Context, req BotCommandRequest) error
}

type BotCommandRequest interface {
    UserID() uint64
    Args() []string
    Text() string
    Inbound() BotInbound
    Reply(text string) error
    Begin(conversation string, state any, ttl time.Duration) error
    HasConversation() bool
    CancelActive() error
}

type BotConversation interface {
    Name() string
    OnMessage(ctx context.Context, req BotConversationRequest) error
    OnCancel(ctx context.Context, req BotConversationRequest) error
}

type BotConversationRequest interface {
    UserID() uint64
    Text() string
    Inbound() BotInbound
    Reply(text string) error
    State(dst any) error
    SetState(v any) error
    Transition(conversation string, state any) error
    End() error
}

type BotCommandRegistry interface {
    Register(cmd BotCommand) error
    RegisterConversation(conv BotConversation) error
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./core/contracts/ -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/core/contracts/bot.go backend/core/contracts/bot_test.go
git commit -m "feat(contracts): add bot command registry interfaces"
```

---

### Task 2: Registry uniqueness

**Files:**

- Create: `backend/plugins/domain/msg_gateway/service/bot_registry.go`
- Create: `backend/plugins/domain/msg_gateway/service/bot_registry_test.go`
- Modify: `backend/plugins/domain/msg_gateway/consts/bot.go` — 增加 `CommandHelp/CommandStart/CommandMe/CommandCancel`

**Interfaces:**

- Consumes: `contracts.BotCommand`, `BotConversation`, `BotCommandRegistry`
- Produces: `NewBotRegistry() *BotRegistry`, `(*BotRegistry).Register`, `RegisterBuiltin(cmd, allowUnbound bool)`, `RegisterConversation`, `LookupCommand(name string) (cmd, builtin, allowUnbound, ok)`, `LookupConversation(name string)`, `List() []CommandMeta`

`CommandMeta` 字段：`Name string`, `Aliases []string`, `Description string`, `Usage string`, `Builtin bool`。

- [ ] **Step 1: Write failing tests** in `bot_registry_test.go`（`package service_test`）

表驱动至少覆盖：

- 成功注册业务命令后 `LookupCommand("Show")` 命中（大小写不敏感）
- 空 Name / Description / Usage → error
- Name 含 `/` 或空白 → error
- 重复 Name → error
- 别名与已有 Name 冲突（`help` 已 builtin，再 Register `foo` alias `HELP` → error）
- 同一命令 Aliases 自撞 → error
- 对话名无 `.` → error；重复对话名 → error
- `RegisterBuiltin` 的命令 `List()` 里 `Builtin==true`

用最小 stub：

```go
type stubCmd struct {
    name, desc, usage string
    aliases           []string
}

func (s stubCmd) Name() string                            { return s.name }
func (s stubCmd) Aliases() []string                       { return s.aliases }
func (s stubCmd) Description() string                     { return s.desc }
func (s stubCmd) Usage() string                           { return s.usage }
func (s stubCmd) Handle(context.Context, contracts.BotCommandRequest) error {
    return nil
}
```

- [ ] **Step 2: Run tests (RED)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -run TestBotRegistry -count=1`

Expected: FAIL — `NewBotRegistry` undefined

- [ ] **Step 3: Implement `bot_registry.go`**

规则：

- 归一化：`strings.ToLower(strings.TrimSpace(name))`
- 占用集合：每个 Name 与每个 Alias 都进同一 map
- `Register` 走业务路径，`allowUnbound=false`，`Builtin=false`
- `RegisterBuiltin(cmd, allowUnbound bool)` 不在 `contracts.BotCommandRegistry` 上（Go 方法多出来不影响接口满足）
- 对话名 `strings.ToLower`，必须包含 `.`
- `BotRegistry` 实现 `contracts.BotCommandRegistry`

错误用 `fmt.Errorf`，文案含冲突的名字，便于启动日志。

- [ ] **Step 4: Run tests (GREEN)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -run TestBotRegistry -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/plugins/domain/msg_gateway/service/bot_registry.go \
        backend/plugins/domain/msg_gateway/service/bot_registry_test.go \
        backend/plugins/domain/msg_gateway/consts/bot.go
git commit -m "feat(msg_gateway): add bot command registry with startup uniqueness checks"
```

---

### Task 3: Parse + inbound dispatch (no conversation yet)

**Files:**

- Create: `backend/plugins/domain/msg_gateway/service/bot_inbound.go`
- Create: `backend/plugins/domain/msg_gateway/service/bot_inbound_test.go`
- Modify: `backend/plugins/domain/msg_gateway/service/bot_pairing.go` — 导出 `ReplyPairingCode(ctx, msg, sendFn) error`（现有未绑定逻辑）；`HandleInboundMessage` 改为：未绑定调用 `ReplyPairingCode`，已绑定 **不再发「绑定成功」**，返回 nil

**Interfaces:**

- Consumes: `BotRegistry.LookupCommand`, dao binding lookup, `ReplyPairingCode`
- Produces: `ParseBotCommand(text) (name string, args []string, isCmd bool)`, `(*BotRegistry).Dispatch(ctx, msg do.InboundMessage, send SendTextFn) error`

`SendTextFn` 签名必须与 runner 一致：

```go
type SendTextFn func(ctx context.Context, channelID uint64, to do.Recipient, text string) error
```

- [ ] **Step 1: Write failing parse tests**

```go
func TestParseBotCommand(t *testing.T) {
    tests := []struct {
        in     string
        isCmd  bool
        name   string
        args   []string
    }{
        {"/help", true, "help", nil},
        {"/HELP", true, "help", nil},
        {"/run@MyBot id1", true, "run", []string{"id1"}},
        {"/run  myseat01", true, "run", []string{"myseat01"}},
        {"hello", false, "", nil},
        {"", false, "", nil},
        {"  ", false, "", nil},
    }
    // assert ParseBotCommand
}
```

Dispatch 测试（sqlite + `t.TempDir()`，照 `bot_pairing_test.go` 建表）：

- 未绑定 `/help` 且已 `RegisterBuiltin(help, true)` → 走到 Handle，**不**含「配对码」
- 未绑定 `hello` → 含「配对码」
- 未绑定未知 `/foo` → 配对码（不是「未知命令」）
- 已绑定 `/help` → Handle
- 已绑定未知 `/foo` → 回复含 `/help`
- 已绑定 `hello`（非命令）→ 不发送
- 已绑定 `/start` 若是 help 的 alias → 与 `/help` 同一 Handle（可先在 Task 5 用真 help；本期用 stub 把 start 注册为 alias）

`HandleInboundMessage` 已绑定用例改为：`sentText == ""`（不再「绑定成功」）。未绑定用例保持配对码。

- [ ] **Step 2: Run tests (RED)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -run 'TestParseBotCommand|TestBotRegistry_Dispatch|TestHandleInboundMessage' -count=1`

Expected: FAIL — `ParseBotCommand` / `Dispatch` undefined；`TestHandleInboundMessage_BoundUser` 仍期望「绑定成功」

- [ ] **Step 3: Implement parse + Dispatch**

`Dispatch` 伪代码：

```
trim text; if empty return nil
lookup binding by channel+platform
parse command
if unbound:
    if isCmd, lookup, allowUnbound: call Handle with UserID=0; return
    return ReplyPairingCode(...)
// bound
fill msg.BindingUserID / contracts.BotInbound.UserID
if isCmd:
    if lookup ok: Handle (cancel/conversation 下任务再加)
    else: Reply "未知命令，发送 /help 查看可用指令"
    return
return nil  // drop non-command
```

请求对象（本任务最小实现）：

```go
type commandRequest struct {
    userID uint64
    args   []string
    text   string
    in     contracts.BotInbound
    send   SendTextFn
}

func (r *commandRequest) Reply(text string) error {
    return r.send(ctx, r.in.ChannelID, do.Recipient{ChatID: r.in.ChatID, PlatformUserID: r.in.PlatformUserID}, text)
}
func (r *commandRequest) Begin(string, any, time.Duration) error { return errNoConversationYet }
func (r *commandRequest) HasConversation() bool                 { return false }
func (r *commandRequest) CancelActive() error                   { return nil }
```

Handle panic 用 `defer recover` 打日志，return nil。Handle error 只 `logger.ErrorF`，不 Reply。

把 `do.InboundMessage` 映到 `contracts.BotInbound`（字段一一对应，`UserID` 来自 binding）。

- [ ] **Step 4: Run tests (GREEN)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -run 'TestParseBotCommand|TestBotRegistry_Dispatch|TestHandleInboundMessage' -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/plugins/domain/msg_gateway/service/bot_inbound.go \
        backend/plugins/domain/msg_gateway/service/bot_inbound_test.go \
        backend/plugins/domain/msg_gateway/service/bot_pairing.go \
        backend/plugins/domain/msg_gateway/service/bot_pairing_test.go
git commit -m "feat(msg_gateway): dispatch inbound slash commands before pairing"
```

---

### Task 4: Conversation occupancy

**Files:**

- Create: `backend/plugins/domain/msg_gateway/service/bot_session.go`
- Modify: `backend/plugins/domain/msg_gateway/service/bot_inbound.go` — Dispatch 接入对话
- Modify: `backend/plugins/domain/msg_gateway/service/bot_inbound_test.go`

**Interfaces:**

- Consumes: `contracts.CacheService` via 现有 `GetCache`；`RegisterConversation`
- Produces: `Begin/Transition/End/HasConversation/CancelActive` 在真实 request 上可用

缓存 key：`msg_gateway:bot_conv:{channelID}:{platformUserID}`

记录：

```go
type conversationRecord struct {
    Name     string          `json:"name"`
    State    json.RawMessage `json:"state"`
    TTLNanos int64           `json:"ttl_nanos"`
    ExpireAt int64           `json:"expire_at"`
}
```

无 cache 时 Begin 返回 error（测试注入 memory cache，仿 IGoLibGo 的 `memoryCache`，或写 20 行 map+json 实现 `contracts.CacheService`）。

- [ ] **Step 1: Write failing tests**

用 stub 对话记录 `OnMessage`/`OnCancel` 收到的文本。

- 命令 `Begin("igo.login_auth", state, 5*time.Minute)` 后，下一条 `cookie-text` 进 `OnMessage`，不进其它命令
- 对话中 `/cancel`：先 **不要**静默清占位；`CancelActive` 调 `OnCancel` 再删 key；再发非命令不再进 OnMessage
- 对话中已注册 `/show`：静默清占位（不调 OnCancel），然后 Show.Handle
- 对话中未注册 `/nope`：进 OnMessage（不当未知命令）
- `Transition("igo.checkin_auth", newState)` 后下一条进新对话
- `Begin` 未注册名 → error
- cache miss / 过期 → 当无对话

测试里 `RegisterBuiltin` 一个 cancel stub：`HasConversation` 则 `CancelActive` 并 Reply「已取消」；否则 Reply「当前没有进行中的操作」。

- [ ] **Step 2: Run tests (RED)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -run TestBotRegistry_Conversation -count=1`

Expected: FAIL

- [ ] **Step 3: Implement session + Dispatch 规则**

已绑定且 `isCmd`：

```
if name == "cancel" { // 归一化后
    Handle  // request.CancelActive inside builtin later
    return
}
if lookup ok {
    silent delete occupancy (no OnCancel)
    Handle
    return
}
if occupancy exists { OnMessage; return }
Reply 未知命令
```

已绑定非命令 + occupancy → OnMessage。

`CancelActive`：load record → LookupConversation → OnCancel → Delete key。无占位则 nil。

`Begin`：LookupConversation 必须存在；json marshal state；Set cache ttl。

无 cache：Begin 返回 `fmt.Errorf("bot conversation cache unavailable")`。

- [ ] **Step 4: Run tests (GREEN)**

Run: `cd backend && go test ./plugins/domain/msg_gateway/service/ -count=1`

Expected: PASS（含既有 pairing/dispatch/runner）

- [ ] **Step 5: Commit**

```bash
git add backend/plugins/domain/msg_gateway/service/bot_session.go \
        backend/plugins/domain/msg_gateway/service/bot_inbound.go \
        backend/plugins/domain/msg_gateway/service/bot_inbound_test.go
git commit -m "feat(msg_gateway): add bot conversation occupancy routing"
```

---

### Task 5: Builtin `/help` `/start` `/cancel`

**Files:**

- Create: `backend/plugins/domain/msg_gateway/bot/help.go`
- Create: `backend/plugins/domain/msg_gateway/bot/cancel.go`
- Create: `backend/plugins/domain/msg_gateway/bot/builtins.go`
- Create: `backend/plugins/domain/msg_gateway/bot/help_test.go`

**Interfaces:**

- Consumes: `service.CommandMeta` / `List()`；`BotCommandRequest.CancelActive`
- Produces: `bot.RegisterBuiltins(reg *service.BotRegistry) error`

`help.go`：

```go
type HelpCommand struct{ List func() []service.CommandMeta }

func (*HelpCommand) Name() string        { return consts.CommandHelp }
func (*HelpCommand) Aliases() []string   { return []string{consts.CommandStart} }
func (*HelpCommand) Description() string { return "查看全部可用指令" }
func (*HelpCommand) Usage() string       { return "/help" }
```

Handle 按 Builtin 分成「系统指令」「业务指令」两段。系统行：`/start, /help` 合成一行（看到 alias `start` 时并进 help 那行）。无业务指令则省略业务段。

`cancel.go`：`AllowUnbound=true`。Handle：`HasConversation` → `CancelActive` + Reply「已取消当前操作」；否则 Reply「当前没有进行中的操作」。未绑定不得发配对码。

`builtins.go`：

```go
func RegisterBuiltins(reg *service.BotRegistry) error {
    help := &HelpCommand{List: reg.List}
    if err := reg.RegisterBuiltin(help, true); err != nil { return err }
    if err := reg.RegisterBuiltin(&CancelCommand{}, true); err != nil { return err }
    return nil
}
```

（`/me` 下一任务再挂上，避免本任务范围膨胀。`RegisterBuiltins` 本任务只注册 help+cancel；Task 6 改它加 me。）

- [ ] **Step 1: Write failing tests** in `bot/help_test.go`

- 构造 Registry，业务 stub `show`，注册 builtins，`Dispatch` `/help` 与 `/start`，两段文本相等，且含 `/show` 与 `/me` 尚未出现时不含 me
- `/help` 含「系统指令」和「业务指令」
- 未绑定 `/cancel` 回复「当前没有进行中的操作」，不含「配对码」

需要 sqlite 绑定表 + memory cache 的辅助函数，可放 `service` 测试 helper 或 bot 测试里重复最小建库（允许 30 行 helper）。

更简单：help 单测直接调 `HelpCommand.Handle` + fake request 记录 Reply，不断 Dispatch。cancel 同样。Dispatch 集成放到 Task 6。

本期 `help_test.go`：

```go
type fakeReq struct{ replies []string /* implement contracts.BotCommandRequest */ }

func TestHelpCommand_ListsBuiltinAndBusiness(t *testing.T) {
    cmd := &HelpCommand{List: func() []service.CommandMeta {
        return []service.CommandMeta{
            {Name: "help", Aliases: []string{"start"}, Description: "查看全部可用指令", Usage: "/help", Builtin: true},
            {Name: "cancel", Description: "取消当前进行中的操作", Usage: "/cancel", Builtin: true},
            {Name: "show", Description: "列出配置", Usage: "/show", Builtin: false},
        }
    }}
    req := &fakeReq{}
    require.NoError(t, cmd.Handle(context.Background(), req))
    require.Len(t, req.replies, 1)
    assert.Contains(t, req.replies[0], "/start, /help")
    assert.Contains(t, req.replies[0], "/show")
    assert.Contains(t, req.replies[0], "系统指令")
    assert.Contains(t, req.replies[0], "业务指令")
}
```

`/start` 与 `/help` 同类型：`HelpCommand.Aliases` 含 `start`，`Name()==help`。

cancel 单测：`HasConversation false` → 固定文案；true 时 `CancelActive` 被调用一次。

- [ ] **Step 2: RED**

Run: `cd backend && go test ./plugins/domain/msg_gateway/bot/ -count=1`

Expected: FAIL — package bot 不存在

- [ ] **Step 3: Implement help/cancel/builtins**

- [ ] **Step 4: GREEN**

Run: `cd backend && go test ./plugins/domain/msg_gateway/bot/ ./plugins/domain/msg_gateway/service/ -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/plugins/domain/msg_gateway/bot/
git commit -m "feat(msg_gateway): add builtin /help and /cancel commands"
```

---

### Task 6: Builtin `/me` + plugin/runner 接线

**Files:**

- Create: `backend/plugins/domain/msg_gateway/bot/me.go`
- Create: `backend/plugins/domain/msg_gateway/bot/me_test.go`
- Modify: `backend/plugins/domain/msg_gateway/bot/builtins.go` — Register me
- Modify: `backend/plugins/domain/msg_gateway/plugin.go`
- Modify: `backend/plugins/domain/msg_gateway/service/bot_runner.go`
- Modify: `backend/plugins/domain/msg_gateway/plugin_registry_test.go`
- Modify: `backend/plugins/domain/msg_gateway/service/bot_inbound.go` if Dispatch 需要进程级 registry

**Interfaces:**

- Runner `onInbound` 改为 `reg.Dispatch(ctx, msg, r.SendText)`
- `plugin.Apply`：`reg := service.NewBotRegistry()`；`bot.RegisterBuiltins(reg)`；`ctx.Provide[contracts.BotCommandRegistry](reg)`；`service.SetBotRegistry(reg)`（或 Runner 持有指针）
- `/me` 未绑定：调用 `service.ReplyPairingCode` 等价文案（或注入 `PairingFn`）。已绑定：UserID、username、nickname（`GetUserService` 可 nil 则只写 ID）、当前频道名/类型/平台 ID、其它绑定列表。禁止 email/phone。

进程级 registry：

```go
var (
    botRegMu sync.RWMutex
    botReg   *BotRegistry
)
func SetBotRegistry(r *BotRegistry) { ... }
func GetBotRegistry() *BotRegistry { ... }
```

`Apply` 的 `OnDispose` 置 nil。Runner：

```go
reg := GetBotRegistry()
if reg == nil { return HandleInboundMessage(...) } // 测试无 registry 时仍可配对
return reg.Dispatch(inboundCtx, msg, r.SendText)
```

测试里若只测 runner 生命周期、不测命令，保持现有 stub 通道即可。

- [ ] **Step 1: Write failing tests**

`me_test.go`：fakeReq + 注入 listBindings/getUser。

- UserID=0 → Reply 含「尚未绑定」或配对引导（测试里 MeCommand 依赖 `UnboundReply func(...) error`，未绑定走它）
- UserID!=0 → 含 username、频道类型，**不含** `@` email 模式、不含「cookie」

`plugin_registry_test.go` 增加：

```go
botReg, err := ctx.Inject[contracts.BotCommandRegistry]()
require.NoError(t, err)
require.NotNil(t, botReg)
```

`plugin.go` 用 `msg_gateway.New(msg_gateway.WithAutoStartRunner(false))` 若现有测试已启动 runner 且通过，不要无故改测试选项；只加 Inject 断言。

Dispatch 集成（service 测试）：RegisterBuiltins 后未绑定 `/me` 含配对码；已绑定 `/me` 含用户 ID；`/start` 与 `/help` 文本相等。

- [ ] **Step 2: RED**

Run: `cd backend && go test ./plugins/domain/msg_gateway/... -count=1`

Expected: FAIL on new me/plugin inject tests

- [ ] **Step 3: Implement me + wire plugin/runner**

`MeCommand` 可以留在 `bot` 包并调用 `dao.ListBindingsByUser`、`dao.GetMessageChannel`、`service.GetUserService`、`service.ReplyPairingCode`。这是网关内部，允许。

`RegisterBuiltins` 在 help/cancel 之后：

```go
if err := reg.RegisterBuiltin(&MeCommand{}, true); err != nil { return err }
```

plugin.go 在 channel factory Register 之后、Task Register 之前插入 registry 组装。`RegisterBuiltins` error 则 `return err`。

- [ ] **Step 4: GREEN + 回归**

Run: `cd backend && go test ./plugins/domain/msg_gateway/... ./core/contracts/ -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/plugins/domain/msg_gateway/bot/ \
        backend/plugins/domain/msg_gateway/plugin.go \
        backend/plugins/domain/msg_gateway/plugin_registry_test.go \
        backend/plugins/domain/msg_gateway/service/bot_runner.go \
        backend/plugins/domain/msg_gateway/service/bot_inbound.go \
        backend/plugins/domain/msg_gateway/service/bot_inbound_test.go
git commit -m "feat(msg_gateway): add /me and wire command dispatcher into bot runner"
```

---

### Task 7: Skill `wv-new-bot-command` + 索引

**Files:**

- Create: `.agents/skills/wv-new-bot-command/SKILL.md`
- Modify: `AGENTS.md` — 在 `wv-push-notification` 下行插入 `wv-new-bot-command`
- Modify: `.agents/skills/README.md` — 「10 个」改为「11 个」；自研表增加一行

**Interfaces:** 无代码接口。技能必须写清：何时用、契约、文件落点、Register、对话、自带保留名、禁止事项、测试、`make format`/`go test`。

- [ ] **Step 1: Write SKILL.md**（完整可触发，不要链到「见 spec」代替正文）

frontmatter：

```yaml
---
name: "wv-new-bot-command"
description: "Wavelet 项目专用：当业务插件要注册 Bot 斜杠命令、多轮对话，或修改 msg_gateway 命令分发/自带指令时必须使用。指导 contracts.BotCommandRegistry、bot/ 包落点、/help 元数据与启动查重。"
metadata:
  origin: Wavelet
---
```

正文至少含：

1. 网关 vs 业务插件边界（禁止 import msg_gateway 内部包）
2. `bot/` 一命令一文件
3. 实现 `BotCommand` 必填 Name/Description/Usage
4. `ctx.Bind(func(reg contracts.BotCommandRegistry) { if err := reg.Register(...); err != nil { panic(err) } })`
5. 对话 `插件名.xxx` + `Begin`/`Transition`/`End`
6. 保留名 `help` `start` `me` `cancel`；`/start`≡`/help`；身份用 `/me`
7. 禁止自己解析 `/`、自己查绑定表、自己接 Telegram、本地 map 会话
8. 测试 mock `BotCommandRequest`
9. 门禁：`cd backend && go test ./plugins/domain/msg_gateway/...` 与业务插件 `./path/to/plugin/bot/`

示例用虚构 `order` 插件 `/orderstatus`，不要写 igo/TraceInt。

- [ ] **Step 2: Update AGENTS.md table**

行：

`| \`wv-new-bot-command\` | 业务插件注册 Bot 斜杠命令与多轮对话、\`contracts.BotCommandRegistry\`、自带 /help /me /cancel |`

- [ ] **Step 3: Update `.agents/skills/README.md`**

「Wavelet 自研/内核业务」数量 10→11；表新增 `wv-new-bot-command` 一行，职责与 AGENTS 一致。

- [ ] **Step 4: 人工扫一遍技能** — description 含触发词「Bot 命令」「斜杠」「对话」；无 TBD。

- [ ] **Step 5: Commit**

```bash
git add .agents/skills/wv-new-bot-command/SKILL.md AGENTS.md .agents/skills/README.md
git commit -m "docs(skills): add wv-new-bot-command development guide"
```

---

### Task 8: Note 落地 + 质量门禁

**Files:**

- Move: `.agents/notes/proposed/architecture/2026-09-15-bot-command-registry.md` → `.agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md`
- Modify: 该 Note 的 Status、`## Proposal`→`## Decision`（现在时）、`## Acceptance criteria`/`## Risks`→`## Consequences`
- Modify: `backend/core/contracts/bot.go` 与 `msg_gateway/plugin.go` 入口各一行 `// Note: ... 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md`

- [ ] **Step 1: 改写 Note 为 implemented**

Decision 现在时：契约在 `contracts/bot.go`；网关 Provide 注册表；自带指令在 `msg_gateway/bot/`；对话占位走 CacheService。

Consequences：代价 — 业务要注册对话才能接非命令文本；`Bind` 冲突靠 panic（本期业务尚未接入）。收益 — 网关不再含业务命令。IGoLibGo 迁移仍是后续，Note 里写一句「下游尚未迁命令」。

Alternatives 保留。

- [ ] **Step 2: 反向注释**

`contracts/bot.go` 文件包注释附近一行 Note。`plugin.go` Provide 注册表处一行 Note。

- [ ] **Step 3: format + tests**

```bash
make format
cd backend && go test ./core/contracts/ ./plugins/domain/msg_gateway/... -count=1
```

若 `make code-check` 可用且不因无关施工失败，跑一遍；失败则只跑本包测试并在回复里说明跳过原因。

- [ ] **Step 4: Commit（Note + 注释必须与代码同批）**

```bash
git add .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md \
        backend/core/contracts/bot.go \
        backend/plugins/domain/msg_gateway/plugin.go
# 若 proposed 文件需 git rm
git commit -m "docs(msg_gateway): record bot command registry architecture"
```

不要把其它脏文件加进来。

---

## Spec coverage (self-review)

| Spec 节 | Task |
| :--- | :--- |
| contracts 接口 | 1（含 CancelActive 补丁） |
| 注册校验 / 启动查重 | 2 |
| 分发、配对、未知命令、已绑定不再成功文案 | 3 |
| 对话占位、cancel vs 其它命令打断 | 4 |
| `/help`≡`/start`、`/cancel` AllowUnbound | 5 |
| `/me`、Provide、Runner | 6 |
| 命令开发指南 skill | 7 |
| Note 原子提交 | 8 |
| IGoLibGo 迁移 | **明确不做** |

## Follow-up (not this plan)

IGoLibGo：`git fetch wavelet` → `git merge wavelet/main` → `igo/bot/{show,run,login_auth,checkin_auth}.go` → 删网关上的 `bot_command_handler.go`。另开计划。
