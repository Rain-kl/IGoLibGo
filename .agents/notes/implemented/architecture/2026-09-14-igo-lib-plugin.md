# Agent Note: IGoLibrary 下游插件落点与 API 前缀

Status: implemented

## Problem

IGoLibrary-Ex 是单机 Avalonia 客户端。要挂到 Wavelet 上，必须先定插件落点、HTTP 前缀和多租户边界，否则后续表、任务、前端会各写一套路径。

## Decision

IGoLibrary 业务作为单个 Cordis 插件 `igo` 放在 `backend/igo-lib/plugins/igo`。所有 HTTP 都挂在 `/api/v1/igo/` 下，由 `contracts.AuthService` 鉴权，数据按 Wavelet `user_id` 隔离。手机控制、cloudflared、通知渠道配置和桌面专属能力不进入这个插件。

13 张表使用 `igo_*` 前缀（不用平台 `w_`），按 `user_id` 隔离。TraceInt Cookie/GraphQL/签到客户端已接入会话、场馆、预约、协议、设置与四个任务节拍（`igo:tick`）。抢座/占座/捡漏/明日预约成功、任务失败、Cookie 即将过期与会话失效通过 `contracts.PushRegistry` 注册为通知中心内置事件，触发走 `notification:push`。备份导入导出不做。Cookie / WebDAV 密码列目前明文存储。

## Package 拓扑

- 入口：`backend/igo-lib/plugins/igo/plugin.go`，在 `cmd/app.go` 于 domain 插件之后、`driver_http` 之前 `app.Use(igo.New())`
- 调用方只依赖 `core`、`contracts`、`pkg/response`、`pkg/idgen`；never import `plugins/domain/*`
- 表前缀 `igo_*`（不用平台 `w_`），无物理外键，主键为 snowflake `BIGINT`
- 模板 `custom_example` 仍留在 `backend/igo-lib/plugins/custom_example`，不注册进 App

## Alternatives considered

- **按菜单拆成多个插件（grab / occupy / checkin）** — 共享 Cookie、场馆锁定和 TraceInt 传输层，拆开会逼出循环依赖或再做一个 kernel 插件。单插件内按 controller 文件切分即可。
- **把代码放进 `backend/plugins/domain/`** — 那是上游通用域。IGoLibrary 是下游产品，must 留在 `backend/igo-lib/`。
- **复用手机控制那套 token 查询串做鉴权** — Web 已有 Session；再发明一套会绕过平台风控。桌面本地 HTTP 不迁。
- **表名沿用平台 `w_` 前缀** — `w_` 是 Wavelet 官方表命名空间。下游 IGo 表 must 用 `igo_*`，避免和平台表抢前缀。

## Consequences

- **收益**：路径、信封、鉴权从第一天就和 wavelet 其它 API 一致；前端可以按清单 mock/对接。
- **代价与已知上限**：明日预约未接入 TraceInt WebSocket 排队通道，按设定触发时间直接 warmup/save。抢座节拍在 TaskService 不可用时只落库不轮询。推送渠道不在 IGo 设置页配置，must 在平台通知中心绑定；未启用的内置事件不会发信。若 TraceInt 协议按学校分叉到必须独立部署，再把传输层抽成可替换端口，而不是先拆插件。

## Verification

`go test ./igo-lib/plugins/igo/...` 核对路由信封、`igo_*` 表前缀、Goose 落地、user_id 隔离，以及 Cookie 登录后拉场馆。`go test ./cmd/ -run TestNewWaveletAppProfiles` 插件数含 `igo`。
