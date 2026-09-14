# Agent Note: 裁撤 IGoLibrary WebDAV 云端同步特性

Status: implemented

## Problem

原 Avalonia 桌面版 IGoLibrary 包含 WebDAV 多端数据与设置同步能力。迁移至 Wavelet Web 服务端架构后，所有用户配置、场馆绑定与任务状态均已原生持久化于 PostgreSQL / SQLite 数据库并按 `user_id` 严格隔离，Web 端天生具备云端漫游能力，额外的 WebDAV 同步成为非必要且增加维护负担与密码明文存储风险的冗余特性。

## Decision

彻底移除 IGo 插件内所有 WebDAV 同步相关逻辑：
1. **数据库层**：从初始迁移（`00001_initial.sql`）与 `consts.OwnedTables` 中裁撤 `igo_webdav` 表与 `entity.WebDAV` 模型。
2. **后端层**：移除 `GET /api/v1/igo/webdav`、`PUT /api/v1/igo/webdav` 与 `POST /api/v1/igo/webdav/sync` 路由及对应的 Controller、Service、DAO 方法。
3. **前端层**：删除 `WebDAVSyncTab` 组件及设置页（`/igo-settings`）中的 WebDAV 选项卡、前端服务 API（`getWebDAV`、`saveWebDAV`、`syncWebDAV`）及对应中英文 i18n 文案。
4. **与平台存储隔离**：Wavelet 自身的基础设施对象存储驱动（`plugins/infra/storage/objectstore/webdav.go`）保持独立存在且不受影响。

## Alternatives considered

- **保留 WebDAV 用于备份文件导出/导入传输** — 平台已有标准文件存储与上传下载能力，备份导入导出通过浏览器原生下载与上传即可完成，无需强依赖第三方 WebDAV 凭据。
- **保留表结构与接口仅标记为 Deprecated** — 插件尚处于重构迁移阶段，保留死代码与冗余表增加认知开销与维护负担，直接彻底剪除最清晰。

## Consequences

- **收益**：
  - 代码量减少，消除 1 个实体、1 张数据表、3 个 API 路由及对应的前端 UI 组件与网络状态。
  - 规避了在数据库中存储第三方 WebDAV 账密的安全审计隐患。
  - 前端系统设置页更加聚焦于运行与网络参数、协议模板覆盖与备份恢复。
- **代价**：
  - 用户无法直接将设置与数据通过第三方 WebDAV 云盘（如坚果云/Nextcloud）与历史 Avalonia 客户端进行跨端双向自动同步。若需迁移历史数据，后续统一使用基于 JSON 的本地数据备份导入还原。

## Verification

- `go test -v ./igo-lib/plugins/igo/...`：全部通过，路由清单自 50 条修正为 47 条，表契约校验与 Goose 迁移验证通过。
- `bun run check` 与 `bun run build`：前端类型检查与编译通过。
