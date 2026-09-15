# 一条龙自动化 (All-in-One Automation) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建多账号卡片式「一条龙自动化 (All-in-One Automation)」子系统，包含卡片配置管理、全流程执行引擎（凭据预检 $\rightarrow$ 座位可用性检查 $\rightarrow$ 自动占座 $\rightarrow$ 自动签到）、消息网关 Bot 交互指令与多轮会话状态机，以及前端卡片控制台页面。

**Architecture:** 
- 后端在 `backend/igo-lib/plugins/igo` 自包含实现 `w_igo_pipeline_configs` 存储、DAO、Service、Controller 与 Pipeline 执行流；
- 消息网关在 `backend/plugins/domain/msg_gateway` 注册指令路由与基于 `CacheService` 的 5 分钟多轮会话状态机，当凭据过期时回复微信授权链接并挂起会话；
- 前端在 `frontend/app/(main)/pipeline` 构建响应式卡片网格，提供多步骤新增/编辑向导、就地凭据补录弹窗与立即执行进度展示。

**Tech Stack:** Go (Gin, GORM, Goose SQL), Next.js 16 (React 19, TypeScript, TailwindCSS, shadcn/ui, next-intl, lucide-react), Telegram/QQ Bot Gateway.

**Spec:** `docs/superpowers/specs/2026-09-15-all-in-one-automation-design.md`

## Global Constraints

- 下游业务改动严格收敛在 `backend/igo-lib/`、`backend/plugins/domain/msg_gateway` 与 `frontend/`，严禁修改上游核心微内核 `backend/core/`。
- 数据库表结构严禁使用 GORM AutoMigrate，统一编写 Goose SQL 双数据库方言迁移（Postgres 与 SQLite）。
- 前端包管理器严格锁定为 `bun`，代码格式化工具严格锁定为 `Biome` (`bun run format`)，严禁使用 npm/pnpm/Prettier。
- 遵守 Conventional Commits，完成后依次运行 `make code-check`、`make format`、`bun run lint` 与 `bun run format`。

---

### Task 1: 数据库迁移与数据模型 (Migrations & Models)

**Files:**
- Create: `backend/igo-lib/plugins/igo/migrations/postgres/00002_pipeline_configs.sql`
- Create: `backend/igo-lib/plugins/igo/migrations/sqlite/00002_pipeline_configs.sql`
- Create: `backend/igo-lib/plugins/igo/model/entity/pipeline.go`
- Create: `backend/igo-lib/plugins/igo/model/do/pipeline.go`
- Modify: `backend/igo-lib/plugins/igo/consts/consts.go`

**Interfaces:**
- Produces: `entity.PipelineConfig`, `do.PipelineConfigDTO`, `do.CreatePipelineConfigRequest`, `do.UpdatePipelineConfigRequest`, `do.PipelineRunResult`

- [ ] **Step 1: 编写 Goose SQL 迁移文件与常量定义**
  - 在 `consts/consts.go` 中添加 `TablePipelineConfigs = "w_igo_pipeline_configs"`。
  - 在 `migrations/postgres/00002_pipeline_configs.sql` 和 `migrations/sqlite/00002_pipeline_configs.sql` 中编写 `CREATE TABLE w_igo_pipeline_configs` 与索引定义及回滚脚本。

- [ ] **Step 2: 编写 GORM Entity 与 DO 模型**
  - 创建 `model/entity/pipeline.go`：定义 `PipelineConfig` 结构体，映射主键 `ID` (string)、`UserID`、`Name`、`Cookie`、`LibraryID`、`SeatKey`、`AutoCheckin`、`BeaconUUID`、`Latitude`、`Longitude` 等字段。
  - 创建 `model/do/pipeline.go`：定义输入输出 DTO 及脱敏映射方法。

- [ ] **Step 3: 运行测试验证模型与表名映射**
  - 运行 `go test ./backend/igo-lib/plugins/igo/model/... -v` 验证表名与序列化。

- [ ] **Step 4: Commit**
  - `git commit -m "feat(igo): add pipeline configs migration and models"`

---

### Task 2: DAO 数据访问层与测试 (Pipeline DAO)

**Files:**
- Create: `backend/igo-lib/plugins/igo/dao/pipeline.go`
- Create: `backend/igo-lib/plugins/igo/dao/pipeline_test.go`

