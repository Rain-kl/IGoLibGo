# IGoLibrary Go + Next 迁移路线

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `IGoLibrary/IGoLibrary-Ex` 的可迁移能力重写为 Wavelet Cordis 下游插件 + Next.js 页面，保留原功能、对齐 wavelet 规范。

**Architecture:** 单插件 `igo` 落在 `backend/igo-lib/plugins/igo`，HTTP 前缀 `/api/v1/igo/`。桌面单用户模型改为按 Wavelet `user_id` 隔离。任务协调器落到 Asynq；通知复用平台 `msg_gateway`。前端在现有 sidebar 增加与原授权侧栏等量的业务菜单（剔除明确不迁移项）。

**Tech Stack:** Go 1.27 / Gin / GORM / Goose（PG + SQLite）/ Asynq；Next.js + bun + Biome；响应信封 `Wavelet/pkg/response`。

## Global Constraints

- 后端代码只放 `backend/igo-lib/`；插件分层必须是 `plugin.go` + `consts/` + `controller/` + `service/` + `dao/` + `model/{entity,do}/` + `migrations/{postgres,sqlite}/`。
- API 前缀固定 `/api/v1/igo/` + 原始路径（有桌面 HTTP 的保留原后缀；其余按资源名词补齐）。
- 成功 `{ "data": ... }`，分页附 `meta`，错误 `{ "error": { "code", "message", "details" }, "error_msg" }`；禁止全 200。
- 不迁移：手机控制、cloudflared、通知设置（走平台通知）、客户端专属（托盘、开机启动、阻止休眠、本地外观、本地存储路径、自动更新、局域网快传、桌面 Toast/提示音）。
- 严禁 import `plugins/domain/*` 内部实现；鉴权只走 `contracts.AuthService`。
- 禁止 GORM AutoMigrate；表前缀 `igo_*`（不用平台 `w_`）；无物理外键。
- 前端包管理器 `bun`，格式化 `Biome`。

---

## 范围对照

### 原侧栏（授权后 10 项）→ Web 菜单（8 项）

| 原菜单 | 迁移 | Web 路由 |
| :--- | :--- | :--- |
| 首页 | 是 | `/igo` |
| 账户与场馆 | 是 | `/igo/account` |
| 抢座 | 是 | `/igo/grab` |
| 全域捡漏 | 是 | `/igo/global-leak` |
| 明日预约 | 是 | `/igo/tomorrow` |
| 占座 | 是 | `/igo/occupy` |
| 远程签到 | 是 | `/igo/checkin` |
| 手机控制 | 否 | — |
| 自动通知 | 否（平台通知模块） | — |
| 系统设置 | 是（裁剪） | `/igo/settings` |

系统设置原 5 个分类：常规 / 外观 / 网络与接口 / 存储与日志 / 关于。Web 只保留：**常规（任务默认值）**、**网络与接口（协议模板 + 超时重试，去掉 Cloudflare/手机控制）**、**备份与恢复（本地备份导入导出）**。外观、关于、日志路径走平台。

### 原桌面 HTTP（手机控制）路径映射

这些路径属于手机控制功能本身不迁 UI，但操作是 Web 也需要的，因此保留原后缀：

| 原路径 | 新路径 |
| :--- | :--- |
| `GET /api/status` | `GET /api/v1/igo/status` |
| `GET /api/task-records` | `GET /api/v1/igo/task-records` |
| `GET /api/session/auth-qrcode` | `GET /api/v1/igo/session/auth-qrcode` |
| `POST /api/tasks/{kind}/start` | `POST /api/v1/igo/tasks/:kind/start` |
| `POST /api/tasks/{kind}/cancel` | `POST /api/v1/igo/tasks/:kind/cancel` |
| `POST /api/reservation/cancel` | `POST /api/v1/igo/reservation/cancel` |
| `POST /api/session/cookie/refresh` | `POST /api/v1/igo/session/cookie/refresh` |

`kind` 扩展为 `grab | occupy | global-leak | tomorrow`（原手机控制只有 grab/occupy）。

### 明确不迁

- 手机控制页、局域网/Cloudflare Tunnel、Clash/Mihomo 兼容、快传落地页
- 通知渠道配置（SMTP/Telegram/Bark/Server酱/WxPusher/本地 Toast/提示音）→ 任务事件改为 `contracts.PushService`
- 启动器、更新器、托盘、阻止休眠、记住窗口尺寸、本地主题、本地数据目录迁移

---

## 施工阶段

### 阶段 1 — 接口桩（本轮）

只注册路由、Request/Response DTO、Swagger、501 `not_implemented` 信封。不建表、不调 TraceInt。

### 阶段 2 — 实体与迁移

按用户隔离落表（双方言 Goose）：

- `igo_sessions` Cookie 会话
- `igo_venues` 绑定场馆
- `igo_favorites` 收藏座位
- `igo_seat_labels` 座位标签
- `igo_protocol_overrides` 协议覆盖
- `igo_settings` 每用户设置 JSON
- `igo_task_runs` 任务运行状态
- `igo_task_launch_history` 启动历史
- `igo_global_leak_targets` 捡漏场馆
- `igo_global_leak_blacklist` 捡漏黑名单
- `igo_checkin_sessions` 远程签到会话
- `igo_dashboard_metrics` 首页累计成功/守护时长

### 阶段 3 — 逻辑实现

