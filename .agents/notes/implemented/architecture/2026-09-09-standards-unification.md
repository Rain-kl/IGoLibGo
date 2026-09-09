# Agent Note: 统一项目开发规范与 Skills 体系

Status: implemented

## Problem

在引入丰富技能体系后，项目出现了严重的“三方规则冲突”：
1. **API 契约冲突**：通用技能 `api-design` 规定了 RESTful 状态码与 `{ data, meta, error: { code, message, details } }`，而既有规范与 `AGENTS.md` 写死为 `{ error_msg, data }` 并全部返回 HTTP 200；
2. **日志架构冲突**：通用技能 `go-logging` 强制使用 Go 标准库 `log/slog` 并禁止 Zap 与格式化字符串，而 Wavelet 实际深度依赖基于 Zap 的 `backend/pkg/logger` 及专用于管理台 WebSocket 实时控制台流的 5000 行 `GlobalRingBuffer`；
3. **消息推送过时**：`wv-push-notification` 仍然充斥重构前废弃的 `internal/apps/admin/push/` 说明；
4. **插件子包分层脱节**：`wv-new-api` 和 `wv-new-async-task` 文档提及在插件根目录平铺 `handlers.go`/`tasks.go`，且官方下游基准模板 `custom_example` 只有空包；
5. **数据库迁移权限边界模糊**：未界定下游定制业务插件能否向平台共享参数表（如 `w_settings`）插入业务初始化参数。

## Decision

以 **“标准统一，以 Skills 为主”** 为核心原则落地整改：
1. **API 响应全面对齐 `api-design`**：重构 `backend/pkg/response`，提供标准 RESTful 状态码（200/201/204/400/404 等）与统一信封（`Response[T]`, `PagedResponse[T]`, `ErrorResponse`），同时在错误响应中继续透出 `error_msg`；并全面同步更新前端对接层（`types.ts`、`api-client.ts`、`base.service.ts`、`api-envelope.ts` 与 `proxy.ts`），安全解包 204 No Content，结构化提取错误码与明细，确保老前端/调用方完全平滑兼容。
2. **重构日志规范为项目专属技能 `wv-logging`**：将 `go-logging` 重命名为 `wv-logging`，全面以 `backend/pkg/logger` + `contracts.LoggerService` 结合 OTel Trace 与 5000 行 `GlobalRingBuffer` 为唯一准则。
3. **彻底纠正 `wv-*` 架构描述**：
   - `wv-push-notification` 按 `backend/plugins/domain/msg_gateway`、物理子包与 `contracts.PushService` 重写；
   - `wv-new-api` 与 `wv-new-async-task` 废弃根目录平铺描述，规范化为物理子包隔离，并修正模块路径为 `Wavelet/...`。
4. **增强 `wv-database-guide` 迁移规约**：
   - 严禁非所属插件对表结构执行 DDL（表单一所有者防线）；
   - 明确允许定制化下游插件在自身 migrations 中向平台共享表执行 DML `INSERT INTO` 注入初始参数。
5. **完善 `custom_example` 基准模板**：提供完整 `controller -> service -> dao -> model` 物理子包实现及双方言迁移示例。
6. **前端工具链规则收敛**：通用前端技能保持通用，由 `AGENTS.md` 强制规则最高优先级覆盖：全项目前端严格锁定 `bun` + `Biome`。

## Alternatives considered

- **方案 A：API 响应维持旧的 `{ error_msg, data }` 且全 200**：放弃。全 200 抹杀了 HTTP 状态码的语义优势（例如无法使用标准的 201 Created、204 No Content，缓存与网关监控无法感知 4xx/5xx），且违背了用户明确的“以 api-design 通用技能为主”指令。
- **方案 B：直接在现有代码中推翻底层 Zap 强行切换为 `log/slog`**：放弃。管理后台的实时 WebSocket 控制台日志流严重依赖 Zap Core 的 `MultiWriteSyncer` 和 `GlobalRingBuffer`，贸然重写底层会导致管理后台功能瘫痪。将技能改造为 `wv-logging` 是契合项目现状的最佳路径。

## Consequences

- **收益**：
  - 架构规则、Skills 指南、代码模板和 `AGENTS.md` 四位一体 100% 严丝合缝；
  - 新业务或下游定制插件开发时，直接复制 `custom_example` 即可获得符合规范的可用骨架；
  - API 具备规范的 RESTful 语义与错误结构，同时向后兼容现有前端调用。
- **代价**：
  - 开发者编写 HTTP API 时需要注意方法语义（如 POST 创建需调用 `response.Created` 返回 201，DELETE 调用 `response.NoContent` 返回 204）。
