# Agent Note: 统一前后端 Session Cookie 键名默认值为 wavelet_session 与补齐登录重定向

Status: implemented

## Problem

在用户登录后访问业务页面时，除首页 `/`（公共放行）外，所有页面均被重定向到 `/login?callbackUrl=...`：
1. **前后端默认键名脱节**：后端框架驱动 `driver_http` 与 `auth` 域默认配置标签为 `session_cookie_name: "wavelet_session"`，在未配置该环境变量时下发的 Cookie 名为 `wavelet_session`；而前端 Next.js 代理层 `frontend/proxy.ts` 与部署模板曾使用 `wavelet_session_id` 作为默认查找目标，导致 Next.js 边缘代理读取 Cookie 为空，触发 307 临时重定向。
2. **`LoginForm` 登录成功后缺少显式页面跳转**：账号密码普通登录成功后仅更新了 Context 状态，缺少像 `register-form` 那样调用 `router.replace(target)` 安全重定向的逻辑。

## Decision

1. **统一 Cookie 键名默认值为 `wavelet_session`**：
   - 更新 `manifest/config/config.default.yaml` 与 `frontend/.env.example` 中 `session_cookie_name` / `WAVELET_SESSION_COOKIE_NAME` 默认值为 `wavelet_session`；
   - 在前端代理层 `frontend/proxy.ts` 中，默认查找 `process.env.WAVELET_SESSION_COOKIE_NAME || 'wavelet_session'`，并优雅兼容历史 `wavelet_session_id` 降级读取。
2. **补齐 `LoginForm` 成功重定向逻辑**：
   - 在 `frontend/components/auth/login-form.tsx` 的 `loginMutation.onSuccess` 中，解析并跳转目标 `safeRedirectTarget(callbackUrl || storedRedirect || '/home')`，并清理 `sessionStorage` 临时标志。

## Alternatives considered

- **强行将后端默认值全部重命名为 `wavelet_session_id`**：否决。后端涉及 `auth`、`driver_http` 以及多处既有生产配置与测试用例，`wavelet_session` 命名更紧凑，通过统一为 `wavelet_session` 并对旧名称做无痛降级容错更平滑。

## Consequences

- **收益**：
  - 前后端本地开发、Docker 单容器构建与独立容器部署时的 Session Cookie 键名完全自洽；
  - 登录后访问 `/venue`、`/admin/*` 等业务子路径正常放行并渲染，彻底消除 307 重定向循环；
  - 账密登录成功后能够顺畅跳转回用户原先意图访问的 `callbackUrl`。
- **代价**：
  - 无破坏性变更。