按原 Application 服务迁移，不改 TraceInt 语义：

1. TraceInt HTTP/GraphQL 客户端（协议模板可覆盖）
2. Session：授权链接解析 / Cookie / 刷新 / 恢复 / 登出
3. Venue：拉馆、预览、锁定、布局、规则、收藏、标签
4. Reservation：刷新、取消
5. Grab / GlobalLeak / Occupy / Tomorrow 状态机 + Asynq 循环
6. Remote check-in：独立授权、设备、签到
7. Protocol editor、Settings、Backup
8. 任务成功/失败/Cookie 到期 → 平台推送（不迁通知设置页）

### 阶段 4 — 其他后端

Cron（Cookie 到期扫描）、活动日志、限流、配置 schema、健康检查。`make swagger && make format && make code-check`。

### 阶段 5 — 前端

1. sidebar 对齐 8 个业务菜单 + i18n
2. 按页重写，对照原 ViewModel，功能不漏：
   - 首页：问候/状态/累计成功/守护时长/当前预约/引擎摘要
   - 账户与场馆：授权二维码、链接解析、Cookie、恢复会话、馆列表、预览、锁定、座位图、收藏、标签
   - 抢座：多目标、策略、轮询模式、定时、监控日志
   - 全域捡漏：多馆、扫描间隔、黑名单、实时日志
   - 明日预约：单目标、触发时间、立即执行一次
   - 占座：当前预约、重预约间隔、检查间隔模式
   - 远程签到：独立微信授权、Beacon、坐标、签到
   - 系统设置：任务默认、协议模板、超时重试、备份
3. 通知入口接到平台通知设置，不新建渠道页

---

## 阶段 1 API 清单（47）

鉴权：全部挂 `AuthService.RequireAuthMiddleware()`。未实现返回 HTTP 501，`error.code = not_implemented`。

### 首页 / 状态

- `GET /api/v1/igo/dashboard`
- `GET /api/v1/igo/status`（原 `/api/status`）
- `GET /api/v1/igo/activity-logs`

### 会话

- `GET /api/v1/igo/session`
- `GET /api/v1/igo/session/auth-qrcode`
- `POST /api/v1/igo/session/from-code`
- `POST /api/v1/igo/session/from-cookie`
- `POST /api/v1/igo/session/cookie/refresh`
- `DELETE /api/v1/igo/session`

### 场馆与座位

- `GET /api/v1/igo/libraries`
- `GET /api/v1/igo/libraries/bound`
- `POST /api/v1/igo/libraries/bound/refresh`
- `GET /api/v1/igo/libraries/:id`
- `GET /api/v1/igo/libraries/:id/layout`
- `GET /api/v1/igo/libraries/:id/rule`
- `POST /api/v1/igo/libraries/:id/bind`
- `POST /api/v1/igo/libraries/:id/preview`
- `GET /api/v1/igo/libraries/:id/favorites`
- `PUT /api/v1/igo/libraries/:id/favorites`
- `PUT /api/v1/igo/libraries/:id/seat-labels`
- `DELETE /api/v1/igo/libraries/:id/seat-labels`

### 预约

- `GET /api/v1/igo/reservation`
- `POST /api/v1/igo/reservation/refresh`
- `POST /api/v1/igo/reservation/cancel`

### 任务

- `GET /api/v1/igo/tasks`
- `GET /api/v1/igo/task-records`
- `POST /api/v1/igo/tasks/tomorrow/run-now`
- `POST /api/v1/igo/tasks/:kind/start`
- `POST /api/v1/igo/tasks/:kind/cancel`
- `GET /api/v1/igo/global-leak/blacklist`
- `PUT /api/v1/igo/global-leak/blacklist`
- `GET /api/v1/igo/global-leak/selected-libraries`
- `PUT /api/v1/igo/global-leak/selected-libraries`

### 远程签到

- `GET /api/v1/igo/checkin/session`
- `GET /api/v1/igo/checkin/auth-qrcode`
- `POST /api/v1/igo/checkin/from-code`
- `GET /api/v1/igo/checkin/devices`
- `POST /api/v1/igo/checkin/sign`
- `DELETE /api/v1/igo/checkin/session`

### 协议 / 设置 / 备份

- `GET /api/v1/igo/protocol/templates`
- `GET /api/v1/igo/protocol/templates/defaults`
- `PUT /api/v1/igo/protocol/templates`
- `POST /api/v1/igo/protocol/templates/reset`
- `GET /api/v1/igo/settings`
- `PUT /api/v1/igo/settings`
- `POST /api/v1/igo/backup/export`
- `POST /api/v1/igo/backup/import`

---

## Key Decisions

1. **单插件 `igo`，不是按菜单拆插件** — 场馆、任务、签到共享 Cookie 与 TraceInt 客户端，拆插件会制造循环依赖。
2. **多租户按 Wavelet user_id** — 桌面是单机单用户；Web 必须隔离。
3. **任务跑 Asynq，不在 HTTP 里死循环** — 原协调器是桌面长驻循环。
4. **通知只发事件给平台** — 不迁渠道配置页。
5. **协议模板仍可覆盖** — 原「自定义 API 地址」要保留，否则学校接口变更无法热修。

## 验证

阶段 1：`go test ./igo-lib/plugins/igo/...` 断言 47 条路由 + 501 信封 + 校验 400。`TestNewWaveletAppProfiles` 插件数 18，含 `igo`。
