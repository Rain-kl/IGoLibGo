# 签到信息管理与账户凭证拆分

## 1. 背景与目标

远程签到和一条龙现在把「谁在打卡」和「用哪组 Beacon 打卡」缠在一起：

- `/checkin` 每个 Wavelet 用户只有一份 `igo_checkin_sessions` Token，Beacon 按场馆塞在 `igo_settings` JSON。
- 一条龙卡片把 Cookie、签到 Token、Beacon 全部内嵌。关掉自动签到后，手动打卡只能再填一遍。
- 双号场景（1 号占座、2 号打卡）会失败：签到信息若带着 1 号的 `wechatSESS_ID`，不能给 2 号用。Beacon 却可以共用。

本轮把凭证收到账户表，把模拟参数收到签到信息表。一条龙和 `/checkin` 都按 ID 引用。抢座/占座/捡漏/明日预约继续使用 `igo_sessions`，本轮不改。

成功标准：

- `/checkin` 能增删改签到信息；打卡时选账户 + 签到信息，不再手填全套参数。
- 一条龙创建不再内嵌 Cookie / Token / Beacon；自动打卡改为选签到信息，可选另一个打卡账户。
- 占座号 ≠ 打卡号时，使用打卡账户的 Token + 共用签到信息的 Beacon。
- 打卡失败且凭证失效时，引导更新该账户的签到凭证，而不是改签到信息。

## 2. 范围

**做：**

- 下游 `igo` 插件新增 `igo_accounts`、`igo_checkin_infos`。
- 一条龙加引用列并改执行/Bot 补录路径。
- `/checkin` 整页改为账户列表 + 签到信息列表 + 按引用打卡。
- Goose 建表 + 启动幂等回填。

**不做：**

- 不改 `backend/core/`、`backend/pkg/`、上游 `plugins/`。
- 不改抢座/占座/捡漏/明日预约对 `igo_sessions` 的读写。
- 不删一条龙旧列（`cookie`、`checkin_token`、`beacon_*`），只停止新写入。
- 不删旧 `/checkin/session`、`/checkin/from-code`、`/checkin/profiles` 后端路由（前端停用；下轮再删）。
- 不覆盖工作区里未提交的 pipeline 其它改动。

## 3. 数据模型

无物理外键。主键 snowflake `BIGINT`。按 `user_id` 隔离。Go 零值与库默认值一致。

### 3.1 `igo_accounts`

凭证和账号资料。同一 Wavelet 用户多条。

| 列 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `id` | BIGINT PK | | snowflake |
| `user_id` | BIGINT NOT NULL | | 索引 `idx_igo_accounts_user` |
| `name` | VARCHAR(128) NOT NULL | | 展示名 |
| `cookie` | TEXT NOT NULL | `''` | TraceInt 占座登录；纯打卡号可空串 |
| `cookie_expires_at` | TIMESTAMP NULL | | |
| `cookie_source` | VARCHAR(32) NOT NULL | `''` | |
| `checkin_token` | TEXT NOT NULL | `''` | 微信 `wechatSESS_ID`；纯占座号可空串 |
| `checkin_expires_at` | TIMESTAMP NULL | | |
| `nickname` | VARCHAR(128) NOT NULL | `''` | devices 回填 |
| `school` | VARCHAR(128) NOT NULL | `''` | |
| `student_name` | VARCHAR(128) NOT NULL | `''` | |
| `student_number` | VARCHAR(64) NOT NULL | `''` | |
| `created_at` / `updated_at` | TIMESTAMP NOT NULL | | |

回填时按 `(user_id, cookie)` 去重：非空 Cookie 相同则复用账户。空 Cookie 的纯打卡号不去重合并。

### 3.2 `igo_checkin_infos`

只有模拟参数，无凭证。字段均可空，真正打卡前再校验。

| 列 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `id` | BIGINT PK | | snowflake |
| `user_id` | BIGINT NOT NULL | | 索引 `idx_igo_checkin_infos_user` |
| `name` | VARCHAR(128) NOT NULL | | 展示名 |
| `beacon_uuid` | VARCHAR(64) NOT NULL | `''` | |
| `major` | INT NOT NULL | `0` | 有效范围 0–65535 |
| `minor` | INT NOT NULL | `0` | 有效范围 0–65535 |
| `latitude` | VARCHAR(32) NOT NULL | `''` | |
| `longitude` | VARCHAR(32) NOT NULL | `''` | |
| `created_at` / `updated_at` | TIMESTAMP NOT NULL | | |

不绑场馆。`expected_library_id` 来自调用上下文：一条龙用卡片场馆；手动打卡由请求传入。

### 3.3 `igo_pipeline_configs` 新增列

| 列 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `account_id` | BIGINT NOT NULL | `0` | 占座账户 |
| `checkin_account_id` | BIGINT NOT NULL | `0` | 打卡账户；`0` = 与占座同号 |
| `checkin_info_id` | BIGINT NOT NULL | `0` | 签到信息；`0` 且 `auto_checkin=false` = 不自动打卡 |

