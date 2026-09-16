<!--
  【维护提醒】
  1. 现行法律契约：此文件在代码重构、改名或路径变更时，必须同批原子就地更新事实。
  2. 源码物理绑定：请在受此决定保护的核心代码入口处标注：
     // Note: <简述> — 见 .agents/notes/implemented/<class>/<filename>.md
  3. 时态门禁：严禁出现 ## Proposal、## Plan、## Acceptance criteria 等计划态标题。
-->

# Agent Note: 签到信息与账户凭证拆分

Status: implemented

## Problem

远程签到和一条龙把「谁在打卡」和「用哪组 Beacon 打卡」缠在一起：`igo_checkin_sessions` 每用户一份 Token，Beacon 按场馆塞在 settings JSON，一条龙卡片再内嵌 Cookie、Token、Beacon。关掉自动签到后只能再填一遍；1 号占座打卡后把凭证写进签到信息，2 号无法复用 Beacon 却会误用 1 号 Token。

## Decision

凭证跟号走，Beacon 跟场馆模拟参数走：

1. **`igo_accounts`**：名称、TraceInt Cookie、微信签到 Token、过期时间、学号姓名快照。同一 Wavelet 用户多条。纯占座号或纯打卡号允许对应凭证为空。
2. **`igo_checkin_infos`**：名称与 UUID/Major/Minor/经纬度。无凭证，不绑场馆。字段可空，真正打卡前再校验。
3. **`igo_pipeline_configs`** 只存 `account_id`（占座）、`checkin_account_id`（0 表示与占座同号）、`checkin_info_id`（0 且未开自动打卡表示不打卡）。旧 Cookie/Token/Beacon 列保留只读。
4. **`/checkin`** 改为账户列表 + 签到信息列表 + 按引用打卡。`POST /checkin/infos/:id/sign` 用账户 Token + 签到信息 Beacon。Token 失效返回 HTTP 409 `need_auth`，引导更新该账户凭证。
5. **一条龙执行** 占座用占座账户 Cookie，自动打卡用打卡账户 Token + 签到信息 Beacon + 卡片场馆。Bot 补登录/补签到写到对应账户。`RunPipeline` 仍用 HTTP 200 + `need_auth`。
6. **Goose `00003` 建表加列**；`Service.BackfillAccountsAndCheckinInfos` 幂等搬迁 session、卡片、settings profile。`igo_sessions` 仍给抢座/占座/捡漏/明日预约用。

核心入口：`backend/igo-lib/plugins/igo/model/entity/account.go`。

## Alternatives considered

- **把 `igo_sessions` 扩成 1:N 并合并签到 Token** — 少一张表，但要改最热的会话读取和全部任务节拍，回归面大。
- **只抽签到信息表、凭证仍留在卡片上** — Beacon 可复用，双号仍会串 Token。
- **签到信息带凭证** — 1 号打卡资料无法安全给 2 号用。

## Consequences

- **收益**：一条龙和 `/checkin` 按 ID 引用；占座号可以不等于打卡号；Beacon 只维护一份。
- **代价与已知上限**：Cookie 仍有 `igo_accounts` 与 `igo_sessions` 两处。历史双号卡片回填会把 Token 写到占座账户，需用户改 `checkin_account_id`。旧 `/checkin/session|profiles` 路由本轮保留。
- **校验**：`go test ./igo-lib/plugins/igo/...`；Playwright `e2e/checkin.spec.ts` 与 `e2e/pipeline.spec.ts`。
