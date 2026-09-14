# AGENTS.md

## Git 提交规范

每次完成一个功能点开发或修复一个问题后，务必提交 Git commit , 禁止推送远程仓库。
遵循 Conventional Commits：`<type>(<scope>): <subject>`（例：`feat(auth): support email login`）。
若涉及非琐碎改动（技术选型/架构重构/核心改动/踩坑复盘），必须将对应的 Agent Note（`.agents/notes/...`）与代码同批原子提交。
- **改动范围约束**：仅允许提交自己明确修改的文件，严禁提交非本次任务修改的文件（杜绝盲目 `git add .` 或 `git commit -a` 卷入无关变动）。
- **工作区保护**：严禁随意还原、丢弃或覆盖他人或当前工作区正在施工的文件。
- **测试容错与告知**：若工作区正在施工的文件导致全局测试失败，仅验证本次修改相关的测试用例，跳过受无关施工影响的测试并明确告知跳过原因。

## 务必阅读匹配的 Skill

> 完整 Skills 上游仓库索引、作用职责说明与离线拷贝更新指南详见 [.agents/skills/README.md](file:///.agents/skills/README.md)。

| Skill | 何时使用 |
| :--- | :--- |
| `write-notes-like-deepseek` | 涉及技术选型、架构重构、核心接口/行为变更、修复非直觉缺陷/踩坑复盘，或裁撤冗余特性时；开发前必读既有 Note 并立项记录，与代码原子提交 |
| `wv-new-api` | 基于 Cordis 插件开发业务 HTTP API、通过 `ctx.Router()` 声明路由与挂载中间件 |
| `wv-new-async-task` | 基于 Cordis 插件通过 `ctx.Task()` 与 `ctx.Schedule()` 注册 Asynq 异步任务与定时调度 |
| `wv-new-setting` | 基于 Cordis 插件通过 `ctx.Settings()` 声明配置 Schema、绑定 YAML 配置或管理台热加载设置 |
| `wv-database-guide` | 插件自包含 `embed.FS` 独立 Goose SQL 迁移（PG/SQLite 双方言、ClickHouse 分析库）；以及 ClickHouse 批量写入、`pkg/batchwriter` 接入、分析表异步 flush 与背压策略 |
| `wv-cache-framework` | 基于 `ctx.Cache()` 与 `contracts.CacheService` 访问三层缓存（RAM L1 + Redis L2 + Pub/Sub 同步） |
| `wv-logstore` | 日志/分析用途表、`plugins/domain/risk_control/logstore`、切换日志主库、PG/SQLite 回落 |
| `wv-file-upload` | 业务上传文件、Worker 程序化摄取、`upload.Ingest` / `contracts.StorageService`、文件访问与统计 |
| `wv-push-notification` | 系统通知推送事件、统一触发器投递、带消息推送的业务功能 |
| `wv-logging` | 结构化日志与链路追踪（`backend/pkg/logger` 基于 Zap + otelzap + 5000 行环形缓冲区，支持 Admin WebSocket 日志流）及 `contracts.LoggerService` |
| `wv-release-guide` | 根据自上一正式版本 Tag 以来的提交整理 Version Bump 提交信息以触发双语 Release |
| `code-review-skill` | 进行代码审查（Code Review）、PR 评审、代码质量与安全性审查、检查代码坏味道 |
| `shadcn` | 添加、修改或组合 shadcn/ui 组件 |
| `golang-patterns` | 编写或重构 Go 核心代码、设计接口与并发结构、优化内存与 Goroutine 治理时 |
| `golang-testing` | 编写 Go 测试、表驱动测试、Benchmark 基准测试或 Fuzzing 模糊测试时 |
| `api-design` | 设计 RESTful API 路径、HTTP 状态码、分页规范、错误 Envelope 结构时 |
| `hexagonal-architecture` | 领域模型边界设计、Ports & Adapters 契约抽象与依赖反转时 |
| `frontend-patterns` | React 与 Next.js 组件架构、状态拆分与前端开发模式时 |
| `react-performance` | 针对 Next.js / React 进行性能优化、消除重渲染、消除请求瀑布、打包压缩时 |
| `react-patterns` | 抽象复杂 React 组件复用、Compound Components、Hooks 组合时 |
| `react-testing` | 编写 React / Next.js 组件测试、MSW 网络 Mock 与 a11y 断言时 |
| `nextjs-turbopack` | 配置或优化 Next.js Turbopack 增量编译与构建性能时 |
| `design-system` | 声明或规范 Design Tokens、语义色彩变量、组件变体设计时 |
| `accessibility` | 前端无障碍规范（WCAG 2.2 AA）、键盘导航与无障碍语义审查时 |
| `motion-ui` | 在 Next.js / React 页面中添加 UI 动效、页面转场、微交互动画时 |
| `motion-patterns` | 按钮、Modal、Toast、Stagger 等常用 UI 动效标准化实现时 |
| `clickhouse-io` | ClickHouse 分析库 DDL 设计、分区主键优化、批量导入与分析查询优化时 |
| `data-throughput-accelerator` | 大数据摄取、批量写入、异步 ETL 与高吞吐量数据管道设计时 |
| `content-hash-cache-pattern` | 基于内容哈希（SHA-256）实现跨路径与精确缓存控制时 |
| `security-review` | 处理敏感输入、鉴权授权、Secrets 处理、API 端点安全审查（OWASP）时 |
| `security-bounty-hunter` | 排查严重越权、远程代码执行、注入漏洞与高危安全隐患时 |
| `production-audit` | 生产就绪度检查、单点故障排查、探针与高可用容灾配置时 |
| `docker-patterns` | 编写或优化 Dockerfile、多阶段构建、Docker Compose 编排时 |
| `deployment-patterns` | CI/CD Pipeline 设计、发布流程、金丝雀与回滚策略时 |
| `e2e-testing` | 编写 Playwright 端到端测试、Page Object Model 与回归套件时 |
| `tdd-workflow` | 执行测试驱动开发（TDD）完整 Red-Green-Refactor 流程与高覆盖率保证时 |
| `using-superpowers` | 每次对话开始、面临任何开发任务前，优先检索与匹配可用 Skills 并规范执行 |
| `brainstorming` | 接到新需求、做新特性/组件设计前，深入探索用户意图、技术权衡与架构边界 |
| `systematic-debugging` | 遭遇代码缺陷、测试失败或意外行为时，强制按“重现-分析-假说证伪-修复”四步科学排错 |
| `writing-plans` | 收到多步骤复杂任务时，动手写代码前先撰写带检查点与可验证步骤的结构化执行计划 |
| `executing-plans` | 配合执行实现计划，分步验证，并在关键检查点与用户对齐确认 |
| `verification-before-completion` | 声称修复/完成或 Git 提交前，强制运行验证命令并检查控制台真实输出证据 |
| `receiving-code-review` | 收到 Code Review 反馈时，进行理性技术推导与实际验证，杜绝盲目认同或机械盲改 |
| `using-git-worktrees` | 开启需要高度隔离的特性开发或执行多任务计划时，使用 Git Worktree 创建干净工作区 |

## 严格遵循事项 (Guardrails)

### 上游优先与下游合并规范 (Upstream-First)

- Wavelet是一个 Cordis 微内核插件化架构的开源项目，所有框架层改动必须在上游修改并合并到下游，禁止直接在下游修改框架层代码。
- **归属判定**：改动位于框架层（`backend/core/`、`backend/pkg/`、通用 `backend/plugins/drivers|infra|domain/`、Cordis/微内核机制）→ **必须先在上游修改**；仅下游业务（`backend/downstream/`、产品页面与业务插件）可直接在下游修改。
- **标准流程**：上游改代码 + 补测试 + 本地 commit（禁止 push）→ 下游 `git fetch wavelet` 确认带入范围（避免拖入无关提交）→ `git merge wavelet/main` → 下游重跑相关测试验证。
- **严禁**直接在下游修改框架层代码（会造成双源分叉、合并冲突）。若已误改，先 `git revert` 撤销下游改动，再按标准流程合并上游。

- 切勿删除 `frontend/node_modules`。
- 保持 `backend/pkg/` 绝对纯净，属于底层通用基础库，**严禁依赖项目上层包（如 `Wavelet/core/*`、`Wavelet/plugins/*`）**；保持 `backend/pkg/util/` 绝对纯净无状态，禁止导入 Gin、GORM、sessions 等 Web/数据库框架包。
- 测试用例禁止硬编码相对路径创建临时目录，统一使用 Go 内置 `t.TempDir()`。
- 修改 API Handler 后运行 `make swagger`，完成代码开发后必须依次运行 `make code-check` 与 `make format`。

### Cordis 架构核心防线与分层规范
- **微内核 (`backend/core/`)**：
  - 上下文总线（`Context`）、泛型依赖注入（`Container`）、生命周期编排（`Lifecycle`）、扩展点定义（`extpoints/`）与领域事件总线（`EventBus`）。
  - **严禁**包含任何具体业务逻辑，**严禁** import `gin`、`gorm`、`asynq` 等具体运行时依赖。
- **服务契约 (`backend/core/contracts/`)**：
  - 跨插件通信的统一公开 Go Interface（如 `AuthService`、`UserService`、`CacheService`、`DBService`、`StorageService`）与公共 DTO。
  - **严禁**包含任何具体业务实现或 SQL 操作。
- **自包含插件 (`backend/plugins/`)**：
  - 所有业务功能与驱动实现均以插件形式存在（`backend/plugins/drivers/`、`backend/plugins/infra/`、`backend/plugins/domain/` 或下游 `backend/downstream/`）。
  - 每个插件实现 `core.Plugin`（`Name() string` 与 `Apply(ctx *core.Context) error`）。
  - **统一插件分层架构与标准模板**：
    - **开发模板唯一基准**：所有插件统一以 `backend/downstream/plugins/custom_example` 为基准模板构建。
    - **物理子包隔离规范**：统一采用物理子包结构（`plugin.go`, `consts/`, `controller/`, `service/`, `dao/`, `model/` [含 `entity/`, `do/`], `migrations/` [含 `postgres/`, `sqlite/`]）。**严禁在根包平铺 `handlers_*`、`service_*`、`dao_*` 等前缀文件**，子包内文件直接按业务实体命名（如 `hello.go`, `user.go`），严格约束 `controller -> service -> dao -> model` 单向依赖。
- **插件通信与依赖隔离**：
  - **严禁跨包 import internal/私有实现**：插件之间严禁直接 import 对方具体实现包代码。
  - **单向服务契约调用**：调用方仅面向 `backend/core/contracts` 编程，在 `Apply` 中通过 `core.Provide[contracts.XxxService](ctx, svc)` 注册服务，通过 `core.Inject[contracts.XxxService](ctx)` 或 `ctx.Using(func(svc contracts.XxxService) { ... })` 声明式解析。
  - **事件总线广播**：状态联动与解耦通信统一通过强类型事件 `ctx.Events().Emit()` 广播，由感兴趣的插件通过 `ctx.Events().On()` 订阅，消除双向依赖与循环引用。
- **扩展点自包含注册**：
  - **HTTP 路由与白名单机制**：
    - 插件自包含在 `Apply` 中通过 `ctx.Router().Group(...)` 挂载路由与中间件，禁止跨插件散落注册。
    - **白名单机制**：`driver_http` 与微内核扩展点提供路由白名单支持（`ctx.Router().RegisterWhitelist(patterns...)`），支持精确路径与通配符（如 `/api/v1/oauth/*`）。
    - **所有权主动声明**：认证域（`auth` 插件）与各业务插件必须在 `Apply` 中主动注册其公开/免鉴权接口（如 `/api/v1/user/login`、`/api/v1/oauth/callback`、`/api/v1/cap/*` 等）。
    - **鉴权中间件放行防线**：`auth` 提供的登录鉴权中间件（`LoginRequired`）必须先执行白名单匹配并自动放行，彻底杜绝免鉴权接口被全局或组级鉴权中间件误拦截（返回 401 Unauthorized）。
  - **异步与定时任务**：插件自包含在 `Apply` 中通过 `ctx.Task().Register(...)` 与 `ctx.Schedule().RegisterCron(...)` 声明。
  - **静态启动配置**：插件自包含在 `Apply` 中通过 `ctx.Config().Bind("<prefix>", &cfg)` 读取**自己声明**的配置，字段以 tag 表达来源：`config`（yaml 路径）、`env`（覆盖变量名）、`default`、`autoEnable`（该变量存在即置真）、`secret`（导出脱敏）。需要在 `Apply` 之前被门禁求值的键，必须在 `DeclareConfig()` 中提前声明并实现 `core.ConfigGatedPlugin`。新增基础设施 key 保持顶层命名（`redis.*`），插件私有配置归 `plugins.<name>.*`。**严禁**再造全局配置单例或在 `backend/pkg/` 读取配置。
  - **动态设置**：插件自包含在 `Apply` 中通过 `ctx.Settings().Register(core.SettingSchema{...})` 声明可热更新的管理台设置模式（与上面的静态启动配置分属两层）。
  - **数据迁移**：插件自包含在内部维护 `migrations/*.sql`，通过 `//go:embed` 打包并在 `Apply` 中通过 `ctx.Migrations().Register(pluginID, embedFS)` 注入。
- **表单一所有者原则与跨插件迁移边界 (Single Owner Principle & Cross-Plugin Migration Boundaries)**：
  - 每张数据表有且仅由一个所有者插件声明与维护（表名使用插件前缀如 `w_order_*`）。
  - **严禁跨插件 DDL**：在业务插件（尤其是下游定制化插件）自身的 migrations 脚本中，绝对严禁执行针对非自身所属表的结构变更（禁止 `CREATE`、`ALTER`、`DROP TABLE`、`ADD/MODIFY COLUMN` 等）。修改共享表结构必须回到上游所有者插件中统一进行。
  - **明确允许跨插件 DML 数据插入 (INSERT)**：当定制化下游业务插件需要预置业务初始化参数或设置项时（例如向平台共享的系统参数表 `w_settings` 插入定制插件所需的配置键值），允许且必须在业务插件自身的 migrations 中执行 `INSERT INTO` 脚本，并保证幂等性。严禁将下游定制项目的初始化数据反向塞入上游通用系统插件中。
  - 严禁插件 B 跨过所有者插件 A 直接在运行时代码编写 SQL/GORM 读写表 A，必须调用插件 A 暴露的 `contracts` 接口或订阅事件。
- **平台服务复用**：
  - 文件摄取统一使用 `upload.Ingest` / `contracts.StorageService`，禁止绕过存储域直接操作底层 Bucket 或直写文件表。
  - 业务缓存统一使用 `ctx.Cache()`（`contracts.CacheService`）或标准缓存框架，禁止自研不带失效广播的本地 map。
  - 数据库操作通过 `ctx.DB()`（`contracts.DBService`）获取受事务与 Trace 保护的连接。

## 后端开发规范

### API 响应规范 (以 `api-design` 规范为主)
- **RESTful 状态码标准**：
  - GET / PUT / PATCH：成功返回 HTTP 200，写出 `c.JSON(http.StatusOK, response.OK(data))`；分页写出 `c.JSON(http.StatusOK, response.Paged(items, meta))`。
  - POST 创建资源：成功返回 HTTP 201，写出 `response.Created(c, location, data)` 并附带 Location 响应头。
  - DELETE / 无返回体操作：成功返回 HTTP 204，写出 `response.NoContent(c)`。
  - 客户端/校验错误：返回 HTTP 400/422，使用 `response.AbortBadRequest(c, msg)` 或 `response.AbortBadRequestWithCode(c, errCode, msg, details...)`。
  - 未登录 / 权限不足 / 资源未找到 / 服务端异常：分别返回 HTTP 401 (`AbortUnauthorized`)、403 (`AbortForbidden`)、404 (`AbortNotFound`)、500 (`AbortInternal`)。
- **标准信封**：
  - 成功数据：`{ "data": ... }`
  - 分页数据：`{ "data": [...], "meta": { "total": ..., "page": ..., "per_page": ... } }`
  - 错误数据：`{ "error": { "code": "...", "message": "...", "details": [...] } }`
  - 兼容字段：`error_msg` 会在错误响应中继续透出，无缝向下兼容老前端调用。
- **错误文案**：使用语义化、清晰的错误文案或错误码（snake_case，如 `validation_error`, `user_not_found`），禁止暴露底层数据库/系统错误细节给客户端。
- **Service/Logics 分工**：业务逻辑层只接受 `context.Context`，返回 `(result, error)`，严禁依赖 `*gin.Context` 或调用 `c.JSON`/`Abort*`。
- **错误日志**：底层错误在 Handler/Logic 边界用 `backend/pkg/logger` 打印日志，禁止使用 `_ = ...` 静默吞掉关键错误。

### 数据库操作
- 插件数据库表结构严禁使用 GORM AutoMigrate，统一编写 Goose SQL 迁移并嵌入二进制。
- 不创建物理外键（显式建索引）；Go 模型零值需与数据库默认值匹配。
- **SQL LIKE 查询防注入与转义**：所有含用户输入的模糊查询必须调用 `backend/pkg/util.EscapeLike` 转义通配符，并显式指定 `ESCAPE '\\'` 语法（如 `Where("username LIKE ? ESCAPE '\\'", util.EscapeLike(keyword)+"%")`），同时兼容 PostgreSQL 与 SQLite 方言并杜绝通配符注入攻击。

### 并发与安全防护规范
- **Goroutine 安全**：禁止直接使用裸 `go func()`；统一使用 `backend/pkg/util.Go`，确保具备未捕获 panic 恢复和调用栈日志记录能力。
- **Pub/Sub 监听并发安全**：启动 Redis Pub/Sub 订阅监听前，必须捕获局部客户端实例，禁止在 goroutine 闭包中直读可变全局变量；提供停止监听接口时必须维护 `done` 通道等待 goroutine 完整退出后再重置状态，消除数据竞争。
- **Session 固定攻击防御**：用户登录/授权成功后，必须调用 Session 轮换逻辑，防止 Session 固定攻击。
- **防账户枚举与时序攻击**：
  - 登录失败统一返回模糊报错；当查询用户不存在时，必须调用 `pkg/util.DummyCheckPassword` 执行同等开销的 bcrypt 哈希计算，彻底消除时序侧信道攻击。
  - 验证码、签名 Token 等敏感字符串比对必须使用 `crypto/subtle.ConstantTimeCompare` 常量时间比对。
- **敏感端点限流**：登录尝试、OAuth 授权发起等敏感接口必须接入基于 Redis 的滑动窗口限流机制，防止暴力破解与缓存资源耗尽。

## 前端开发规范

> **最高优先级规则说明**：通用前端技能（如 `frontend-patterns`、`shadcn`、`design-system` 等）中文档示例可能提及 npm/pnpm 或 Prettier，在本项目中**一律以本规范为最高准则**：**包管理器与运行环境严格锁定为 `bun`，代码格式化工具严格锁定为 `Biome`（`bun run format`），严禁使用 Prettier**。

- 新特性开发前参考 Next.js 文档与 `frontend/app/(main)/admin/demo` 示例代码。
- **页面容器与标题栏**：
    - 页面根容器统一使用全宽 `w-full`，最外层统一用 `py-6` 或 `py-6 px-1` 对齐边距。
    - 标题容器统一 `flex items-center gap-2`（带操作按钮用 `justify-between`）。
    - 图标直接使用 Lucide 组件（`size-5 text-primary`），禁止包裹背景小卡片或装饰边框。
    - 标题文字统一使用 `<h1 className="text-2xl font-semibold tracking-tight">`。
- **无障碍语义与色彩规范 (a11y & WCAG)**：
    - **标题层级规范 (Heading Hierarchy)**：页面中非顶级结构化标题（如空状态提示、加载提示、卡片眉题/卡片标题、抽屉区块名）严禁滥用 `<h3>`/`<h4>`，统一使用 `<p>` 配合样式，保证屏幕阅读器感知的标题层级连续。
    - **无文本控件无障碍**：所有仅包含图标的按钮（如仅有 Icon 的 Button、Switch、无文本的 SelectTrigger）必须显式添加 `aria-label`。
    - **色彩对比度**：正文、提示、徽章等小字颜色在亮色/暗色模式下必须满足 WCAG AA（对比度 ≥ 4.5:1）。
- **组件拆分与维护**：
    - 物理路由页面 `page.tsx` 仅维护高级骨架与布局。
    - 单文件超过 600 行或含多 Tab/大复杂区块时，必须按就近原则拆分为子组件存放在路由同级的 `components/` 局部目录中。
- **样式与服务**：
    - 优先使用 shadcn/ui 的 `variant` 和全局 CSS 变量，不要在业务代码中硬编码颜色/背景。
    - 前端请求统一在 `frontend/lib/services/<name>/` 中继承 `BaseService` 编写并在 `index.ts` 注册。
- **国际化 (i18n)**：
    - 使用 `next-intl`（**无 URL locale 前缀** / non-routing provider 模式），兼容 `NEXT_STANDALONE_EXPORT` 静态导出。
    - 支持语言：`zh-CN`、`en`；默认 `zh-CN`。
    - 解析优先级：cookie `NEXT_LOCALE`（用户显式选择）→ 浏览器语言 → 默认 `zh-CN`。
    - 文案统一放在 `frontend/messages/{locale}.json`，按命名空间嵌套（`common` / `layout` / `auth` / `settings` / 业务域）。
    - 组件内用户可见文案必须通过 `useTranslations()` / `getTranslations()` 读取；**禁止**新增中英硬编码 UI 字符串（后端返回的 `error_msg`、日志、调试信息除外）。
    - key 使用 camelCase 分层（如 `auth.login.submit`）；完整短语作为 value，禁止在组件内拼接句子。
    - 新增或修改文案时必须**同步**更新 `zh-CN.json` 与 `en.json`，保持 key 树一致。
    - 语言选项展示用自称：`中文` / `English`（不随当前 UI 语言翻译）。
    - 日期/数字格式化使用 locale 感知 helper（如 `formatDateTime`），禁止写死 `'zh-CN'` / `date-fns` 的 `zhCN`。