旧列 `cookie`、`checkin_token`、`beacon_uuid`、`major`、`minor`、`latitude`、`longitude` 本轮保留只读。新的创建/更新只写引用列和座位字段；执行读账户表和签到信息表。

`auto_checkin=true` 时 `checkin_info_id` 必须非 0。保存时不要求账户已有 Token、不要求签到信息 Beacon 已填完；执行时再查。

## 4. 回填

Goose `00003` 只建表和加列。数据搬迁在 `igo` 插件 `Apply` 里做幂等回填（SQLite / PostgreSQL 同一套 Go）：

1. 每个有非空 Cookie 的 `igo_sessions` → 该用户一条名为「默认」的账户（该用户已有任意账户则跳过，避免重复点）。
2. `igo_checkin_sessions` 的 Token：挂到该用户名为「默认」的账户，没有则取最早创建的账户；若该用户还没有任何账户，则新建「默认」且只写入签到 Token。目标账户 `checkin_token` 已非空则不覆盖。
3. 每张 `account_id=0` 的一条龙：用 `(user_id, cookie)` 复用或新建账户（名称用卡片 `name`）；Beacon 任一非空则新建签到信息（名称「{卡片名} 签到」），回写 `account_id`、`checkin_info_id`；`auto_checkin` 保持原值。卡片上的 `checkin_token` 写入占座账户（无法推断双号，用户事后可改 `checkin_account_id`）。
4. `igo_settings.CheckInProfiles`：每个场馆 profile 生成一条签到信息，名称用 `library_name`；不删 JSON，前端不再读。

回填失败只打日志，不阻断启动。一条龙执行若仍看到 `account_id=0`，对该行再跑一次步骤 3，然后继续；仍为 0 则返回 `need_auth: LOGIN`。

## 5. API

前缀 `/api/v1/igo`，登录鉴权，信封与现有 `igo` 一致。ID 在 JSON 里 `string`。不回传 Cookie / Token 明文，只给 `has_cookie`、`cookie_masked`、`has_checkin_token`。

### 5.1 账户

| 方法 | 路径 | 成功 |
|---|---|---|
| GET | `/accounts` | 200 列表 |
| POST | `/accounts` | 201 + Location |
| GET | `/accounts/:id` | 200 |
| PUT | `/accounts/:id` | 200 |
| DELETE | `/accounts/:id` | 204；被一条龙 `account_id` 或 `checkin_account_id` 引用 → 409 `conflict` |
| POST | `/accounts/:id/login` | 200；body 与现有会话登录相同：`code` 或 Cookie 原文 |
| POST | `/accounts/:id/checkin-auth` | 200；写入签到凭证，拉 devices，回填学号姓名 |

`POST /accounts` body：`name` 必填；`cookie`、`checkin_token`（或授权链接/code）可选。跨用户 ID → 404。

`checkin-auth` 成功响应额外带 `device`（与现有 `CheckInDeviceResponse` 相同，含 `beacon_uuids`）。前端若正在编辑 UUID 为空的签到信息，用该列表回填。

### 5.2 签到信息

| 方法 | 路径 | 成功 |
|---|---|---|
| GET | `/checkin/infos` | 200 列表 |
| POST | `/checkin/infos` | 201 + Location |
| GET | `/checkin/infos/:id` | 200 |
| PUT | `/checkin/infos/:id` | 200 |
| DELETE | `/checkin/infos/:id` | 204；被一条龙 `checkin_info_id` 引用 → 409 |
| POST | `/checkin/infos/:id/sign` | 200 |

`POST .../sign` body：

```json
{
  "account_id": "1",
  "expected_library_id": 101,
  "expected_library_name": "主馆"
}
```

`account_id`、`expected_library_id` 必填。服务端用该账户 `checkin_token` + 该条 Beacon 调用现有 `SignCheckIn`。

失败：

| 条件 | HTTP | code |
|---|---|---|
| Token 空或 `GetCheckInDevices` / 打卡判定失效 | 409 | `need_auth`：data 含 `need_auth: "CHECKIN"`、`account_id`、授权 URL。一条龙 `RunPipeline` 仍用 HTTP 200 + `PipelineRunResult.need_auth`，不改成 409 |
| UUID 无法 normalize，或经纬度空 | 400 | `validation_error` 文案要求补签到信息 |
| major/minor 越界 | 400 | `validation_error` |
| 账户或签到信息不属于当前用户 | 404 | `not_found` |

旧路由 `/checkin/session`、`/checkin/from-code`、`/checkin/devices`、`/checkin/sign`、`/checkin/profiles` 本轮保留实现，前端不再调用。

### 5.3 一条龙

`CreatePipelineConfigRequest` / `UpdatePipelineConfigRequest`：

- 新增 `account_id`（创建必填）、`checkin_account_id`、`checkin_info_id`。
- 不再接收 `cookie`、`checkin_token`、`beacon_uuid`、`major`、`minor`、`latitude`、`longitude`。创建向导改为先选/建账户。
- `auto_checkin=true` 且 `checkin_info_id=0` → 400。
- 引用的账户/签到信息必须属于同一 `user_id`，否则 404。

