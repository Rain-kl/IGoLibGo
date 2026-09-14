# Agent Note: 认证中间件向下游标准 context.Context 注入登录态对象

Status: implemented

## Problem

在基于 Cordis 插件的业务开发中，API 路由通常挂载 `authSvc.RequireAuthMiddleware().(gin.HandlerFunc)`，业务 Handler 按照标准 Go HTTP / Cordis 规范从 `c.Request.Context()` 中获取登录态用户（例如 `authSvc.GetCurrentUser(c.Request.Context())`）。

此前，`LoginRequiredMiddleware` 与 `AdminRequiredMiddleware` 在通过 Session 或 Access Token 鉴权成功后，仅调用了 `ginutil.SetToContext(c, contracts.AuthUserObjKey, user)` 将用户挂载到 Gin 的私有 `c.Keys` 字典中，**未将用户对象及用户 ID 注入到底层的 `c.Request.Context()`**。

这导致：
1. 下游业务 Controller/Service 调用 `authSvc.GetCurrentUser(c.Request.Context())` 时，`ctx.Value(contracts.AuthUserObjKey)` 为 `nil`，且内部尝试转换 `ctx.(*gin.Context)` 失败；
2. 下游业务直接报错 401 Unauthorized（例如 `{"error_msg":"未登录","error":{"code":"unauthorized","message":"未登录"}}`），前端 Axios 响应拦截器捕获 401 后立即触发 `initiateLogin` 并跳回登录页，造成“登录成功后访问业务页面无限重定向回登录页”的恶性循环。

## Decision

1. **中间件请求上下文注入**：在 `LoginRequiredMiddleware` 与 `AdminRequiredMiddleware` 中，鉴权成功后除保持 `ginutil.SetToContext(c, ...)` 外，同步通过 `context.WithValue` 将 `contracts.AuthUserObjKey` 和 `contracts.AuthUserIDKey` 写入 `c.Request.Context()`，并更新 `c.Request = c.Request.WithContext(reqCtx)`。
2. **测试用例覆盖**：在 `service_test.go` 的 `TestLoginRequiredMiddlewarePopulatesServiceContext` 中，为 Session 与 Access Token 双链路分别增加从 `c.Request.Context()` 读取 `GetCurrentUser` 的断言，确保中间件能无缝打通下游任意遵循标准 `context.Context` 契约的代码。

## Alternatives considered

- **仅在下游业务 Controller 中强转或传入 `c` (`*gin.Context`)**：否决。根据 Hexagonal 与 Cordis 规范，Service 层与通用逻辑只能依赖标准 `context.Context`，强制所有下游代码穿透感知 `*gin.Context` 违背了架构分层与依赖反转原则。
- **让 `GetCurrentUserID` 强行解析普通 Context**：否决。既有单元测试与历史契约明确约束 `GetCurrentUserID` 保留纯粹的 Gin Session 解析语义；而 `GetCurrentUser` 才是读取标准 Context 泛化用户态（Session / Token / Mock）的官方契约。通过在中间件规范化注入 `contracts.AuthUserObjKey`，即可让所有调用方均能通过 `GetCurrentUser(ctx)` 正常获取用户。

## Consequences

- **收益**：
  - 彻底打通上游框架层鉴权中间件与下游所有插件/业务 Handler 之间的 Context 管道；
  - 杜绝已登录用户访问下游接口被误判 401 并跳回登录页的问题；
  - 保持 Gin Session 原有契约测试 100% 兼容通过。
- **代价**：
  - 无。