**Interfaces:**
- Produces:
  - `dao.CreatePipelineConfig(ctx context.Context, row *entity.PipelineConfig) error`
  - `dao.GetPipelineConfig(ctx context.Context, id string) (*entity.PipelineConfig, error)`
  - `dao.ListPipelineConfigsByUser(ctx context.Context, userID uint64) ([]entity.PipelineConfig, error)`
  - `dao.UpdatePipelineConfig(ctx context.Context, row *entity.PipelineConfig) error`
  - `dao.DeletePipelineConfig(ctx context.Context, id string, userID uint64) error`
  - `dao.UpdatePipelineCookie(ctx context.Context, id string, cookie string, exp *time.Time) error`
  - `dao.UpdatePipelineCheckinToken(ctx context.Context, id string, token string, exp *time.Time) error`

- [ ] **Step 1: 编写 DAO 单元测试 `dao/pipeline_test.go`**
  - 测试创建配置、主键重复报错、用户数据隔离、按 ID 查询、更新凭据、删除操作。

- [ ] **Step 2: 实现 `dao/pipeline.go` 业务逻辑**
  - 使用 GORM 执行带事务与 Trace 保护的 DB 操作，严格校验用户属主权限与 ID 合法性。

- [ ] **Step 3: 运行 DAO 测试**
  - 运行 `go test ./backend/igo-lib/plugins/igo/dao -run TestPipeline -v` 确保全部通过。

- [ ] **Step 4: Commit**
  - `git commit -m "feat(igo): implement pipeline dao with unit tests"`

---

### Task 3: 核心服务与执行流引擎 (Pipeline Service)

**Files:**
- Create: `backend/igo-lib/plugins/igo/service/pipeline.go`
- Create: `backend/igo-lib/plugins/igo/service/pipeline_test.go`
- Modify: `backend/igo-lib/plugins/igo/service/service.go`

**Interfaces:**
- Produces:
  - `s.ListPipelineConfigs(ctx context.Context, userID uint64) ([]do.PipelineConfigDTO, error)`
  - `s.GetPipelineConfig(ctx context.Context, userID uint64, id string) (*do.PipelineConfigDTO, error)`
  - `s.CreatePipelineConfig(ctx context.Context, userID uint64, req do.CreatePipelineConfigRequest) (*do.PipelineConfigDTO, error)`
  - `s.UpdatePipelineConfig(ctx context.Context, userID uint64, id string, req do.UpdatePipelineConfigRequest) (*do.PipelineConfigDTO, error)`
  - `s.DeletePipelineConfig(ctx context.Context, userID uint64, id string) error`
  - `s.RunPipeline(ctx context.Context, userID uint64, id string, overrideReq *do.RunPipelineRequest) (*do.PipelineRunResult, error)`
  - `s.HelperVerifySession(ctx context.Context, cookie string) ([]do.LibrarySummary, error)`
  - `s.HelperGetLibraryLayout(ctx context.Context, cookie string, libID int) (*do.LibraryLayoutResponse, error)`
  - `s.HelperVerifyCheckin(ctx context.Context, codeOrToken string) (*do.CheckInDeviceResponse, error)`

- [ ] **Step 1: 编写 Service 测试 `service/pipeline_test.go`**
  - Mock TraceInt 客户端与 Checkin 客户端；
  - 测试正常全流程（可用 $\rightarrow$ 占座 $\rightarrow$ 签到）；
  - 测试座位已被占用拦截中断（`IsOccupied == true`）；
  - 测试登录 Cookie 过期返回需登录鉴权错误；
  - 测试签到 Token 过期返回需签到鉴权错误。

- [ ] **Step 2: 实现 `service/pipeline.go` 核心方法**
  - 实现 CRUD、格式校验（ID 正则 `^[a-zA-Z0-9_-]+$`）、辅助验证方法。
  - 实现 `RunPipeline`：
    1. 查出 `PipelineConfig` 并校验用户属主；
    2. 探测 Cookie 有效性；
    3. 若 `AutoCheckin == true` 探测签到 Token 有效性；
    4. 查阅指定场馆 Layout，定位 `seat_key`，检查 `IsOccupied`。若已占则返回 `SeatOccupiedError`；
    5. 调用 `ReserveSeat` 占座；
    6. 若 `AutoCheckin == true` 调用 `GetCheckInServerTime` 与 `SignCheckIn` 进行蓝牙打卡；
    7. 组装结果并触发领域通知事件。

