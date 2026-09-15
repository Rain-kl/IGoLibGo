# Agent Note: igo 插件全量 API 契约审计、路由与状态对齐复盘

Status: implemented

## Problem

在针对 `igo` 插件（`backend/igo-lib/plugins/igo/`）及前端业务页面进行全量契约与运行时联调排查时，发现数处跨端参数传递、HTTP 语义与字段命名上的不一致：
1. **DELETE 请求 Body 数据截断**：前端 `BaseService.delete` 原未处理 `config?.data`，导致 `deleteSeatLabels` 携带的 `seat_keys` 请求体被 Axios 忽略；
2. **全域捡漏任务路由错位**：前端 `IGoTaskService` 中 `startLeak` 与 `cancelLeak` 路由请求了旧路径 `/tasks/leak/start`，而微内核注册的权威路由为 `/tasks/global-leak/start`；
3. **Query 参数嵌套包装错误**：`listActivityLogs` 与 `listTaskRecords` 在前端调用 `this.get` 时错误将参数多重包裹为 `{ params: query }`，导致 URL 查询字符串序列化异常；同时后端 `controller.go` 的 `parsePage` 原仅支持 `per_page`，未兼容前端习惯传入的 `limit`；
4. **任务调度启动与取消响应契约**：前端抢座、占座、全域捡漏与明日预约页面在调度操作后均调用 `setStatus(res)` 消费最新的 `CoordinatorStatus`，而后端 `StartTask` / `CancelTask` / `RunTomorrowNow` 原返回 204 No Content，导致前端 `status` 被置空；
5. **一条龙配置已有凭据保留与权威单命名字段**：
   - 一条龙卡片编辑时，后端返回脱敏后的 `cookie_masked` 与 `has_cookie: true`，前端若强制重新校验 Cookie 会阻断编辑流程；
   - 彻底裁撤与清理所有冗余的别名字段（如 `BeaconMac`/`BeaconLat`/`BeaconLng` 及 `CookieValid`/`CheckinTokenValid`），全面收敛至权威规范命名（`beacon_uuid`, `latitude`, `longitude`, `has_cookie`, `has_checkin_token`）；
   - 执行结果 `PipelineRunResult` 填充 `name` 字段，确保结果弹窗能正确展示配置名。

## Decision

1. **底层 BaseService 与 Query 参数修复**：
   - 在 `frontend/lib/services/core/base.service.ts` 中修正 `delete` 方法，从 `config.data` 或 `params.data` 提取 payload 并注入 Axios 请求体；
   - 修正 `dashboard.ts` 与 `tasks.ts` 中的参数直传，消除多余的 `{ params: ... }` 包装；
   - 在 `backend/igo-lib/plugins/igo/controller/controller.go` 中增强 `parsePage`，优先使用 `per_page`，回落至 `limit`。
2. **统一权威单路由与状态同步**：
   - 将 `tasks.ts` 中捡漏路径严格对齐至 `/tasks/global-leak/start` 与 `/tasks/global-leak/cancel`，杜绝双路由别名；
   - 在 `service/task.go` 中新增 `GetTaskStatus(ctx, userID, kind)`，并将 `StartTask`、`CancelTask`、`RunTomorrowNow` 更新为返回 `200 OK` 携带最新 `CoordinatorStatus`，与前端类型契约无缝对齐；
   - 将 `listTaskRecords` 增强为支持按 `kind` 过滤。
3. **一条龙（Pipeline）全生命周期契约强化与别名裁撤**：
   - 遵从“禁止任何别名”原则，全面裁撤 `BeaconMac`、`BeaconLat`、`BeaconLng`、`CookieValid`、`CheckinTokenValid` 等冗余别名字段，在 Go DO 模型、TS Service 类型、UI 组件及 E2E 测试中统一仅使用标准权威字段；
   - 在 `RunPipeline` 及其各阶段返回中填充 `Name: row.Name`，确保执行结果弹框能够完整显示配置名；
   - 在前端 `pipeline-dialog.tsx` 中增加对已有有效凭据的保留机制（`PRESERVE_EXISTING`），支持仅修改目标场馆/座位而不必重复授权；
   - 在 `pipeline-card.tsx` 中使用 `has_cookie` / `has_checkin_token` 标准字段评估凭据状态。

## Consequences

- **收益**：
  - 全量 43 个 API 端点达成 100% 契约一致性，零别名杂音，涵盖路由路径、HTTP Method、Query 参数、Request Body 与 Response Envelope；
  - 彻底消除各类因为未传 `data`、嵌套 `params`、缺少响应体或字段命名引起的隐性 Bug；
  - 全部单元测试、架构合规性检查（`make code-check`）、Swagger 规范生成（`make swagger`）、代码格式化（`make format`）及 Playwright E2E 测试 100% 通过。
- **代价**：
  - 无破坏性改动，所有接口保持向后兼容。
