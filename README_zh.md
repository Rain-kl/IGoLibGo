# Wavelet - IGoLibrary

🚀 现代化、生产就绪的全栈图书馆自动化与智能化服务平台

[English](./README.md)

[![License: Apache2.0](https://img.shields.io/badge/License-Apache2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)

## 📖 项目简介

**Wavelet - IGoLibrary** 是一个基于 **Cordis 微内核插件化架构** 构建的生产就绪全栈应用，后端基于 **Go (Gin + GORM)**，前端采用 **Next.js 16 (App Router + React 19 + Shadcn UI + Tailwind CSS 4)**。项目专注于提供高性能、可靠的图书馆自动化占座、抢座、自动续座、全域捡漏以及远程打卡签到全流程体验。

项目采用 **微内核 + 自包含插件** 的分层设计：底层借由 Wavelet 通用基础设施（数据库双方言、多级缓存、消息网关、任务队列）提供坚实支撑，上层在 `backend/igo-lib/plugins/igo` 中实现了完整的 IGoLibrary 业务生态。

### ✨ 主要特性

- ⚡ **「一条龙」自动化全流程 (All-in-One Automation Pipeline)** — 支持多账号/卡片化自动化预定与远程签到，一键完成“查座 $\rightarrow$ 锁定座位 $\rightarrow$ 基于 iBeacon 模拟打卡”，内置凭据有效性探测、失效重授权引导与执行结果大盘。
- 🤖 **消息网关与 Bot 智能交互 (Message Gateway & Bot Integration)** — 支持 Telegram / QQ 等多平台 Bot 斜杠命令（`/help`、`/show`、`/run [ID]`），凭据失效时自动推送微信 OAuth 授权链接并支持交互式补录凭据。
- 🚀 **四大自动化引擎 (4 Automation Engines)** — 
  - **抢座引擎**：秒级定时轮询抢座与并发预约策略；
  - **在座守护引擎**：守护在座状态并在释放前自动续座；
  - **全域捡漏引擎**：多场馆全局扫描与空余座位实时锁定；
  - **明日预约引擎**：明日开放预约的预加载与定时秒杀。
- 🏛️ **可视化场馆与座位排布导览 (Interactive Venue Navigation)** — 实时拉取座位布局排布图，支持快速查座、选座、收藏备选座位与自定义座位备注标签。
- 🧩 **Cordis 微内核架构** — 解耦设计的内核生命周期、服务契约（`contracts`）、领域事件总线与模块化插件（`drivers`、`infra`、`domain`、`igo`）
- 🔐 **多认证体系** — 本地账号密码登录/注册 + 微信 OAuth 授权绑定，支持 Token 脱敏与热更新
- ⚙️ **动态系统配置与设置** — 声明式 Schema 配置管理，支持实时热重载，可通过管理后台界面直接操作
- 📋 **异步任务队列与定时调度** — 基于 [Asynq](https://github.com/hibiken/asynq)（Redis 驱动）的后台任务处理系统与进程内 Cron 调度，附带任务执行看板
- 💾 **双方言数据库支持** — 深度支持 PostgreSQL 与零配置 SQLite 回落，内置双方言 Goose SQL 迁移；支持 ClickHouse 分析库与日志存储
- ⚡ **多级缓存体系** — 高性能三层缓存（RAM L1 + Redis L2 + DB L3），内置分布式 Pub/Sub 缓存失效广播
- 📊 **全链路可观测性** — 结构化日志（Zap）+ 分布式链路追踪（OpenTelemetry）+ 内存环形缓冲区日志实时流
- 🌐 **完整国际化 (i18n)** — 基于 `next-intl` 实现的双语支持（`zh-CN` / `en`）
- 🎨 **现代化 UI** — 基于 Next.js 16、React 19、Tailwind CSS 4 和 Shadcn UI 构建的响应式、支持深色模式的设计系统
- 📦 **单二进制文件内嵌部署** — 支持将前端构建资源完整内嵌至 Go 二进制中，实现零外部依赖单文件部署

## 🏗️ 架构概览

```
┌──────────────────────────────────────────────────────────────┐
│                  前端界面 (Next.js 16 App Router)            │
│   • 一条龙自动化 (/pipeline) • 抢座引擎 (/grab)             │
│   • 在座守护 (/occupy)        • 全域捡漏 (/leak)             │
│   • 明日预约 (/tomorrow)      • 场馆选座 (/venue)            │
│   • React 19 / Tailwind CSS 4 / Shadcn UI / next-intl        │
└──────────────────────────────┬───────────────────────────────┘
                               │ HTTP / WebSocket (端口: 8000)
┌──────────────────────────────▼───────────────────────────────┐
│                      后端服务 (Go 1.25+)                     │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │               Cordis 微内核 (core/)                    │  │
│  │  • 上下文总线 (Context) • 服务契约 (contracts/)         │  │
│  │  • 依赖注入容器 (DI)    • 领域事件总线 (EventBus)       │  │
│  │  • 生命周期编排         • 扩展点注册 (extpoints/)       │  │
│  └──────────────────────────┬─────────────────────────────┘  │
│                             │                                │
│  ┌──────────────────────────▼─────────────────────────────┐  │
│  │              自包含插件体系 (plugins/ & igo-lib/)       │  │
│  │  • 驱动层: HTTP (Gin+内嵌前端), Asynq Worker, Cron     │  │
│  │  • 基础层: Database (Goose), Redis, Cache, Storage     │  │
│  │  • 领域层: Auth, User, Admin, Upload, MsgGateway, Risk  │  │
│  │  • 业务层: IGoLibrary downstream plugin (igo)          │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                     CLI 命令体系 (cmd/)                │  │
│  │  • all (融合默认模式) • api    • worker   • scheduler   │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬───────────────────────────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐       ┌─────────────────┐
│   数据库层   │       │  缓存与队列  │       │    对象存储     │
│ • PostgreSQL │       │ • Redis      │       │ • S3 / OSS      │
│ • SQLite     │       │ • Valkey     │       │ • 本地磁盘      │
│ • ClickHouse │       └──────────────┘       │ • WebDAV        │
└──────────────┘                              └─────────────────┘
```

## 🛠️ 技术栈

### 后端
- **[Go 1.25+](https://go.dev/doc)** — 主语言
- **Cordis 微内核** — 插件化扩展体系、依赖注入与生命周期编排
- **[Gin](https://github.com/gin-gonic/gin)** — HTTP Web 框架
- **[GORM](https://github.com/go-gorm/gorm)** — ORM，支持 PostgreSQL、SQLite 与 ClickHouse
- **[Goose](https://github.com/pressly/goose)** — 插件自包含嵌入式双方言 SQL 迁移
- **[Redis](https://github.com/redis/redis) / [Valkey](https://valkey.io)** — 缓存、Session 存储与任务队列后端
- **[Asynq](https://github.com/hibiken/asynq)** — 分布式任务队列（Redis 驱动）
- **[Cobra + Viper](https://github.com/spf13/cobra)** — CLI 入口与配置管理
- **[OpenTelemetry](https://opentelemetry.io)** — 分布式链路追踪与可观测性
- **[Zap](https://github.com/uber-go/zap)** — 结构化高性能日志
- **[Swagger (Swaggo)](https://github.com/swaggo/swag)** — 自动生成 OpenAPI/Swagger 文档
- **多存储引擎 SDK** — AWS S3 v2、阿里云 OSS v2、本地磁盘、WebDAV
- **[Snowflake](https://github.com/bwmarrin/snowflake)** — 分布式 ID 生成

### 前端
- **[Next.js 16](https://github.com/vercel/next.js)** — React 框架（App Router 与 Turbopack）
- **[React 19](https://github.com/facebook/react)** — UI 库（支持 React Compiler 优化）
- **[TypeScript](https://github.com/microsoft/TypeScript)** — 完整类型安全
- **[Tailwind CSS 4](https://github.com/tailwindlabs/tailwindcss)** — 原子化 CSS 框架
- **[Shadcn UI](https://github.com/shadcn-ui/ui)** & **[Radix UI](https://www.radix-ui.com/)** — 可访问、可组合的组件库
- **[next-intl](https://next-intl-docs.vercel.app/)** — 无 URL 路由前缀的类型安全国际化
- **[Bun](https://bun.sh/)** — 高性能 JavaScript 包管理器与运行时
- **[Playwright](https://playwright.dev/)** — 端到端 (E2E) 自动化测试套件

## 📋 环境要求

- **Go** >= 1.25
- **Bun** >= 1.2（用于前端依赖管理、打包与 Playwright E2E 测试）
- **Node.js** >= 18.0（使用 Bun 时可选）
- **PostgreSQL** >= 14（可选；支持零配置 SQLite 自动回落，无需外部数据库即可直接运行）
- **Redis** >= 6.0 或 **Valkey** >= 7.0（单进程轻量开发时可选）

## 🚀 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/Rain-kl/IGoLibGo.git
cd IGoLibGo
```

### 2. 配置环境

复制 YAML 配置文件模板或环境变量模板：

```bash
# 方案 A：YAML 配置文件（默认）
cp manifest/config/config.default.yaml config.yaml

# 方案 B：环境变量文件（优先级高于 config.yaml）
cp .env.example .env
```

按需修改 `config.yaml` 或 `.env`。若禁用 PostgreSQL 与 Redis，应用会自动回退到 SQLite (`wavelet.db`) 和进程内存缓存。

### 3. 启动本地依赖服务（可选）

```bash
# 启动本地 PostgreSQL (18-alpine) 与 Valkey/Redis
docker compose up -d

# 可选：同时启动 ClickHouse
docker compose --profile clickhouse up -d

# 如果使用独立安装的 PostgreSQL，手动创建数据库：
createdb -h <主机> -p 5432 -U postgres wavelet

# 数据库表结构在应用首次启动时通过嵌入式 Goose 脚本自动迁移，无需手动执行
```

### 4. 启动开发服务

通过 `make` 命令可一键并发启动前后端开发服务：

```bash
make dev
```

或在不同终端窗口分别启动：

**后端：**
```bash
# 安装依赖并生成 Swagger 接口文档
cd backend
go mod tidy
make -C .. swagger

# 以融合模式启动后端（同时启动 API、Worker 和 Scheduler）
go run main.go all
# 或在仓库根目录下运行：
# make dev-b
```

> 后端 CLI 也支持按独立进程角色运行：
> ```bash
> go run main.go api        # 仅启动 HTTP API 服务
> go run main.go worker     # 仅启动 Asynq 异步任务消费 Worker
> go run main.go scheduler  # 仅启动 Cron 定时调度器
> ```

**前端：**
```bash
cd frontend

# 安装依赖
bun install

# 启动开发服务器（Turbopack）
bun dev
# 或在仓库根目录下运行：
# make dev-f
```

### 5. 访问应用

| 服务 | 地址 |
|---|---|
| 前端界面 | http://localhost:3000 |
| 后端 API | http://localhost:8000 |
| Swagger 接口文档 | http://localhost:8000/swagger/index.html |
| 健康检查 | http://localhost:8000/api/health |

## 📁 项目结构

```
IGoLibGo/
├── Makefile                 # 自动化脚本（开发、Swagger、格式化、代码检查、构建）
├── docker-compose.yml       # 本地基础设施编排（PostgreSQL 18、Valkey、Jaeger、ClickHouse）
├── manifest/                # 项目清单与部署编排
│   ├── config/              # 配置模板（config.default.yaml）
│   ├── docker/              # Dockerfile 镜像文件（生产镜像、跨平台编译镜像）
│   └── deploy/              # 生产部署清单（Kubernetes / Helm）
├── backend/                 # Go 后端（Cordis 微内核插件化架构）
│   ├── main.go              # 程序主入口（cmd.Execute）
│   ├── cmd/                 # CLI 命令集（all、api、scheduler、worker、reset_passwd）
│   ├── core/                # Cordis 微内核（Context、Container、Lifecycle、Events）
│   │   └── contracts/       # 跨插件公开服务契约（AuthService、DBService 等）
│   ├── pkg/                 # 底层纯通用基础库（logger、trace、idgen、response）
│   ├── plugins/             # 自包含模块化插件
│   │   ├── drivers/         # 运行时驱动（HTTP 内嵌前端驱动、Asynq Worker、Cron）
│   │   ├── infra/           # 基础设施插件（database、redis、cache、storage、config）
│   │   └── domain/          # 业务领域插件（auth、user、admin、upload、system、msg_gateway）
│   ├── igo-lib/             # IGoLibrary 下游专属领域插件与商业化模块
│   │   └── plugins/igo/     # IGo 主业务插件（一条龙、抢座、占座、捡漏、定时预约、场馆）
│   └── docs/                # Swagger 自动生成的 API 文档
└── frontend/                # Next.js 前端应用
    ├── app/                 # Next.js App Router 页面与布局
    │   └── (main)/          # 业务主路由 (/pipeline, /grab, /occupy, /leak, /tomorrow, /venue)
    ├── components/          # 可复用组件（ui、igo、common、layout、theme）
    ├── e2e/                 # Playwright 端到端自动化测试套件
    ├── hooks/               # 自定义 React Hooks
    ├── lib/                 # 基础服务类、API 服务层 (services/igo) 与工具库
    ├── messages/            # i18n 国际化翻译文案（zh-CN.json、en.json）
    └── types/               # TypeScript 类型定义
```

## 📚 接口文档

Swagger 接口文档在后端启动后自动可用：

```
http://localhost:8000/swagger/index.html
```

前端文档中心（路径 `/docs`）内置以下内容：
- **使用指南** — 分步入门教程
- **接口文档** — 详细接口说明
- **隐私政策** — 隐私政策模板（请按需自定义）
- **服务条款** — 服务条款模板

## 🧪 测试与质量门禁

```bash
# 1. 后端全量单元测试与集成测试
cd backend && go test ./...

# 2. 前端 Playwright 端到端 (E2E) 测试
cd frontend && bunx playwright test

# 3. 前端类型检查与 ESLint 校验
cd frontend && bunx tsc --noEmit && bun run lint

# 4. 全量架构防线与质量检查 (含 Cordis 架构防线、golangci-lint、TypeScript & ESLint)
make code-check

# 5. 全局自动格式化 (Biome + gofumpt)
make format
```

## 🚀 部署发布

### 1. 内嵌单二进制部署（推荐）

一条命令构建内嵌完整前端资源的高性能独立二进制文件：

```bash
# 构建二进制文件 → ./bin/wavelet
make build-embedded

# 启动服务
./bin/wavelet all
```

内嵌二进制在单个端口（`:8000`）上同时托管前端界面与后端 API，无任何 Node.js 运行时或静态文件外挂依赖。

### 2. 跨平台多架构编译

基于 Docker BuildKit 一键构建覆盖 6 大平台的静态二进制分发包（Linux / macOS / Windows × amd64 / arm64）：

```bash
# 构建全部 6 个平台二进制 → ./bin/
make cross-build

# 指定发行版本号
make cross-build VERSION=v1.0.0

# 指定目标系统或架构
make cross-build GOOS=linux GOARCH=amd64
```

### 3. Docker 容器化运行

```bash
# 构建镜像
docker build -f manifest/docker/Dockerfile -t igolib .

# 运行容器（传入环境配置）
docker run -d -p 8000:8000 \
  --env-file .env \
  igolib all
```

## 📄 许可证

本项目基于 [Apache 2.0 许可证](LICENSE) 开源。
