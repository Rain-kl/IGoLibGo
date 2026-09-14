# Agent Note: IGoLibrary 下游插件落点与 API 前缀

Status: implemented

## Problem

IGoLibrary-Ex 是单机 Avalonia 客户端。要挂到 Wavelet 上，必须先定插件落点、HTTP 前缀和多租户边界，否则后续表、任务、前端会各写一套路径。

## Decision

IGoLibrary 业务作为单个 Cordis 插件 `igo` 放在 `backend/igo-lib/plugins/igo`。所有 HTTP 都挂在 `/api/v1/igo/` 下，由 `contracts.AuthService` 鉴权，数据按 Wavelet `user_id` 隔离。手机控制、cloudflared、通知渠道配置和桌面专属能力不进入这个插件。

阶段 1 只注册路由和 DTO，Service 返回 `not_implemented`（HTTP 501）。实体、TraceInt 客户端和协调器尚未落地。

## Package 拓扑

- 入口：`backend/igo-lib/plugins/igo/plugin.go`，在 `cmd/app.go` 于 domain 插件之后、`driver_http` 之前 `app.Use(igo.New())`
- 调用方只依赖 `core`、`contracts`、`pkg/response`；never import `plugins/domain/*`
- 表前缀约定 `w_igo_*`（阶段 2 才建）
- 模板 `custom_example` 仍留在 `backend/igo-lib/plugins/custom_example`，不注册进 App

## Alternatives considered

- **按菜单拆成多个插件（grab / occupy / checkin）** — 共享 Cookie、场馆锁定和 TraceInt 传输层，拆开会逼出循环依赖或再做一个 kernel 插件。单插件内按 controller 文件切分即可。
- **把代码放进 `backend/plugins/domain/`** — 那是上游通用域。IGoLibrary 是下游产品，must 留在 `backend/igo-lib/`。
- **复用手机控制那套 token 查询串做鉴权** — Web 已有 Session；再发明一套会绕过平台风控。桌面本地 HTTP 不迁。

## Consequences

- **收益**：路径、信封、鉴权从第一天就和 wavelet 其它 API 一致；前端可以按清单 mock/对接。
- **代价与已知上限**：501 桩在阶段 3 填实前不能跑真实抢座。若 TraceInt 协议按学校分叉到必须独立部署，再把传输层抽成可替换端口，而不是先拆插件。

## Verification

`go test ./igo-lib/plugins/igo/...` 核对 50 条路由和 501 信封。`go test ./cmd/ -run TestNewWaveletAppProfiles` 插件数含 `igo`。
