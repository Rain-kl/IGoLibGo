# Agent Note: 彻底移除 API 信封 error_msg 兼容字段

Status: implemented

## Problem

在 [standards-unification](../architecture/2026-09-09-standards-unification.md) 中，项目全面引入了 `api-design` 规范的标准 RESTful 响应契约（`{ data, meta, error: { code, message, details } }`）。为了平滑向下兼容重构前的历史老前端/老调用方，当时在后端的错误响应信封（`Response[T]`, `ErrorResponse`, `Any`）以及前端的解析层（`types.ts`, `api-client.ts`, `api-envelope.ts`, `proxy.ts`）中保留了顶层的 `error_msg` 字符串字段。

随着全系统前端组件、服务层及 API 客户端全部基于标准 `error` 结构体运作，顶层 `error_msg` 字段已成为冗余的技术债务：
1. **契约不纯粹**：同一份错误信息在 JSON 中既存在于 `error.message`，又并存于顶层 `error_msg`，违反单一事实来源原则；
2. **多重回退开销与维护负担**：前端与测试代码需要针对 `res.error` 和 `res.error_msg` 双重判定与回退。

## Decision

彻底清理前后端全链路 API 信封中的 `error_msg` 兼容字段，实现纯净的 `api-design` 错误响应标准：
1. **后端信封收敛**：
   - 从 `backend/pkg/response` 的 `ErrorResponse`、`Response[T]` 及 Swagger 类型 `Any` 中移除 `ErrorMsg` 字段，响应体仅返回标准 `{ "error": { "code": "...", "message": "...", "details": [...] }, "data": null }`；
   - 构造器 `Err` 与 `ErrWithCode` 不再赋值 `ErrorMsg`；
   - `driver_http` 中的 405 MethodNotAllowed 统一使用 `response.ErrWithCode`；
   - 全面更新单元测试断言与 Swagger OpenAPI 文档。
2. **前端消费层收敛**：
   - `frontend/lib/services/core/types.ts` 中的 `ApiResponse` 和 `ApiError` 移除 `error_msg`；
   - `api-client.ts` 响应/错误拦截器与 `extractErrorMessage` 仅依赖 `data.error`；
   - `api-envelope.ts` 信封解析与形状判断彻底消除 `error_msg`；
   - `proxy.ts` 代理网关在 429/502 错误时输出纯净的标准结构化错误；
   - 前端在线 API 文档（`api.tsx`）及中英文语言包移除 `error_msg` 说明。

## Alternatives considered

- **方案 A：继续保留 `error_msg` 字段仅打上 `@deprecated` 注解**：放弃。全栈均为自闭环现代代码库，不存在外部依赖黑盒，继续保留冗余字段只会模糊架构边界，延迟必然的技术清理。
- **方案 B：仅从后端移除，前端保留兜底逻辑**：放弃。前后端同源维护，前端保留已无意义的废弃字段检查会增加死代码体积并误导后续贡献者。

## Consequences

- **收益**：
  - 全局 API 错误响应格式严格符合 `api-design` 契约，零历史包袱；
  - 减少 JSON 序列化传输冗余并简化前后端错误解包判定逻辑；
  - Swagger 文档和前端 API 文档彻底呈现现代化单一错误契约。
- **代价**：
  - 外部未经升级、仍硬编码读取响应根路径 `body.error_msg` 的调用方将无法获取错误文案，必须按照规范迁移至 `body.error.message`。
