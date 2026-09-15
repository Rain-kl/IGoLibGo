---
name: "wv-new-bot-command"
description: "Wavelet 项目专用：当业务插件要注册 Bot 命令、斜杠命令、多轮对话，或修改 msg_gateway 命令分发/自带指令时必须使用。指导 contracts.BotCommandRegistry、bot/ 包落点、/help 元数据与启动查重。"
metadata:
  origin: Wavelet
---

# Bot 斜杠命令与多轮对话开发规范 (Cordis 插件化架构)

本技能是 Wavelet 在 Cordis 微内核下，业务插件注册 Bot 斜杠命令、多轮对话，以及修改 `msg_gateway` 命令分发 / 自带指令时的唯一指导。契约以落地代码 `backend/core/contracts/bot.go` 为准，禁止发明额外 API。

**何时使用**

- 业务插件要新增 / 修改 Bot 斜杠命令（如 `/orderstatus`）
- 业务插件要做多轮对话（补录、确认、向导）
- 修改 `msg_gateway` 入站分发、对话占位或自带 `/help` `/me` `/cancel`

开始前阅读仓库根目录 [AGENTS.md](../../../AGENTS.md)，遵守插件自包含与「只面向 contracts」防线。

---

## 1. 网关 vs 业务插件边界

`msg_gateway` 只做消息底层：通道适配、Runner、身份绑定、斜杠解析、命令/对话注册表、对话占位、经已连接通道 `Reply`。业务语义留在业务插件自己的 `bot/` 包。

```text
Telegram / QQ adapter
        │  标准化 contracts.BotInbound
        ▼
msg_gateway dispatcher
        │
        ├─ 未绑定 ─ AllowUnbound 自带指令 (/help /start /me /cancel) → Handle
        │           其它文本 / 业务命令 ──────────────────────────→ 配对码
        │
        └─ 已绑定 ─ 已注册命令 ─ /cancel：保留占位，由 cancel.Handle 结束对话
                    │            其它命令：静默清占位，再 Handle
                    ├─ 非命令 + 有对话 ────────────────────→ Conversation.OnMessage
                    └─ 非命令 + 无对话 ────────────────────→ 丢弃
```

| 职责 | 归属 | 公开接入 |
| :--- | :--- | :--- |
| 命令表、斜杠解析、对话占位、配对、通道 Reply | `plugins/domain/msg_gateway` | `contracts.BotCommandRegistry`（`Provide`） |
| 自带 `/help` `/start` `/me` `/cancel` | `msg_gateway/bot/` | 内部 `RegisterBuiltin`，**不在**公开接口上 |
| 业务斜杠命令与多轮对话 | 各插件自己的 `bot/` | `Register` / `RegisterConversation` |

依赖方向：

```text
业务插件 bot/  ──implements──►  contracts.BotCommand / BotConversation
msg_gateway/bot  ──implements──►  同上（仅自带指令）
msg_gateway.Apply  ──Provide──►  contracts.BotCommandRegistry
业务插件 Apply  ──Bind Register──►  BotCommandRegistry
```

**严禁**业务插件 `import "Wavelet/plugins/domain/msg_gateway/..."`（含 `bot/`、`service/`、`dao/`、`model/`）。只面向 `Wavelet/core/contracts`。`RegisterBuiltin` 业务插件调不到。

`msg_gateway` 未加载时 `Bind` 不会触发，业务 HTTP API 仍可启动（软依赖）。产品若要把 Bot 当硬依赖，再在该插件 `Inject()` 里声明 `contracts.BotCommandRegistry`。

---

## 2. 文件落点：`bot/` 一命令一文件

在**自己的插件**下建 `bot/` 物理子包。不要改 `msg_gateway`，不要把命令平铺在插件根目录或 `controller/`。

```text
backend/plugins/domain/order/          # 虚构示例插件
├── plugin.go                          # Apply 里 Bind 注册表
├── bot/
│   ├── orderstatus.go                 # /orderstatus 命令（一命令一文件）
│   ├── orderstatus_test.go
│   ├── confirm_cancel.go              # 对话 order.confirm_cancel（一对话一文件）
│   └── confirm_cancel_test.go
└── service/
    └── order.go                       # Handle 只调 Service，不写 SQL
```

下游定制插件同理：`backend/downstream/plugins/<name>/bot/`。

修改网关自带指令或分发时，落点固定为：

```text
backend/plugins/domain/msg_gateway/
  bot/           # help.go / me.go / cancel.go / builtins.go
  service/       # bot_registry.go、bot_inbound.go、bot_session.go
  plugin.go      # NewBotRegistry → RegisterBuiltins → Provide → SetBotRegistry
```

`msg_gateway/service` 不 import `msg_gateway/bot`；由 `plugin.go` 组装，避免循环依赖。

---

## 3. 实现 `contracts.BotCommand`

落地接口（`backend/core/contracts/bot.go`）：

