# Agent Note: 修复一条龙自动化凭据验证契约对齐与前端多维有效性判定

Status: implemented

## Problem

在用户配置一条龙自动化卡片第一步录入凭据时：
1. 后端接口 `POST /api/v1/igo/pipeline/verify-session` 返回 HTTP 200 成功响应：
   ```json
   {
       "data": {
           "cookie": "Authorization=eyJhbGciOiJSUzI1NiJ9...",
           "expires_at": "2026-09-15T12:08:41+08:00",
           "libraries": [ ... ]
       }
   }
   ```
2. 前端 `PipelineDialog` 组件在 `handleVerifySession` 和 `handleNextFromStep1` 中过度严格依赖布尔键 `res.valid`（`if (res.valid)` / `valid = Boolean(sessionRes?.valid)`）。
3. 因后端原控制器响应体 Map 遗漏了 `"valid": true` 键，前端解析 `res.valid` 为 `undefined`（falsy），导致即使后端已成功换取 Cookie 并拉取到场馆列表，前端仍然误弹 Toast 报错 `TraceInt 凭据验证失败，请重新获取授权链接`，阻断向步骤 2（选择目标座位）流转。

## Decision

1. **后端控制器显式下发 `valid: true`**：
   - 在 `backend/igo-lib/plugins/igo/controller/pipeline.go` 的 `HelperVerifySession` 与 `HelperVerifyCheckin` 中显式注入 `"valid": true`；
   - 同步修正 Swagger 注解中的路由路径为规范单路由 `/api/v1/igo/pipeline/verify-session`、`/library-layout` 与 `/verify-checkin`。
2. **前端契约与判定鲁棒化（双重防线）**：
   - 在 `frontend/lib/services/igo/types.ts` 中将 `PipelineHelperVerifySessionResponse` 与 `PipelineHelperVerifyCheckinResponse` 的 `valid` 标记为可选 `valid?: boolean`；
   - 在 `frontend/app/(main)/pipeline/components/pipeline-dialog.tsx` 中实施弹性判定：若 `res.valid` 为真，或包含非空 `cookie`、或包含非空场馆列表 `libraries`，均认定凭据有效并正常进入下一步；
   - 并在签到凭据验证 `handleVerifyCheckin` 中同步应用弹性判定（`res.valid || res.token`）。
3. **补齐回归测试防线**：
   - 在 `backend/igo-lib/plugins/igo/pipeline_e2e_test.go` 中断言 `verifyRes.Data["valid"] == true`；
   - 在 `frontend/e2e/pipeline.spec.ts` 中新增针对后端省略 `valid` 字段时的容错端到端回归用例，确保全链路畅通。

## Consequences

- **收益**：
  - 彻底解决用户提交有效凭据或授权链接后被误报凭据无效的阻断缺陷；
  - 前端具备多维容错能力，无论后端返回 `valid: true` 还是仅返回凭据与场馆数据，均能正常流转至步骤 2。
- **代价**：
  - 无破坏性改动，完全向下兼容。