- [ ] **Step 3: 运行 Service 测试**
  - 运行 `go test ./backend/igo-lib/plugins/igo/service -run TestPipeline -v`。

- [ ] **Step 4: Commit**
  - `git commit -m "feat(igo): implement pipeline execution engine and helper services"`

---

### Task 4: HTTP 控制器与路由注册 (Controller & Routes)

**Files:**
- Create: `backend/igo-lib/plugins/igo/controller/pipeline.go`
- Modify: `backend/igo-lib/plugins/igo/controller/controller.go`
- Modify: `backend/igo-lib/plugins/igo/plugin.go`

**Interfaces:**
- Produces: `/api/v1/igo/pipeline/**` RESTful 接口

- [ ] **Step 1: 实现 `controller/pipeline.go`**
  - 绑定 `GET /pipeline/configs`, `POST /pipeline/configs`, `GET /pipeline/configs/:id`, `PUT /pipeline/configs/:id`, `DELETE /pipeline/configs/:id`, `POST /pipeline/configs/:id/run` 以及辅助接口。
  - 使用 `response.OK`, `response.Created`, `response.NoContent`, `response.AbortBadRequest` 遵循规范。

- [ ] **Step 2: 在 `plugin.go` 中挂载路由组**
  - 在 `ctx.Router().Group(consts.APIPrefix, authMW)` 下添加 `/pipeline` 路由组。

- [ ] **Step 3: 运行 Controller 与 Plugin 集成测试**
  - 运行 `go test ./backend/igo-lib/plugins/igo/... -v`。

- [ ] **Step 4: Commit**
  - `git commit -m "feat(igo): add pipeline http controllers and route bindings"`

---

### Task 5: 消息网关 Bot 指令与多轮会话状态机 (Bot Gateway Commands)

**Files:**
- Create: `backend/plugins/domain/msg_gateway/service/bot_command_handler.go`
- Create: `backend/plugins/domain/msg_gateway/service/bot_command_handler_test.go`
- Modify: `backend/plugins/domain/msg_gateway/service/bot_runner.go`
- Modify: `backend/plugins/domain/msg_gateway/plugin.go`

**Interfaces:**
- Produces: `/help`, `/start`, `/show`, `/run [id]`, `/cancel` 指令处理与多轮鉴权状态机。

- [ ] **Step 1: 编写 Bot 指令处理测试 `bot_command_handler_test.go`**
  - 测试 `/help` 和 `/start` 输出格式；
  - 测试未绑定用户提示；
  - 测试已绑定用户发送 `/show` 显示卡片列表；
  - 测试 `/run [id]` 在凭据有效时直接执行并返回结果；
  - 测试 `/run [id]` 在凭据过期时下发微信授权链接并进入 `waiting_login_auth` 状态；
  - 测试在 `waiting_login_auth` 状态下回复 Cookie/链接后的凭据更新与继续执行或流转到 `waiting_checkin_auth`；
  - 测试 `/cancel` 清除会话。

- [ ] **Step 2: 实现 `service/bot_command_handler.go`**
  - 接入 `contracts.CacheService` 处理会话状态（`igo_bot_session:{channel_id}:{platform_user_id}`，TTL 5分钟）；
  - 实现命令解析与分发器。

- [ ] **Step 3: 在 `msg_gateway` 启动和消息接收入口中接入处理器**
  - 连接 Telegram / QQ 消息入站与命令调度。

- [ ] **Step 4: 运行测试**
  - 运行 `go test ./backend/plugins/domain/msg_gateway/service -run TestBotCommand -v`。

- [ ] **Step 5: Commit**
  - `git commit -m "feat(msg_gateway): implement interactive bot commands for pipeline automation"`

---

### Task 6: 前端 API Service 接入 (Frontend Services)

**Files:**
- Create: `frontend/lib/services/igo/pipeline.ts`
- Modify: `frontend/lib/services/igo/index.ts`
- Modify: `frontend/lib/services/index.ts`