```go
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
```

`Name` / `Description` / `Usage` **trim 后不得为空**，否则 `Register` 返回 error、进程起不来。`Name` 与别名禁止含 `/` 或空白；`Name()` 不要带前导 `/`（用户输入才是 `/orderstatus`）。`Aliases()` 无别名时返回 `nil`。

`/help` 的业务行列出注册表里的 `Name` / `Aliases` / `Description`，必须填写用户可读的元数据。

```go
package bot

import (
	"context"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/order/service"
)

type OrderStatusCommand struct {
	svc *service.OrderService
}

var _ contracts.BotCommand = (*OrderStatusCommand)(nil)

func (*OrderStatusCommand) Name() string        { return "orderstatus" }
func (*OrderStatusCommand) Aliases() []string   { return nil }
func (*OrderStatusCommand) Description() string { return "查询订单状态" }
func (*OrderStatusCommand) Usage() string       { return "/orderstatus <订单号>" }

func (c *OrderStatusCommand) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
	args := req.Args()
	if len(args) == 0 {
		return req.Reply("请提供订单号，用法：/orderstatus <订单号>")
	}
	// req.UserID() 已是绑定后的 Wavelet 用户；未绑定进不了业务 Handle
	status, err := c.svc.StatusForUser(ctx, req.UserID(), args[0])
	if err != nil {
		return req.Reply("未找到该订单，请确认订单号后重试")
	}
	return req.Reply("订单 " + args[0] + " 当前状态：" + status)
}
```

要点：

- 用户文案只走 `req.Reply`。`Handle` 返回的 `error` 仅打日志，网关**不会**把 `err.Error()` 回进聊天。
- 禁止在 `Reply` 里输出内部 error / SQL / panic 字符串。
- `HasConversation` / `CancelActive` 供自带 `/cancel` 使用；业务命令需要多轮时调 `Begin`，不要自己实现取消指令。
- `Begin(name, state, ttl)` 的 `name` 必须已 `RegisterConversation`，否则返回 error。同一身份同时只有一个对话，后 `Begin` 覆盖旧占位。

业务 `Register` 一律要求已绑定。未绑定用户发送 `/orderstatus` 不会进 `Handle`，走配对码。

---

## 4. 在 `Apply` 中 `Register`（失败必须 panic）

`ctx.Bind` 回调签名是 `func(reg T)`，不能返回 error。`Register` 失败时必须 `panic(err)`，让进程启动中止。禁止 `_ = reg.Register(...)` 或只打日志。

```go
func (p *Plugin) Apply(ctx *core.Context) error {
	ctx.Bind(func(reg contracts.BotCommandRegistry) {
		if err := reg.Register(&bot.OrderStatusCommand{svc: p.svc}); err != nil {
			panic(err)
		}
		if err := reg.RegisterConversation(&bot.ConfirmCancel{svc: p.svc}); err != nil {
			panic(err)
		}
	})
	return nil
}
```

只允许在插件 `Apply` 里注册。禁止在 goroutine、HTTP Handler、命令 `Handle` 或请求路径里再 `Register`。

命令名 + 所有别名与已注册项同一命名空间，**大小写不敏感**全局查重（含自带指令）。冲突、空 Description/Usage、Name 含 `/` 均返回 error → `panic` → 启动失败。

---

## 5. 多轮对话：`插件名.xxx` + `Begin` / `Transition` / `End`

落地接口：

```go
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
```

对话 `Name()` 必须含 `.`，约定 `插件名.xxx`（如 `order.confirm_cancel`）。不含 `.` 或与已注册对话重名（大小写不敏感）则 `RegisterConversation` 失败。

```go
package bot

import (
	"context"
	"strings"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/order/service"
)

type confirmCancelState struct {
	OrderID string `json:"order_id"`
}

type ConfirmCancel struct {
	svc *service.OrderService
}

var _ contracts.BotConversation = (*ConfirmCancel)(nil)

func (*ConfirmCancel) Name() string { return "order.confirm_cancel" }

func (c *ConfirmCancel) OnMessage(ctx context.Context, req contracts.BotConversationRequest) error {
	var st confirmCancelState
	if err := req.State(&st); err != nil {
		if endErr := req.End(); endErr != nil {
			return endErr
		}
		return req.Reply("会话已失效，请重新发起")
	}
	if strings.TrimSpace(req.Text()) != "确认" {
		return req.Reply("请回复「确认」以取消订单，或发送 /cancel 放弃")
	}
	if err := c.svc.CancelForUser(ctx, req.UserID(), st.OrderID); err != nil {
		return req.Reply("取消失败，请稍后重试")
	}
	if err := req.End(); err != nil {
		return err
	}
	return req.Reply("订单 " + st.OrderID + " 已取消")
}

func (*ConfirmCancel) OnCancel(_ context.Context, _ contracts.BotConversationRequest) error {
	return nil // /cancel 随后会 Reply「已取消当前操作」；此处只做业务清理
}
```

