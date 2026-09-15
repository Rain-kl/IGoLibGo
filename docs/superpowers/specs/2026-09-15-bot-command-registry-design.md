# msg_gateway Bot 命令接入契约

Status: proposed
Date: 2026-09-15

`msg_gateway` 只做消息底层：通道、配对、命令表、对话占位、回复。业务命令在各自插件的 `bot/` 包里实现 `contracts.BotCommand` / `BotConversation`，启动时注册。网关不出现一条龙、TraceInt 等业务词。

## 1. Problem

当前入站路径把两件不该在一起的事焊死了：

1. **上游 Wavelet** 的 `HandleInboundMessage` 只做配对：未绑定发配对码，已绑定回「账号已成功绑定」。没有命令表。
2. **下游 IGoLibGo** 在 `msg_gateway/service/bot_command_handler.go` 里直接 import `igo` 的 dao/service/traceint，把 `/show`、`/run` 和微信补录写进框架插件。这违反「插件不得跨包 import 对方内部实现」。
3. Runner 把 inbound 固定交给配对函数，IGo handler 实际上进不了已连接通道，命令路径是断的。

目标：其它插件能在自己的包里声明并注册指令；`msg_gateway` 对外只暴露接入契约。

## 2. Goals / Non-goals

**Goals**

- 网关提供：通道适配、Runner、身份绑定、斜杠命令解析、命令/对话注册表、经已连接通道回复。
- 自带指令收拢在 `msg_gateway/bot/`：`/help`（别名 `/start`）、`/me`、`/cancel`。
- 业务指令在业务插件的 `bot/` 包实现，必须声明 Name、Description、Usage。
- `/help` 列出全部已注册指令（自带 + 业务）。
- 启动时命令名/别名全局查重，冲突则 `Apply` 失败、进程退出。
- 多轮输入（如补录）用对话占位：网关管「下一条交给谁」，插件管「这段文本是什么」。

**Non-goals（第一期不做）**

- 群聊、按钮、权限角色、引号参数、对话过期回调、Telegram `setMyCommands`。
- 业务插件直连 Telegram/QQ 或 import `msg_gateway` 内部包。

## 3. Architecture

```text
Telegram / QQ adapter
        │  标准化 BotInbound
        ▼
msg_gateway dispatcher
        │
        ├─ 未绑定 ─ AllowUnbound 自带指令 (/help /start /me /cancel) → Handle
        │           其它 ──────────────────────────────────────────→ 配对码
        │
        └─ 已绑定 ─ 已注册命令 ─ /cancel：保留占位，由 cancel.Handle 结束对话
                    │            其它命令：静默清占位，再 Handle
                    ├─ 非命令 + 有对话 ────────────────────→ Conversation.OnMessage
                    └─ 非命令 + 无对话 ────────────────────→ 丢弃
```

依赖方向：

```text
igo/bot  ──implements──►  contracts.BotCommand / BotConversation
msg_gateway/bot  ──implements──►  同上（自带）
msg_gateway/service  ──Provide──►  contracts.BotCommandRegistry
业务插件 Apply  ──Bind Register──►  BotCommandRegistry
```

`msg_gateway/service` 不 import `msg_gateway/bot`；由 `plugin.go` 组装，避免循环依赖。业务插件只 import `Wavelet/core/contracts`。

## 4. Builtin vs business commands

| 种类 | 注册 API | 包 | 未绑定 | 保留名 |
| :--- | :--- | :--- | :--- | :--- |
| 自带 | 内部 `RegisterBuiltin` | `plugins/domain/msg_gateway/bot/` | 仅标记 AllowUnbound 的可进入 | `help` `start` `cancel` `me` |
| 业务 | 公开 `Register` | 各插件自己的 `bot/` | 一律要求已绑定 | 不得占用保留名 |

### 4.1 自带指令

| 指令 | 别名 | AllowUnbound | Handle |
| :--- | :--- | :--- | :--- |
| `help` | `start` | 是 | 同一函数：按注册表拼帮助。`/start` 与 `/help` 作用完全一致 |
| `me` | 无 | 是 | 绑定状态 + 身份信息，不含业务数据 |
| `cancel` | 无 | 是 | 有对话则 `OnCancel` + `End`；未绑定或无对话则回复「当前没有进行中的操作」（不要改发配对码） |

`/help` 输出形态（业务行来自注册表，网关不写死）：

```text
系统指令：
/start, /help  - 查看全部可用指令
/me            - 查看绑定状态与个人信息
/cancel        - 取消当前进行中的操作

业务指令：
/show           - 列出一条龙自动化配置
/run <配置ID>   - 立即执行指定配置
```

