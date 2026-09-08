# wavelet

🚀 A modern, production-ready full-stack boilerplate for building scalable web applications

[中文](./README_zh.md)

[![License: Apache2.0](https://img.shields.io/badge/License-Apache2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)

## 📖 Introduction

**wavelet** is a generic, production-ready full-stack boilerplate built on **Go (Gin + GORM)** with a **Cordis-inspired microkernel plugin architecture** on the backend and **Next.js (App Router + Shadcn UI + Tailwind CSS 4)** on the frontend. It ships with everything you need to bootstrap a modern SaaS, internal tool, or developer platform — without the boilerplate headaches.

The project was designed from the ground up to be **framework-first and business-agnostic**: plug in your own domain logic or downstream custom plugins while reusing the battle-tested infrastructure that comes out of the box.

### ✨ Key Features

- 🧩 **Cordis Microkernel Architecture** — Decoupled design with kernel lifecycle, service contracts (`contracts`), domain event bus, and modular plugins (`drivers`, `infra`, `domain`, `downstream`)
- 🔐 **Multi-auth System** — Local password login/registration + pluggable OIDC/OAuth2 providers (supports multiple auth sources simultaneously)
- 🗝️ **Personal Access Tokens (PAT)** — API key management for programmatic access; supports `Authorization: Bearer` and `X-Access-Token` headers
- 👤 **User & Access Management** — Admin panel for listing, searching, filtering, and enabling/disabling user accounts
- ⚙️ **Dynamic System Config & Settings** — Declarative schema with live reload and admin UI control
- 📋 **Async Task Queue & Scheduler** — Background job processing with [Asynq](https://github.com/hibiken/asynq) (Redis-backed) and in-process cron scheduler with execution dashboard
- 💾 **Dual-Dialect Database** — PostgreSQL support with zero-config SQLite fallback, embedded dual-dialect Goose SQL migrations, plus ClickHouse for analytics and log stores
- ⚡ **Multi-tier Cache** — High-performance 3-layer cache (RAM L1 + Redis L2 + DB L3) with distributed Pub/Sub cache invalidation
- 📁 **Unified Storage** — S3-compatible, Aliyun OSS, Local disk, and WebDAV storage engines with local disk caching
- 📊 **Observability** — Structured logging (Zap) + distributed tracing (OpenTelemetry) + ring buffer log streaming
- 🌐 **Internationalization (i18n)** — Built-in bilingual support (`zh-CN` / `en`) powered by `next-intl`
- 🎨 **Modern UI** — Responsive, dark-mode-ready design system built with Next.js 16, React 19, Tailwind CSS 4, and Shadcn UI
- 📦 **Single Embedded Binary** — Compile frontend and backend into a single self-contained binary with zero deployment dependencies
- 📖 **Built-in Documentation** — Integrated docs portal with usage guides, Swagger API reference, privacy policy, and terms of service

## 🏗️ Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                     Frontend (Next.js 16)                    │
│   • React 19          • Tailwind CSS 4      • Shadcn UI      │
│   • TypeScript        • next-intl (i18n)    • TanStack Query │
└──────────────────────────────┬───────────────────────────────┘
                               │ HTTP / WebSocket (Port: 8000)
┌──────────────────────────────▼───────────────────────────────┐
│                      Backend (Go 1.25+)                      │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │               Cordis Microkernel (core/)               │  │
│  │  • Context Bus       • Service Contracts (contracts/)  │  │
│  │  • DI Container      • Domain Event Bus                │  │
│  │  • Lifecycle Hooks   • Extension Points (extpoints/)   │  │
│  └──────────────────────────┬─────────────────────────────┘  │
│                             │                                │
│  ┌──────────────────────────▼─────────────────────────────┐  │
│  │              Modular Plugins (plugins/ & downstream/)  │  │
│  │  • Drivers: HTTP (Gin + Embed UI), Asynq Worker, Cron  │  │
│  │  • Infra: Database (Goose), Redis, Cache, Storage      │  │
│  │  • Domain: Auth, User, Admin, Upload, System, Risk     │  │
│  │  • Downstream: Custom business plugins                 │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                     CLI Commands (cmd/)                │  │
│  │  • all (Default)   • api     • worker     • scheduler  │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬───────────────────────────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐       ┌─────────────────┐
│   Database   │       │ Cache & Queue│       │  Object Storage │
│ • PostgreSQL │       │ • Redis      │       │ • S3 / OSS      │
│ • SQLite     │       │ • Valkey     │       │ • Local Disk    │
│ • ClickHouse │       └──────────────┘       │ • WebDAV        │
└──────────────┘                              └─────────────────┘
```

## 🛠️ Tech Stack

### Backend
- **[Go 1.25+](https://go.dev/doc)** — Primary programming language
- **Cordis Microkernel** — Pluggable architecture with dependency injection and lifecycle orchestration
- **[Gin](https://github.com/gin-gonic/gin)** — HTTP web framework
- **[GORM](https://github.com/go-gorm/gorm)** — ORM with PostgreSQL, SQLite, and ClickHouse support
- **[Goose](https://github.com/pressly/goose)** — Dual-dialect SQL migrations embedded directly in plugins
- **[Redis](https://github.com/redis/redis) / [Valkey](https://valkey.io)** — Cache, session store, and task queue backend
- **[Asynq](https://github.com/hibiken/asynq)** — Distributed task queue (Redis-backed)
- **[Cobra + Viper](https://github.com/spf13/cobra)** — CLI entrypoint and configuration management
- **[OpenTelemetry](https://opentelemetry.io)** — Distributed tracing and observability
- **[Zap](https://github.com/uber-go/zap)** — High-performance structured logging
- **[Swagger (Swaggo)](https://github.com/swaggo/swag)** — Auto-generated OpenAPI/Swagger documentation
- **Multi-Storage SDKs** — AWS S3 v2, Aliyun OSS v2, Local disk, WebDAV
- **[Snowflake](https://github.com/bwmarrin/snowflake)** — Distributed ID generation

### Frontend
- **[Next.js 16](https://github.com/vercel/next.js)** — React framework with App Router & Turbopack
- **[React 19](https://github.com/facebook/react)** — Modern UI library with React Compiler support
- **[TypeScript](https://github.com/microsoft/TypeScript)** — Type safety
- **[Tailwind CSS 4](https://github.com/tailwindlabs/tailwindcss)** — Utility-first CSS styling
- **[Shadcn UI](https://github.com/shadcn-ui/ui)** & **[Radix UI](https://www.radix-ui.com/)** — Accessible, composable component library
- **[next-intl](https://next-intl-docs.vercel.app/)** — Type-safe internationalization without URL routing prefixes
- **[Bun](https://bun.sh/)** — Fast JavaScript package manager and runtime

## 📋 Requirements

- **Go** >= 1.25
- **Bun** >= 1.2 (for frontend dependency management and builds)
- **Node.js** >= 18.0 (optional when using Bun)
- **PostgreSQL** >= 14 (optional; SQLite fallback works out of the box with zero external dependencies)
- **Redis** >= 6.0 or **Valkey** >= 7.0 (optional for single-process development)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/Rain-kl/Wavelet.git
cd Wavelet
```

### 2. Configure Environment

Copy either the YAML configuration template or the environment variable template:

```bash
# Option A: YAML configuration (default)
cp manifest/config/config.default.yaml config.yaml

# Option B: Environment variables (takes precedence over config.yaml)
cp .env.example .env
```

Edit `config.yaml` or `.env` as needed. If PostgreSQL and Redis are disabled, Wavelet automatically falls back to SQLite (`wavelet.db`) and in-memory caching.

### 3. Start Local Dependencies (Optional)

```bash
# Start local PostgreSQL (18-alpine) and Valkey/Redis
docker compose up -d

# Optional: also start ClickHouse
docker compose --profile clickhouse up -d

# If using an external PostgreSQL instance, create the database:
createdb -h <host> -p 5432 -U postgres wavelet

# Database migrations run automatically on application startup via embedded Goose SQL migrations.
```

### 4. Start Development Servers

You can start both backend and frontend development servers concurrently using `make`:

```bash
make dev
```

Or run them individually in separate terminal sessions:

**Backend:**
```bash
# Install Go dependencies & generate Swagger docs
cd backend
go mod tidy
make -C .. swagger

# Start backend in fusion mode (API + Worker + Scheduler combined)
go run main.go all
# Alternatively, use make target from repository root:
# make dev-b
```

> The backend CLI also supports standalone execution profiles:
> ```bash
> go run main.go api        # HTTP API server only
> go run main.go worker     # Asynq async task worker only
> go run main.go scheduler  # Cron job scheduler only
> ```

**Frontend:**
```bash
cd frontend

# Install dependencies
bun install

# Start development server with Turbopack
bun dev
# Alternatively, use make target from repository root:
# make dev-f
```

### 5. Access the Application

| Service | URL |
|---|---|
| Frontend Portal | http://localhost:3000 |
| Backend API | http://localhost:8000 |
| Swagger API Docs | http://localhost:8000/swagger/index.html |
| Health Check | http://localhost:8000/api/health |

## ⚙️ Configuration

Key configuration options (see `manifest/config/config.default.yaml` and `.env.example` for the complete reference):

| YAML Key | Environment Variable | Description | Default |
|---|---|---|---|
| `app.addr` | `APP_ADDR` | Backend listening address | `:8000` |
| `app.env` | `APP_ENV` | Environment mode (`development` / `production`) | `production` |
| `database.enabled` | `DB_ENABLED` | Enable PostgreSQL (`false` falls back to SQLite) | `true` |
| `database.host` | `DB_HOST` | PostgreSQL host | `127.0.0.1` |
| `database.database` | `DB_NAME` | Database name | `wavelet` |
| `database.sqlite_path` | `SQLITE_PATH` | SQLite file path (when SQLite is used) | `wavelet.db` |
| `redis.enabled` | `REDIS_ENABLED` | Enable Redis/Valkey cache and task queue | `true` |
| `redis.addrs` | `REDIS_ADDR` | Redis address | `127.0.0.1:6379` |
| `storage.type` | `STORAGE_TYPE` | Storage engine (`s3`, `oss`, `local`, `webdav`) | `local` |

## 🔧 Development Guide

### Makefile Commands

From the repository root:

```bash
# Run both frontend & backend concurrently in development
make dev

# Regenerate Swagger API documentation
make swagger

# Format all backend Go code and frontend code
make format

# Run comprehensive architecture check, golangci-lint, TypeScript typecheck & ESLint
make code-check

# Compile single embedded binary (frontend bundled inside backend)
make build-embedded

# Cross-compile release binaries for Linux / macOS / Windows
make cross-build
```

### Frontend Commands

```bash
cd frontend

# Development server (Turbopack)
bun dev

# Production build
bun run build

# Production build with static export (for embedding into Go binary)
bun run build:embed

# Start production server
bun start

# Lint & format
bun run lint
bun run format
```

## 📁 Project Structure

```
wavelet/
├── Makefile                 # Automation scripts (dev, swagger, format, code-check, build)
├── docker-compose.yml       # Local infrastructure services (PostgreSQL 18, Valkey, Jaeger, ClickHouse)
├── manifest/                # Deployment and configuration manifests
│   ├── config/              # Configuration templates (config.default.yaml)
│   ├── docker/              # Dockerfiles (production, backend, frontend, cross-compile)
│   └── deploy/              # Deployment manifests (Kubernetes / Helm)
├── backend/                 # Go backend (Cordis microkernel architecture)
│   ├── main.go              # Backend entrypoint (cmd.Execute)
│   ├── cmd/                 # CLI commands (all, api, scheduler, worker, reset_passwd)
│   ├── core/                # Cordis microkernel (Context, Container, Lifecycle, Events)
│   │   └── contracts/       # Public service interfaces (AuthService, DBService, etc.)
│   ├── pkg/                 # Low-level generic libraries (logger, trace, idgen, response)
│   ├── plugins/             # Pluggable modular plugins
│   │   ├── drivers/         # Runtime drivers (HTTP with embedded frontend, Asynq, Cron)
│   │   ├── infra/           # Infrastructure plugins (database, redis, cache, storage, config)
│   │   └── domain/          # Business domain plugins (auth, user, admin, upload, system)
│   ├── downstream/          # Downstream deployment-specific custom plugins & extensions
│   └── docs/                # Swagger auto-generated documentation
└── frontend/                # Next.js frontend application
    ├── app/                 # Next.js App Router pages and layouts
    ├── components/          # Reusable UI components (ui, common, layout, theme)
    ├── hooks/               # Custom React hooks
    ├── lib/                 # Base service classes, API services, and utilities
    ├── messages/            # i18n translation catalogs (zh-CN.json, en.json)
    └── types/               # TypeScript definitions
```

## 📚 API Documentation

Swagger API documentation is automatically available when running the backend:

```
http://localhost:8000/swagger/index.html
```

The built-in documentation portal at `/docs` contains:
- **Usage Guide** — Step-by-step walkthrough for getting started
- **API Reference** — Detailed interface documentation
- **Privacy Policy** — Template privacy policy (customize as needed)
- **Terms of Service** — Template terms of service

## 🧪 Testing & Code Quality

```bash
# Backend test suite
cd backend && go test ./...

# Full project static analysis & architecture guardrail verification
make code-check

# Frontend linting
cd frontend && bun run lint
```

## 🚀 Deployment

### 1. Self-Contained Embedded Binary (Recommended)

Build a single static executable with the compiled frontend embedded inside:

```bash
# Build binary -> ./bin/wavelet
make build-embedded

# Run application
./bin/wavelet all
```

The embedded binary serves both the Next.js frontend and the Go API from a single port (`:8000`), with zero runtime Node.js or static file dependencies.

### 2. Cross-platform Binaries

Build static binaries for 6 target platforms (Linux / macOS / Windows × amd64 / arm64) using Docker BuildKit:

```bash
# Build all 6 targets -> ./bin/
make cross-build

# Stamp a specific release version
make cross-build VERSION=v1.0.0

# Target a specific OS or architecture
make cross-build GOOS=linux GOARCH=amd64
```

### 3. Docker

```bash
# Build container image
docker build -f manifest/docker/Dockerfile -t wavelet .

# Run with environment variables or mounted configuration
docker run -d -p 8000:8000 \
  --env-file .env \
  wavelet all
```

## 🤝 Contributing

We welcome contributions! Please read the following before submitting code:

- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Contributor License Agreement](CLA.md)

### Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'feat: add your feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
