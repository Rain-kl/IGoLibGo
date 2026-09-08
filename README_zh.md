# wavelet

🚀 现代化、生产就绪的全栈应用脚手架

[English](./README.md)

[![License: Apache2.0](https://img.shields.io/badge/License-Apache2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)

## 📖 项目简介

**wavelet** 是一个通用型、生产就绪的现代全栈脚手架，后端基于 **Go (Gin + GORM)** 并采用 **Cordis 风格微内核插件化架构**，前端采用 **Next.js (App Router + Shadcn UI + Tailwind CSS 4)**。项目开箱即用，内置构建现代 SaaS、内部工具或开发者平台所需的核心基础设施。

项目设计理念是 **框架优先、业务中立**：您可以在沿用经过实战检验的底层基础设施的同时，自由接入自己的业务逻辑或下游定制插件。

### ✨ 主要特性

- 🧩 **Cordis 微内核架构** — 解耦设计的内核生命周期、服务契约（`contracts`）、领域事件总线与模块化插件（`drivers`、`infra`、`domain`、`downstream`）
- 🔐 **多认证体系** — 本地账号密码登录/注册 + 可插拔 OIDC/OAuth2 认证源（支持同时配置多个认证源）
- 🗝️ **个人访问令牌 (PAT)** — API Key 密钥管理，支持程序化接口访问；兼容 `Authorization: Bearer` 和 `X-Access-Token` 请求头
- 👤 **用户与权限管理** — 管理后台提供用户列表、搜索筛选、启用/禁用账号等管理功能
- ⚙️ **动态系统配置与设置** — 声明式 Schema 配置管理，支持实时热重载，可通过管理后台界面直接操作
- 📋 **异步任务队列与定时调度** — 基于 [Asynq](https://github.com/hibiken/asynq)（Redis 驱动）的后台任务处理系统与进程内 Cron 调度，附带任务执行看板
- 💾 **双方言数据库支持** — 深度支持 PostgreSQL 与零配置 SQLite 回落，内置双方言 Goose SQL 迁移；支持 ClickHouse 分析库与日志存储
- ⚡ **多级缓存体系** — 高性能三层缓存（RAM L1 + Redis L2 + DB L3），内置分布式 Pub/Sub 缓存失效广播
- 📁 **统一多引擎存储** — 支持 S3 兼容协议、阿里云 OSS、本地磁盘及 WebDAV，支持本地磁盘缓存
- 📊 **全链路可观测性** — 结构化日志（Zap）+ 分布式链路追踪（OpenTelemetry）+ 内存环形缓冲区日志实时流
- 🌐 **完整国际化 (i18n)** — 基于 `next-intl` 实现的双语支持（`zh-CN` / `en`）
- 🎨 **现代化 UI** — 基于 Next.js 16、React 19、Tailwind CSS 4 和 Shadcn UI 构建的响应式、支持深色模式的设计系统
- 📦 **单二进制文件内嵌部署** — 支持将前端构建资源完整内嵌至 Go 二进制中，实现零外部依赖单文件部署
- 📖 **内置文档中心** — 集成文档门户，包含使用指南、Swagger 接口文档、隐私政策和服务条款

## 🏗️ 架构概览

```
┌──────────────────────────────────────────────────────────────┐
│                     前端 (Next.js 16)                        │
│   • React 19          • Tailwind CSS 4      • Shadcn UI      │
│   • TypeScript        • next-intl (i18n)    • TanStack Query │
└──────────────────────────────┬───────────────────────────────┘
                               │ HTTP / WebSocket (端口: 8000)
┌──────────────────────────────▼───────────────────────────────┐
│                      后端 (Go 1.25+)                         │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │               Cordis 微内核 (core/)                    │  │
│  │  • 上下文总线 (Context) • 服务契约 (contracts/)         │  │
│  │  • 依赖注入容器 (DI)    • 领域事件总线 (EventBus)       │  │
│  │  • 生命周期编排         • 扩展点注册 (extpoints/)       │  │
│  └──────────────────────────┬─────────────────────────────┘  │
│                             │                                │
│  ┌──────────────────────────▼─────────────────────────────┐  │
│  │              模块化插件 (plugins/ 与 downstream/)       │  │
│  │  • 运行时驱动: HTTP (Gin+内嵌前端), Asynq Worker, Cron  │  │
│  │  • 基础设施层: Database (Goose), Redis, Cache, Storage │  │
│  │  • 业务领域层: Auth, User, Admin, Upload, System, Risk  │  │
│  │  • 下游业务层: 业务定制插件与扩展                       │  │
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

## 📋 环境要求

- **Go** >= 1.25
- **Bun** >= 1.2（用于前端依赖管理与构建）
- **Node.js** >= 18.0（使用 Bun 时可选）
- **PostgreSQL** >= 14（可选；支持零配置 SQLite 自动回落，无需外部数据库即可直接运行）
- **Redis** >= 6.0 或 **Valkey** >= 7.0（单进程轻量开发时可选）

## 🚀 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/Rain-kl/Wavelet.git
cd Wavelet
```

### 2. 配置环境

复制 YAML 配置文件模板或环境变量模板：

```bash
# 方案 A：YAML 配置文件（默认）
cp manifest/config/config.default.yaml config.yaml

# 方案 B：环境变量文件（优先级高于 config.yaml）
cp .env.example .env
```

按需修改 `config.yaml` 或 `.env`。若禁用 PostgreSQL 与 Redis，Wavelet 会自动回退到 SQLite (`wavelet.db`) 和进程内存缓存。

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

## ⚙️ 配置说明

主要配置项（完整参数说明请参考 `manifest/config/config.default.yaml` 与 `.env.example`）：

| YAML 键名 | 环境变量 | 说明 | 默认值 |
|---|---|---|---|
| `app.addr` | `APP_ADDR` | 后端服务监听地址 | `:8000` |
| `app.env` | `APP_ENV` | 运行环境（`development` / `production`） | `production` |
| `database.enabled` | `DB_ENABLED` | 是否启用 PostgreSQL（`false` 时回退至 SQLite） | `true` |
| `database.host` | `DB_HOST` | PostgreSQL 主机地址 | `127.0.0.1` |
| `database.database` | `DB_NAME` | 数据库名称 | `wavelet` |
| `database.sqlite_path` | `SQLITE_PATH` | SQLite 文件存储路径（启用 SQLite 时有效） | `wavelet.db` |
| `redis.enabled` | `REDIS_ENABLED` | 是否启用 Redis/Valkey 缓存与队列 | `true` |
| `redis.addrs` | `REDIS_ADDR` | Redis 服务连接地址 | `127.0.0.1:6379` |
| `storage.type` | `STORAGE_TYPE` | 存储引擎类型（`s3`、`oss`、`local`、`webdav`） | `local` |

## 🔧 开发指南

### Makefile 常用指令

在项目根目录下执行：

```bash
# 并发启动前后端开发服务
make dev

# 修改 Handler 后重新生成 Swagger 文档
make swagger

# 格式化前后端代码（Go fmt + Biome format）
make format

# 执行全量质量门禁：Cordis 架构检查、golangci-lint、TypeScript 类型检查与 ESLint
make code-check

# 构建内嵌前端的独立单二进制执行程序
make build-embedded

# 通过 Docker 跨平台编译全量 Release 二进制包 (Linux / macOS / Windows)
make cross-build
```

### 前端开发指令

```bash
cd frontend

# 开发模式（Turbopack 极速热更新）
bun dev

# 生产环境构建
bun run build

# 导出静态资源用于嵌入 Go 二进制文件
bun run build:embed

# 运行生产构建服务
bun start

# 代码 Lint 校验与格式化
bun run lint
bun run format
```

## 📁 项目结构

```
wavelet/
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
│   │   └── domain/          # 业务领域插件（auth、user、admin、upload、system）
│   ├── downstream/          # 下游定制化业务专属插件与扩展
│   └── docs/                # Swagger 自动生成的 API 文档
└── frontend/                # Next.js 前端应用
    ├── app/                 # Next.js App Router 页面与布局
    ├── components/          # 可复用组件（ui、common、layout、theme）
    ├── hooks/               # 自定义 React Hooks
    ├── lib/                 # 基础服务类、API 服务层与工具库
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
# 后端测试用例
cd backend && go test ./...

# 全量质量门禁检查（含架构防线、Lint 与前端类型检查）
make code-check

# 前端代码检查
cd frontend && bun run lint
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
docker build -f manifest/docker/Dockerfile -t wavelet .

# 运行容器（传入环境配置）
docker run -d -p 8000:8000 \
  --env-file .env \
  wavelet all
```

## 🤝 贡献指南

我们欢迎社区贡献！请在提交代码前阅读以下文档：

- [贡献指南](CONTRIBUTING.md)
- [行为准则](CODE_OF_CONDUCT.md)
- [贡献者许可协议](CLA.md)

### 贡献流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -am 'feat: add your feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 打开 Pull Request

## 📄 许可证

本项目基于 [Apache 2.0 许可证](LICENSE) 开源。