命令侧开启对话：

```go
if err := req.Reply("将取消订单 " + orderID + "。请回复「确认」继续，或发送 /cancel 放弃"); err != nil {
	return err
}
return req.Begin("order.confirm_cancel", confirmCancelState{OrderID: orderID}, 5*time.Minute)
```

| 方法 | 行为 |
| :--- | :--- |
| `Begin(name, state, ttl)` | 占位 key 为 `(channelID, platformUserID)`，经 `CacheService`；`name` 必须已注册 |
| `State` / `SetState` | state 对网关透明（JSON）；业务自己定义结构体 |
| `Transition(name, state)` | 换对话名与 state，TTL 刷新为该次 `Begin` 记下的时长 |
| `End()` | 删除占位 |
| `/cancel` | 网关调用 `OnCancel` 再 `End`；业务**不要**自己注册 `cancel` |

过期只靠缓存 TTL，第一期没有 `OnExpire`。不要用本地 `map` 存会话。

---

## 6. 自带保留名

| 指令 | 别名 | 未绑定 | 行为 |
| :--- | :--- | :--- | :--- |
| `help` | `start` | 可进入 | `/start` ≡ `/help`：按注册表拼「系统指令」+「业务指令」 |
| `me` | 无 | 可进入 | **身份**用 `/me`（绑定状态与 Wavelet 用户信息，不含业务数据） |
| `cancel` | 无 | 可进入 | 有对话则 `OnCancel`+`End`；无对话回复「当前没有进行中的操作」（不改发配对码） |

保留名：`help`、`start`、`me`、`cancel`。业务 `Register` 不得占用。自带指令在 `msg_gateway.Apply` 里先 `RegisterBuiltin`，后注册者因全局查重失败。

不要另做「欢迎语」或「绑定成功」自动回复；已绑定用户静默等待命令，身份查询走 `/me`。

---

## 7. 禁止事项

- import `Wavelet/plugins/domain/msg_gateway/...`
- 自己解析 `/`、自己按 `@BotName` 剥命令、自己维护命令表
- 自己查 `w_message_bindings` / 绑定表判断身份（用 `req.UserID()` / `req.Inbound()`）
- 自己接 Telegram / QQ，或为一次 Reply 再 `Connect` 新适配器
- 用本地 `map` / 自研 Cache key 存会话（必须走 `Begin` / `Transition` / `End`）
- 注册保留名 `help` `start` `me` `cancel`
- 在 `Reply` 输出内部 error 字符串
- 在请求处理或 goroutine 里 `Register`
- 把业务命令写进 `msg_gateway/bot/`

---

## 8. 测试：mock `BotCommandRequest`

为 `contracts.BotCommandRequest` / `BotConversationRequest` 写 fake，断言 `Reply` / `Begin` / `Transition` / `End`。表驱动覆盖成功、缺参、需进入对话。不启动 Telegram，不 import `msg_gateway` 内部包。

```go
package bot

import (
	"context"
	"testing"
	"time"

	"Wavelet/core/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReq struct {
	userID  uint64
	args    []string
	replies []string
	begun   string
}

var _ contracts.BotCommandRequest = (*fakeReq)(nil)

func (r *fakeReq) UserID() uint64                { return r.userID }
func (r *fakeReq) Args() []string                { return r.args }
func (r *fakeReq) Text() string                  { return "" }
func (r *fakeReq) Inbound() contracts.BotInbound { return contracts.BotInbound{} }
func (r *fakeReq) Reply(text string) error {
	r.replies = append(r.replies, text)
	return nil
}
func (r *fakeReq) Begin(conversation string, _ any, _ time.Duration) error {
	r.begun = conversation
	return nil
}
func (r *fakeReq) HasConversation() bool { return r.begun != "" }
func (r *fakeReq) CancelActive() error   { return nil }

func TestOrderStatusCommand_RequiresOrderID(t *testing.T) {
	cmd := &OrderStatusCommand{}
	req := &fakeReq{userID: 7}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.NotEmpty(t, req.replies)
	assert.Contains(t, req.replies[0], "/orderstatus")
}
```

fake **必须实现接口全部方法**（含 `HasConversation`、`CancelActive`），否则无法通过 `var _ contracts.BotCommandRequest` 编译。

临时目录用 `t.TempDir()`，禁止硬编码相对路径。

---

## 9. 质量门禁

完成命令 / 对话 / 分发改动后必须运行：

```bash
make format
cd backend && go test ./plugins/domain/msg_gateway/... -count=1
cd backend && go test ./path/to/plugin/bot/ -count=1   # 换成实际业务插件 bot 包
```

改了契约时追加 `./core/contracts/`。涉及网关自带指令时覆盖 `./plugins/domain/msg_gateway/bot/`。项目级收尾再跑 `make code-check`。
