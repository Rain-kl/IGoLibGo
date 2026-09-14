# Agent Note: 将 /igo-settings 合并到 /admin/settings 系统设置

Status: implemented

## Problem

在先前的多页面结构中，IGo 插件拥有独立的 `/igo-settings` 页面并在侧边栏的 IGO 业务分组下挂载了“系统设置”导航项。
这导致了多重体验与架构上的割裂：
1. **侧边栏重复项**：IGO 分组与管理后台（Admin）分组均存在名为“系统设置”的导航项，造成严重的认知混淆。
2. **管理入口分散**：平台已具备统一的 `/admin/settings`（管理后台-系统设置），包含了安全设置、业务设置、系统设置、状态监控等核心管理能力；将 IGo 相关的运行时参数、协议模板与数据备份割裂在外部独立页面不利于统一管理。

## Decision

将 `/igo-settings` 全量合并至 `/admin/settings`：
1. **组件化封装**：将原设置页内容重构成独立的 `IGoTab` 组件（`frontend/components/igo/settings/igo-tab.tsx`），内含运行与网络参数、协议模板、加密备份与恢复、平台通知中心 4 个子面板。
2. **挂载到管理后台**：在 `/admin/settings` 中新增 `igo` 标签（中英文对应 “IGo 设置” / “IGo Settings”），支持 `?tab=igo` URL 查询参数直达并支持动态懒加载（`next/dynamic`）。
3. **消除侧边栏重复项**：从侧边栏（`sidebar.tsx`）的 `igoNavItems` 中移除重复的 `settings` 导航项，统一收口至管理后台的“系统设置”。
4. **历史路由向后兼容**：保留 `/igo-settings` 路由并在服务端通过 Next.js `redirect('/admin/settings?tab=igo')` 平滑重定向，确保历史书签与访问链接不失效。

## Alternatives considered

- **在 IGO 分组保留“业务设置”并重命名**：虽然避免了同名问题，但核心参数（如 GraphQL 协议模板、网络超时、备份还原）本质上属于管理员系统设置范畴，依然存在管理入口分散的问题。
- **直接删除 `/igo-settings` 路由**：可能导致历史书签、浏览器访问历史或外部链接抛出 404 错误，不如 307/308 重定向更健壮。

## Consequences

- **收益**：
  - 侧边栏结构更加清晰，消除了重复的“系统设置”导航项。
  - 管理员可以在 `/admin/settings` 一站式管理平台基础配置与 IGo 业务引擎配置。
  - 保持了完整的向后兼容性与 URL 直达能力。
- **代价**：
  - 非管理员用户如果访问 IGo 设置，会由 `/admin/settings` 统一校验鉴权并跳转至个人资料页（符合管理权限隔离要求）。

## Verification

- `bun x tsc --noEmit`：TypeScript 类型检查通过。
- `bun run format:check`：Biome 代码格式校验通过。
- `bun run build`：Next.js 全量生产构建成功，包含 `/admin/settings` 与 `/igo-settings` 路由。
- `go test ./...`：后端测试套件全量通过。