无业务指令时省略「业务指令」整段。

### 4.2 `/me` 展示

**未绑定：** 提示尚未绑定，附当前频道配对码（`me` 在网关包内，可调用配对能力）。

**已绑定：** 只展示身份层：

- Wavelet 用户 ID、username、nickname（经 `UserService.GetUserByID`；取失败则只展示 ID）
- 当前频道：类型、频道名、平台用户 ID
- 该 Wavelet 用户的其它 Bot 绑定列表（频道名 + 类型）

不展示 email、phone、Cookie、一条龙配置、场馆座位。那些走业务指令。

## 5. Contracts

新文件：`backend/core/contracts/bot.go`。DTO 全部在 contracts，插件不得依赖 `msg_gateway/model/do`。

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
    UserID         uint64 // 已绑定才有；未绑定为 0
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
}

type BotConversation interface {
    Name() string // must be "<plugin>.<id>", e.g. "igo.login_auth"
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

`RegisterBuiltin` **不在**公开接口上，只存在于网关内部实现，业务插件调不到。

### 5.1 注册校验（启动失败）

`Register` / `RegisterBuiltin` 在以下情况返回 error，调用方 `Apply` 必须把 error 向上返回，进程退出：

- `Name`、`Description`、`Usage` 为空（trim 后）
- `Name` 含 `/` 或空白
- 对话 `Name` 不含 `.`（业务对话必须 `插件名.` 前缀）；对话名同样全局唯一（大小写不敏感）
- **全局查重**：大小写不敏感，`Name` + 所有 `Aliases` 与已注册的命令名、别名同一命名空间（自带 + 业务）。自撞别名同样失败

禁止在 goroutine 或请求路径里再 `Register`。只允许插件 `Apply`。

业务 `Register` 若占用 `help`/`start`/`cancel`/`me`，因自带指令在 `msg_gateway.Apply` 里先 `RegisterBuiltin`，后注册者失败。

### 5.2 接线

`msg_gateway.Apply`：创建注册表 → `RegisterBuiltin(help, me, cancel)` → `ctx.Provide[contracts.BotCommandRegistry](reg)`。

业务插件用 `ctx.Bind`（与 `PushRegistry` 相同，不依赖加载顺序）。`Bind` 回调签名是 `func(reg T)`，不能返回 error。因此 `Register` 失败时必须 `panic`（带冲突的命令名），让进程启动中止。禁止吞掉这个 error。

```go
ctx.Bind(func(reg contracts.BotCommandRegistry) {
    if err := reg.Register(&bot.ShowCommand{svc: p.svc}); err != nil {
        panic(err)
    }
})
```

`msg_gateway` 未加载时 `Bind` 不会触发，业务 HTTP API 仍可启动（软依赖）。产品若要把 Bot 当硬依赖，再在该插件 `Inject()` 里声明 `BotCommandRegistry`。

## 6. Dispatcher

命令解析：

1. `strings.TrimSpace`，空文本直接返回。
2. `strings.Fields` 取首段。
3. 必须以 `/` 开头才视为命令。
4. 去掉 `/` 后若含 `@`（Telegram `/run@BotName`），去掉 `@` 及之后。
5. 命令名大小写不敏感。`Args()` 为其余字段。第一期不做引号参数。

分发表：

| 绑定 | 输入 | 行为 |
| :--- | :--- | :--- |
| 未绑定 | AllowUnbound 自带命令（`/help` `/start` `/me` `/cancel`） | `Handle` |
| 未绑定 | 其它任何文本 | 配对码，不进命令表 |
| 已绑定 | 已注册 `/cancel` | **不清占位**，`cancel.Handle` 自己 `OnCancel`+`End` |
| 已绑定 | 其它已注册命令 | 若有对话则**静默清占位**（不调 `OnCancel`），再 `Handle` |
| 已绑定 | 非命令（含未注册 `/xxx`）且有对话 | 只进当前 `OnMessage` |
| 已绑定 | 未知 `/xxx` 且无对话 | 回复「未知命令，发送 /help 查看可用指令」 |
| 已绑定 | 非命令且无对话 | 丢弃，不回 |

回复必须走 Runner 上**已连接**的 `SendText`，禁止为一次 Reply 再 `Connect` 新适配器。

入站分发必须 `recover`，避免单个 Handle panic 打断通道轮询。Handle 返回的 error 只打日志，网关不把 `err.Error()` 回进聊天。用户文案由插件 `Reply`。

## 7. Conversation occupancy

缓存 key：`msg_gateway:bot_conv:{channelID}:{platformUserID}`，经 `contracts.CacheService`，TTL 由 `Begin` 传入。

记录对网关透明：

```go
type conversationRecord struct {
    Name     string          `json:"name"`
    State    json.RawMessage `json:"state"`
    TTLNanos int64           `json:"ttl_nanos"`
    ExpireAt int64           `json:"expire_at"`
}
```

- `Begin(name, state, ttl)`：`name` 必须已 `RegisterConversation`，否则返回 error。覆盖同一身份上的旧占位。
- `Transition(name, state)`：换对话名与 state，TTL 刷新为该次 `Begin` 记下的时长。
- `End()`：删 key。
- 过期只靠缓存 TTL。第一期没有 `OnExpire`（过期时不主动发消息）。

同一身份同时只有一个对话。并发双发不做跨节点锁；后写覆盖。单进程内可不做 extra mutex。

## 8. Package layout

### 8.1 上游 Wavelet（框架）

```text
backend/core/contracts/bot.go          # 契约 + DTO
backend/plugins/domain/msg_gateway/
  bot/
    help.go       # /help + alias /start
    me.go         # /me
    cancel.go     # /cancel
    builtins.go   # RegisterBuiltin 入口
  service/
    bot_registry.go
    bot_dispatch.go     # 解析 + 分发（现有文件职责扩展）
    bot_pairing.go      # 配对仍在这里；已绑定不再写死「绑定成功」回复
    bot_runner.go       # inbound → dispatcher，Reply 走 live SendText
  plugin.go             # 组装：registry + builtins + Provide
```

删除下游误加在网关上的 IGo 命令实现，不把 `igo` import 带进上游。

### 8.2 下游 IGoLibGo（业务）

```text
backend/igo-lib/plugins/igo/
  bot/
    show.go
    run.go
    login_auth.go      # BotConversation "igo.login_auth"
    checkin_auth.go    # "igo.checkin_auth"
  plugin.go            # Bind 注册表，Register 上述命令与对话
```

删除 `msg_gateway/service/bot_command_handler.go` 及其测试。`igo` **never** import `Wavelet/plugins/domain/msg_gateway/...`。

`RunCommand` 发现需要登录补录时：`Reply` 提示 + `Begin("igo.login_auth", state, 5*time.Minute)`。登录完成后若还要签到：`Transition("igo.checkin_auth", state)`。完成则 `End()` 再执行 pipeline。

## 9. Pairing 与 runner 修正

- 未绑定且不是 AllowUnbound 命令：保持现有配对码行为（见 `.agents/notes/implemented/bug-fix/2026-09-15-telegram-bot-start-reply.md`）。
- 已绑定：不再自动回复「账号已成功绑定」。身份查询改走 `/me`。
- Runner 的 `onInbound` 必须进入新 dispatcher，而不是只调用旧 `HandleInboundMessage`。
- `ReplyToInbound` 那种临时 `Connect`/`Disconnect` 的发送路径，命令回复不得使用。

## 10. Testing

**msg_gateway（不含 igo）**

- 解析：大小写、`/run@BotName`、参数拆分。
- 未绑定：`/help` `/start` `/me` `/cancel` 可进入；其它文本走配对。`/cancel` 未绑定不发配对码。
- `/start` 与 `/help` 回复一致。
- `/me` 未绑定含配对码；已绑定含用户与绑定列表、不含业务字段。
- 对话占位：`Begin` 后非命令进 `OnMessage`；`/cancel` 调 `OnCancel`；其它命令静默清占位再 Handle。
- 未知命令；注册冲突（名/别名/大小写/与自带冲突）返回 error。
- Reply 使用注入的 sendFn，不建新通道。

**igo/bot**

- mock `BotCommandRequest` / `BotConversationRequest`，断言 `Reply`/`Begin`/`Transition`/`End`。
- 不启动 Telegram，不 import `msg_gateway` 内部包。

## 11. Command development guide

给业务插件作者。完整例子以 `igo` 为准。

### 11.1 新建文件

在**自己的插件**下建 `bot/` 物理子包，一命令一文件。不要改 `msg_gateway`。

### 11.2 实现命令

```go
package bot

type ShowCommand struct {
    svc *service.Service
}

func (*ShowCommand) Name() string        { return "show" }
func (*ShowCommand) Aliases() []string   { return nil }
func (*ShowCommand) Description() string { return "列出一条龙自动化配置" }
func (*ShowCommand) Usage() string       { return "/show" }

func (c *ShowCommand) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
    // req.UserID() 已是绑定后的 Wavelet 用户
    // 用户文案：req.Reply(...)
    // 需要多轮：req.Begin("igo.login_auth", state, 5*time.Minute)
    return nil
}
```

`Name` / `Description` / `Usage` 任一会导致启动失败。`Name` 不要带 `/`。

### 11.3 在 Apply 注册

```go
func (p *Plugin) Apply(ctx *core.Context) error {
    // ...
    ctx.Bind(func(reg contracts.BotCommandRegistry) {
        mustRegister(reg.Register(&bot.ShowCommand{svc: p.svc}))
        mustRegister(reg.Register(&bot.RunCommand{svc: p.svc}))
        mustRegister(reg.RegisterConversation(&bot.LoginAuth{svc: p.svc}))
        mustRegister(reg.RegisterConversation(&bot.CheckinAuth{svc: p.svc}))
    })
    return nil
}
```

只面向 `contracts`。`msg_gateway` 未加载时 `Bind` 不会触发，命令只是不出现，不应让 igo 的 HTTP API 启动失败——若产品要求 Bot 为强依赖，再在 `Inject()` 里声明 `BotCommandRegistry`。默认：软依赖，有网关才挂上命令。

### 11.4 对话

内部名必须 `插件名.` 前缀。实现 `OnMessage` / `OnCancel`。用 `State`/`SetState`/`Transition`/`End`。不要自己注册 `/cancel`，不要自己用 `map` 存会话（必须走网关占位 + 平台 `CacheService`）。

### 11.5 禁止

- import `Wavelet/plugins/domain/msg_gateway/...`
- 自己解析 `/`、自己查 `w_message_bindings`、自己接 Telegram
- 注册 `help` / `start` / `cancel` / `me`
- 在 `Reply` 里输出内部 error 字符串
- 在请求处理或 goroutine 里 `Register`

### 11.6 测试

为 `BotCommandRequest` 写 fake：记录 `Reply` 文本和 `Begin` 参数。表驱动覆盖成功、缺参、需补录。

## 12. Key decisions

| 决定 | 理由 |
| :--- | :--- |
| 契约放 `core/contracts`，注册表由 msg_gateway Provide | 与 `PushRegistry` 相同，符合「只面向 contracts 编程」 |
| 不做 `ctx.BotCommands()` 微内核扩展点 | 命令是消息域能力，不是内核机制 |
| 不做纯 EventBus 入站广播 | 没有命令表就无法 `/help` 聚合和启动查重 |
| 自带指令也是 `BotCommand`，放 `msg_gateway/bot/` | 分发器只认一张表；`/me` 能调配对是因为住在网关包内 |
| `/start` 是 `/help` 的别名 | 用户要求作用一致，避免两套欢迎文案 |
| `/me` 承载绑定与身份 | `/help` 只列指令；已绑定不再自动刷「绑定成功」 |
| 对话占位是网关原语，state 不透明 | 补录内容在 igo；网关只保证下一条交给注册的 conversation |
| 查重在 Register 时、大小写不敏感、名与别名同一空间 | 启动即失败，避免运行时抢命令 |

## 13. Alternatives considered

- **业务插件自管 session + 非命令 fallback 链：** 每个插件重写占位、TTL、`/cancel`；网关要轮询 `Match()`。否决。占位是基础设施，补录内容才是业务。
- **NoneBot 式阻塞 `session.aget()`：** 占用 goroutine，重启丢失，Go 不适合 replay engine。否决。用 continuation 对象 + 缓存。
- **已绑定自动回复「绑定成功」：** 与 `/me` 重复且打断命令。已绑定静默等待命令。
- **未绑定任意消息都配对、`/help` 也不能进：** `/start` 是 Telegram 默认第一击，必须能看到指令列表；配对改由 `/me` 与「非 AllowUnbound 文本」承担。

## 14. PR plan

分两期。本期只做 Wavelet；IGoLibGo 业务迁移必须等本期 merge 进下游之后另开。

**本期 — Wavelet 框架**

- `contracts/bot.go`：`BotCommand`、`BotConversation`、`BotCommandRegistry`、请求接口
- `msg_gateway` 注册表、入站分发、对话占位
- `msg_gateway/bot/` 自带指令：`/help`（`/start` 同义）、`/me`、`/cancel`
- Runner inbound 改走分发器；已绑定不再自动回复「绑定成功」
- 技能 `.agents/skills/wv-new-bot-command/SKILL.md`，并挂到 `AGENTS.md` 与 skills README
- 网关测试不引用任何 igo 包

**后续（不在本期）— IGoLibGo**

- 下游 `git merge wavelet/main` 后，在 `igo/bot/` 实现 `/show` `/run` 与补录对话
- 删除误放在网关里的 `bot_command_handler.go`

## 15. Open questions

无。下列已在设计讨论中闭合：`/start`≡`/help`、`/me` 字段范围、启动查重失败即退出、对话由网关占位、业务软依赖注册表。
