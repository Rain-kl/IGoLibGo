# Wavelet - IGoLibrary

🚀 A modern, production-ready full-stack library automation and intelligent service platform

[中文](./README_zh.md)

[![License: Apache2.0](https://img.shields.io/badge/License-Apache2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)

## 📖 Introduction

**Wavelet - IGoLibrary** is a production-ready full-stack application built on top of the **Cordis microkernel pluggable architecture**, powered by **Go (Gin + GORM)** on the backend and **Next.js 16 (App Router + React 19 + Shadcn UI + Tailwind CSS 4)** on the frontend. The project is dedicated to providing high-performance, reliable experiences for library automated seat reservations, seat grabbing, auto-renewal, global seat leak scanning, and remote iBeacon check-ins.

The project adopts a layered design of **Microkernel + Self-contained Plugins**: the lower layer relies on Wavelet's general infrastructure (dual database dialects, multi-tier caching, message gateway, task queue) to provide solid foundation support, while the upper layer implements the complete IGoLibrary business ecosystem in `backend/igo-lib/plugins/igo`.

### ✨ Key Features

- ⚡ **"All-in-One" Automation Pipeline** — Multi-account/card-based automated seat reservation and remote check-in, completing "seat lookup $\rightarrow$ seat locking $\rightarrow$ iBeacon-based simulated check-in" in one click. Built-in credential validity probing, re-authorization guidance on expiration, and an execution dashboard.
- 🤖 **Message Gateway & Bot Integration** — Supports slash commands (`/help`, `/show`, `/run [ID]`) across Telegram / QQ bot platforms, automatically pushes WeChat OAuth authorization links on credential expiration, and supports interactive credential entry.
- 🚀 **4 Automation Engines** — 
  - **Seat Grabbing Engine**: Second-level scheduled polling and concurrent reservation strategies;
  - **Seat Occupation Engine**: Monitors seated status and automatically renews seats before release;
  - **Global Seat Leak Engine**: Scans across multiple venues and locks available seats in real time;
  - **Tomorrow Reservation Engine**: Preloads tomorrow's open seats and triggers scheduled reservation.
- 🏛️ **Interactive Venue Navigation** — Fetches real-time seat layout diagrams, supporting quick seat search, selection, favorite seat lists, and custom seat remark tags.
- 🧩 **Cordis Microkernel Architecture** — Decoupled design with kernel lifecycle, service contracts (`contracts`), domain event bus, and modular plugins (`drivers`, `infra`, `domain`, `igo`).
- 🔐 **Multi-Authentication System** — Local username/password login/registration + WeChat OAuth authorization binding, supporting token masking and live refresh.
- ⚙️ **Dynamic System Config & Settings** — Declarative schema configuration management with live hot-reloading, controllable directly via the admin UI.
- 📋 **Async Task Queue & Cron Scheduling** — Background task processing based on [Asynq](https://github.com/hibiken/asynq) (Redis-backed) and in-process Cron scheduler with a task execution dashboard.
- 💾 **Dual-Dialect Database Support** — Deep support for PostgreSQL with zero-config SQLite fallback, embedded dual-dialect Goose SQL migrations; supports ClickHouse analytics and log storage.
- ⚡ **Multi-Tier Cache System** — High-performance 3-layer caching (RAM L1 + Redis L2 + DB L3) with distributed Pub/Sub cache invalidation broadcasting.
- 📊 **Full Observability** — Structured logging (Zap) + distributed tracing (OpenTelemetry) + real-time memory ring buffer log streaming.
- 🌐 **Full Internationalization (i18n)** — Bilingual support (`zh-CN` / `en`) powered by `next-intl`.
- 🎨 **Modern UI** — Responsive, dark-mode-ready design system built with Next.js 16, React 19, Tailwind CSS 4, and Shadcn UI.
- 📦 **Single Embedded Binary Deployment** — Full embedded frontend build assets into Go binary for zero-dependency single-file deployment.

## 🏗️ Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Frontend (Next.js 16 App Router)          │
│   • All-in-One Pipeline (/pipeline) • Seat Grabbing (/grab)   │
│   • Seat Occupation (/occupy)       • Seat Leak (/leak)      │
│   • Tomorrow Reserve (/tomorrow)    • Venue Guide (/venue)   │
│   • React 19 / Tailwind CSS 4 / Shadcn UI / next-intl        │
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
│  │           Modular Plugins (plugins/ & igo-lib/)        │  │
│  │  • Drivers: HTTP (Gin + Embed UI), Asynq Worker, Cron  │  │
│  │  • Infra: Database (Goose), Redis, Cache, Storage      │  │
│  │  • Domain: Auth, User, Admin, Upload, MsgGateway, Risk │  │
│  │  • Business: IGoLibrary downstream plugin (igo)        │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                     CLI Commands (cmd/)                │  │
│  │  • all (Fusion Mode) • api    • worker   • scheduler   │  │
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
- **[Playwright](https://playwright.dev/)** — End-to-End (E2E) automated testing suite

## 📋 Requirements

- **Go** >= 1.25
- **Bun** >= 1.2 (for frontend dependency management, builds, and Playwright E2E testing)
- **Node.js** >= 18.0 (optional when using Bun)
- **PostgreSQL** >= 14 (optional; SQLite fallback works out of the box with zero external database dependencies)
- **Redis** >= 6.0 or **Valkey** >= 7.0 (optional for single-process development)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/Rain-kl/IGoLibGo.git
cd IGoLibGo
```

### 2. Configure Environment

Copy either the YAML configuration template or the environment variable template:

```bash
# Option A: YAML configuration (default)
cp manifest/config/config.default.yaml config.yaml

# Option B: Environment variables (takes precedence over config.yaml)
cp .env.example .env
```

Edit `config.yaml` or `.env` as needed. If PostgreSQL and Redis are disabled, the application automatically falls back to SQLite (`wavelet.db`) and in-memory caching.

### 3. Start Local Dependencies (Optional)

```bash
# Start local PostgreSQL (18-alpine) and Valkey/Redis
docker compose up -d

# Optional: also start ClickHouse
docker compose --profile clickhouse up -d

# If using an external PostgreSQL instance, create the database:
createdb -h <host> -p 5432 -U postgres wavelet

# Database schema migrations run automatically on application startup via embedded Goose scripts.
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

## 📁 Project Structure

```
IGoLibGo/
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
│   │   └── domain/          # Business domain plugins (auth, user, admin, upload, system, msg_gateway)
│   ├── igo-lib/             # IGoLibrary downstream domain plugin and custom features
│   │   └── plugins/igo/     # IGo primary business plugin (pipeline, grab, occupy, leak, tomorrow, venue)
│   └── docs/                # Swagger auto-generated documentation
└── frontend/                # Next.js frontend application
    ├── app/                 # Next.js App Router pages and layouts
    │   └── (main)/          # Business primary routes (/pipeline, /grab, /occupy, /leak, /tomorrow, /venue)
    ├── components/          # Reusable UI components (ui, igo, common, layout, theme)
    ├── e2e/                 # Playwright End-to-End automated test suite
    ├── hooks/               # Custom React hooks
    ├── lib/                 # Base service classes, API service layer (services/igo), and utilities
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
# 1. Full backend unit & integration tests
cd backend && go test ./...

# 2. Frontend Playwright End-to-End (E2E) tests
cd frontend && bunx playwright test

# 3. Frontend typecheck & ESLint validation
cd frontend && bunx tsc --noEmit && bun run lint

# 4. Full architecture guardrail & code check (Cordis guardrails, golangci-lint, TypeScript & ESLint)
make code-check

# 5. Global automatic formatting (Biome + gofumpt)
make format
```

## 🚀 Deployment

### 1. Self-Contained Embedded Binary (Recommended)

Build a single static executable with the compiled frontend embedded inside with one command:

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

### 3. Docker Containerization

```bash
# Build container image
docker build -f manifest/docker/Dockerfile -t igolib .

# Run container (pass environment file)
docker run -d -p 8000:8000 \
  --env-file .env \
  igolib all
```

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