**Interfaces:**
- Produces: `services.igoPipeline.listConfigs()`, `services.igoPipeline.createConfig()`, `services.igoPipeline.updateConfig()`, `services.igoPipeline.deleteConfig()`, `services.igoPipeline.runConfig()`, `services.igoPipeline.verifySession()`, `services.igoPipeline.getLibraryLayout()`, `services.igoPipeline.verifyCheckin()`

- [ ] **Step 1: 编写 `frontend/lib/services/igo/pipeline.ts`**
  - 继承 `BaseService`，定义 TypeScript 类型与 REST 请求方法。

- [ ] **Step 2: 导出服务实例**
  - 在 `frontend/lib/services/igo/index.ts` 和 `frontend/lib/services/index.ts` 注册导出。

- [ ] **Step 3: Commit**
  - `git commit -m "feat(frontend): add igo pipeline api client service"`

---

### Task 7: 前端国际化与侧边栏导航 (i18n & Navigation)

**Files:**
- Modify: `frontend/messages/zh-CN.json`
- Modify: `frontend/messages/en.json`
- Modify: `frontend/components/layout/sidebar.tsx`

- [ ] **Step 1: 在 `zh-CN.json` 和 `en.json` 中添加「自动化」与「一条龙」文案**
  - 在 `igo.nav.pipeline` 中配置「自动化」/「Automation」；
  - 在 `igo.pipeline.*` 中配置卡片、向导表单、执行状态、错误提示等完整中英文对照键。

- [ ] **Step 2: 在 `sidebar.tsx` 中添加导航项**
  - 在 `igoNavItems` 中加入 `{ titleKey: 'pipeline', url: '/pipeline', icon: Workflow }`（或 `Sparkles` / `Layers`）。

- [ ] **Step 3: Commit**
  - `git commit -m "feat(frontend): add pipeline navigation item and i18n dictionaries"`

---

### Task 8: 前端卡片页面与交互组件 (Frontend UI Page & Modals)

**Files:**
- Create: `frontend/app/(main)/pipeline/page.tsx`
- Create: `frontend/app/(main)/pipeline/components/pipeline-card.tsx`
- Create: `frontend/app/(main)/pipeline/components/pipeline-dialog.tsx`
- Create: `frontend/app/(main)/pipeline/components/pipeline-auth-modal.tsx`
- Create: `frontend/app/(main)/pipeline/components/pipeline-result-dialog.tsx`

- [ ] **Step 1: 构建 `pipeline-card.tsx`**
  - 参考 `admin/tasks` 卡片网格，实现渐变边框、配置 ID 徽章、场馆座位概览、自动签到标签、操作下拉菜单与「立即执行」按钮。

- [ ] **Step 2: 构建 `pipeline-dialog.tsx`**
  - 多步骤向导：基本信息（ID & 名称） $\rightarrow$ TraceInt 授权（录入/扫码/解析） $\rightarrow$ 动态拉取场馆与座位排布 $\rightarrow$ 自动签到开关及 Beacon 选取 $\rightarrow$ 校验并保存。

- [ ] **Step 3: 构建 `pipeline-auth-modal.tsx` 与 `pipeline-result-dialog.tsx`**
  - 当执行返回需重新授权时弹出快速补录对话框；
  - 执行完成后弹出步骤详情与执行结果浮层。

- [ ] **Step 4: 组装 `page.tsx`**
  - 页面级全宽容器、标题栏、操作按钮、空状态与骨架加载。

- [ ] **Step 5: 验证与格式化**
  - 运行 `bun run format` 与 `bun run lint`。

- [ ] **Step 6: Commit**
  - `git commit -m "feat(frontend): implement pipeline automation cards and modals"`

---

### Task 9: 全链路验证与门禁检查 (Verification & Code Quality)

**Files:** All modified files

- [ ] **Step 1: 运行后端全部测试与检查**
  - 运行 `go test ./backend/igo-lib/plugins/igo/... -v`
  - 运行 `go test ./backend/plugins/domain/msg_gateway/... -v`
  - 运行 `make code-check` 与 `make format`。

- [ ] **Step 2: 运行前端构建与代码检查**
  - 运行 `cd frontend && bun run build` 确保无 TypeScript 或 JSX 编译错误。

- [ ] **Step 3: Commit & Update Walkthrough**
  - 撰写总结并生成 walkthrough 记录。