`PipelineConfigDTO` 增加三个 ID、占座/打卡账户摘要（名称、`has_cookie`、`has_checkin_token`）、签到信息摘要（名称、`beacon_uuid`）。创建/更新请求不再包含 Cookie、Token、Beacon。DTO 仍可带只读的 `beacon_uuid` 等展示字段（从签到信息 join），前端不把它们当写入来源。

执行：

1. 占座：`account_id.cookie`，失效 → `need_auth: LOGIN`（补到占座账户）。
2. 自动打卡：解析打卡账户（`checkin_account_id` 或占座账户）的 Token + `checkin_info_id` 的 Beacon + 卡片 `library_id`。Token 失效 → `need_auth: CHECKIN`（补到打卡账户）。Beacon 不完整 → 不发起打卡，结果里说明去补签到信息；占座成功仍算卡片执行成功。

Bot：`/run` 不变。`igo.login_auth` 把 Cookie 写占座账户。`igo.checkin_auth` 把 Token 写打卡账户。

`helper/verify-session`、`helper/verify-checkin` 保留，供创建账户时探测。

## 6. 前端

### 6.1 `/checkin`

替换现有单例会话卡 + 场馆 profile 打卡面板，改为：

1. 账户列表：名称、学号姓名、Cookie/签到凭证状态；录入登录、录入签到凭证、删除。
2. 签到信息列表：名称与 Beacon 摘要；新增时可空；编辑时可填 UUID/Major/Minor/经纬度。
3. 打卡条：下拉账户、下拉签到信息、场馆（默认当前锁定场馆 `igo_venues`，否则手动填 `expected_library_id`）。提交 `POST /checkin/infos/:id/sign`。
4. 若 sign 返回 `need_auth: CHECKIN`，打开该账户的签到授权，不修改签到信息。
5. 账户 `checkin-auth` 返回 `beacon_uuids` 且当前编辑中的签到信息 UUID 为空时，自动填入列表第一项（可改）。

文案走 `next-intl`（`zh-CN` / `en`）。页面容器、标题、a11y 规则与现有 `igo` 页一致。

### 6.2 `/pipeline`

创建向导第 1 步：选已有占座账户或粘贴 Cookie/链接现场建账户。打开自动打卡：选签到信息（可跳转 `/checkin` 新建）+ 可选另一打卡账户。第 2 步仍用占座账户 Cookie 拉场馆座位。编辑弹窗同样改为引用，不再出现 Beacon 输入。E2E 按此改断言。

抢座/占座等页不改。

## 7. 错误处理

- 业务错误用 `consts.CodedError`，Handler 映射现有信封。
- 底层 TraceInt 错误在 service 边界打日志，客户端只看语义文案。
- 删除冲突 409，文案带引用的一条龙 ID 列表（最多列出若干条，避免超长）。

## 8. 测试

后端（`backend/igo-lib/plugins/igo`）：

- 账户 CRUD、`user_id` 隔离、删除被引用 409。
- 签到信息 CRUD、删除被引用 409。
- `sign`：Token + Beacon 组合；Token 失效返回 CHECKIN + account_id；Beacon 不完整 400。
- 一条龙：占座账户 ≠ 打卡账户时 Cookie/Token 分别取自两行；`checkin_account_id=0` 时回落到占座账户。
- 回填幂等：跑两次 ID 不变、不复制账户。

前端 E2E：

- `/checkin` 创建签到信息、选账户打卡、凭证失败走更新账户。
- 一条龙向导选账户/签到信息；打开自动打卡不再出现 Beacon 手填框。

不改抢座相关用例。

## 9. 文件落点

全部在下游 `igo` 插件与产品前端：

- `backend/igo-lib/plugins/igo/migrations/{sqlite,postgres}/00003_accounts_and_checkin_infos.sql`
- `model/entity`、`model/do`、`dao`、`service`、`controller` 按实体分文件：`account.go`、`checkin_info.go`（签到信息 CRUD/sign）；`pipeline.go` 只加引用解析
- `plugin.go`：`Apply` 调用幂等回填
- `frontend/lib/services/igo/` 账户与签到信息 client
- `frontend/app/(main)/checkin/` 与 `components/igo/checkin/`
- `frontend/app/(main)/pipeline/components/` 向导改引用
- `frontend/messages/{zh-CN,en}.json`
- `frontend/e2e/pipeline.spec.ts` 与新建 checkin e2e

## 10. 风险

- Cookie 仍有两处：`igo_accounts` 与 `igo_sessions`。本轮接受。以后若统一，让「当前会话」指向某个 `account_id`。
- 回填把卡片 `checkin_token` 写到占座账户，历史双号卡片不会自动拆号，需用户改 `checkin_account_id`。
- 工作区未提交的 pipeline 改动：实现时只加引用，不重写无关执行逻辑。
