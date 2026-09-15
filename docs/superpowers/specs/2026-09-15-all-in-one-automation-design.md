# 一条龙自动化 (All-in-One Automation) 架构与技术设计规范

## 1. 背景与目标

为了满足用户一键完成全流程自动化占座与远程签到的需求，并在多账号场景下方便管理与调度，系统新增「一条龙 (All-in-One Automation)」子系统：
- **卡片化配置管理**：用户在 Web 控制台新增/编辑多个自动化配置卡片，每个卡片拥有独立英数字 ID，并可绑定独立的 TraceInt 账号凭据、目标场馆、目标座位及可选的自动签到打卡参数。
- **全流程原子执行引擎**：提供「立即执行」流程，依序执行「账号凭据有效性检查 $\rightarrow$ 目标座位可用性检查（占用则中断） $\rightarrow$ 自动占座 $\rightarrow$ 自动签到（若开启）」。如遇凭据过期，提供交互式补录引导。
- **消息网关 Bot 深度对接**：在 Telegram / QQ 等机器人渠道提供快捷指令交互（`/help`, `/start`, `/show`, `/run [配置ID]`, `/cancel`），支持凭据过期时的多轮会话状态暂存与自然引导回复。

---

## 2. 系统架构与模块划分

```
                               ┌─────────────────────────────┐
                               │   Web 控制台 (/pipeline)     │
                               │ 卡片管理 / 立即执行 / 凭据补录 │
                               └──────────────┬──────────────┘
                                              │ HTTP REST APIs
                                              ▼
┌──────────────────────┐       ┌─────────────────────────────┐
│ 消息网关 (msg_gateway) │       │   IGo 插件 (backend/igo-lib) │
│ - QQ / Telegram Bot  │──────▶│ - PipelineController        │
│ - 指令解析与状态机     │ Event │ - PipelineService           │
│ - 多轮交互会话缓存     │  Bus  │ - PipelineDAO / Entity      │
└──────────────────────┘       └──────────────┬──────────────┘
                                              │ TraceInt API & Check-in API
                                              ▼
                               ┌─────────────────────────────┐
                               │    TraceInt 图书馆云平台     │
                               │  GraphQL 占座 / 蓝牙模拟打卡  │
                               └─────────────────────────────┘
```

---

## 3. 数据持久化设计

### 3.1 数据库表结构 (`w_igo_pipeline_configs`)

双数据库方言（PostgreSQL 与 SQLite）兼容，纳入 `igo` 插件的 Goose 迁移：

```sql
CREATE TABLE w_igo_pipeline_configs (
    id VARCHAR(64) PRIMARY KEY,                   -- 全局唯一配置ID (正则: ^[a-zA-Z0-9_-]+$)
    user_id BIGINT NOT NULL,                      -- 归属 Wavelet 用户 ID
    name VARCHAR(128) NOT NULL,                   -- 配置展示名称
    cookie TEXT NOT NULL,                         -- TraceInt 登录凭据/Cookie (独立账号)
    cookie_expires_at TIMESTAMP NULL,             -- Cookie 预计过期时间
    library_id INT NOT NULL,                      -- 目标场馆 ID
    library_name VARCHAR(128) NOT NULL,           -- 目标场馆名称
    floor VARCHAR(32) NOT NULL,                   -- 场馆楼层
    seat_key VARCHAR(64) NOT NULL,                -- 目标座位 Key
    seat_name VARCHAR(64) NOT NULL,               -- 目标座位名称
    auto_checkin BOOLEAN NOT NULL DEFAULT FALSE,  -- 是否启用自动签到
    checkin_token TEXT NULL,                      -- 签到 session token (wechatSESS_ID)
    checkin_expires_at TIMESTAMP NULL,            -- 签到凭据过期时间
    beacon_uuid VARCHAR(64) NULL,                 -- 签到蓝牙 Beacon UUID
    major INT NOT NULL DEFAULT 0,                 -- 签到 Major
    minor INT NOT NULL DEFAULT 0,                 -- 签到 Minor
    latitude VARCHAR(32) NULL,                    -- 签到经度坐标
    longitude VARCHAR(32) NULL,                   -- 签到纬度坐标
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_igo_pipeline_configs_user ON w_igo_pipeline_configs(user_id);
```

---

## 4. API 契约设计 (`/api/v1/igo/pipeline`)

遵循项目 RESTful 规范与标准信封：

1. **获取全部配置卡片列表**
   - `GET /api/v1/igo/pipeline/configs`
   - 响应：`{ "data": [ PipelineConfigDTO... ] }`
2. **获取单个配置卡片详情**
   - `GET /api/v1/igo/pipeline/configs/:id`
   - 响应：`{ "data": PipelineConfigDTO }`
3. **创建新配置卡片**
   - `POST /api/v1/igo/pipeline/configs`
   - 请求体：`CreatePipelineConfigRequest`（校验 ID 唯一性与英数字正则）
   - 响应：HTTP 201 `{ "data": PipelineConfigDTO }`
4. **更新配置卡片**
   - `PUT /api/v1/igo/pipeline/configs/:id`
   - 请求体：`UpdatePipelineConfigRequest`
   - 响应：`{ "data": PipelineConfigDTO }`
5. **删除配置卡片**
   - `DELETE /api/v1/igo/pipeline/configs/:id`
   - 响应：HTTP 204
6. **立即执行一条龙任务**
   - `POST /api/v1/igo/pipeline/configs/:id/run`
   - 请求体（可选）：`RunPipelineRequest`（支持在前端检测到过期时一并带入新更新的 Cookie / 签到 Token）
   - 响应：`{ "data": PipelineRunResultDTO }`
7. **凭据预检与场馆座位加载辅助接口**
   - `POST /api/v1/igo/pipeline/helper/verify-session`：传入 Cookie 探测有效性并拉取该账户的场馆列表。
   - `POST /api/v1/igo/pipeline/helper/library-layout`：传入 Cookie + library_id 拉取指定场馆的实时座位状态列表。
   - `POST /api/v1/igo/pipeline/helper/verify-checkin`：传入 checkin_token/code 探测签到授权有效性并拉取可选 Beacon 设备列表。

---

## 5. 一条龙执行流程与异常处理 (Pipeline Execution)

```
                       [ 触发执行: RunPipeline(config_id) ]
                                       │
                                       ▼
                     [ 1. 读取卡片配置并校验权限 ]
                                       │
                                       ▼
                     [ 2. 检查 TraceInt 登录凭据有效性 ]
                                       │
                 ┌─────────────────────┴─────────────────────┐
                 │ 有效                                       │ 过期/未授权
                 ▼                                           ▼
   [ 3. 检查自动签到凭据 (若开启) ]                   [ 返回 NeedAuth: LOGIN ]
                 │                                   [ 附带微信登录 OAuth 授权链接 ]
         ┌───────┴───────┐
         │ 有效           │ 过期/未授权
         ▼               ▼
   [ 4. 检查座位可用性 ]   [ 返回 NeedAuth: CHECKIN ]
         │               [ 附带微信签到 OAuth 授权链接 ]
   ┌─────┴─────┐
   │ 空闲       │ 已被占用
   ▼           ▼
[ 5. 自动占座 ]  [ 终止并返回: "目标座位已被占用" ]
   │
   ├─ 占座失败 ─▶ [ 终止并返回失败错误原因 ]
   │
   ▼ 占座成功
[ 6. 自动签到 (若开启) ]
   │
   ├─ 签到失败 ─▶ [ 返回: "占座成功，但自动签到打卡失败: {err}" ]
   │
   ▼ 签到成功
[ 7. 返回全部成功，记录流水日志并触发系统推送通知 ]
```

---

## 6. 消息网关 Bot 指令与交互状态机

### 6.1 指令定义
- `/help` 或 `/start`：输出帮助说明与所有可用指令。
- `/show`：列出当前绑定用户的所有一条龙配置卡片（ID、名称、场馆、座位、自动签到开关、当前凭据状态）。
- `/run [配置 ID]`：触发指定配置的一条龙执行。
- `/cancel`：随时取消正在进行的凭据补录会话。

### 6.2 多轮交互状态机 (基于 CacheService，TTL = 5 分钟)
- **状态键名**：`igo_bot_session:{channel_id}:{platform_user_id}`
- **状态结构**：
  ```json
  {
    "config_id": "seat01",
    "step": "waiting_login_auth",
    "expire_at": 1726390000
  }
  ```
- **会话流转**：
  1. 用户发送 `/run seat01`。
  2. 若登录凭据已过期：
     - 保存 `step = waiting_login_auth`；
     - 回复登录授权链接：`https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2Fgraphql%3FoperationName%3Dindex%26query%3Dquery%257BuserAuth%257BtongJi%257Brank%257D%257D%257D&response_type=code&scope=snsapi_userinfo&state=1&connect_redirect=1#wechat_redirect`；
     - 提示：「请在微信中打开上述链接授权，并将跳转后的链接或登录 Cookie 发送给机器人：」。
  3. 用户回复内容：
     - 若包含 `code=` 或为 Cookie：提取并校验 Cookie，持久化回写至 `w_igo_pipeline_configs`。
     - 若该卡片开启了自动签到且签到凭据也已过期：
       - 更新 `step = waiting_checkin_auth`；
       - 回复签到授权链接：`https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2FwxApp%2FwechatAuth.html%3Fr%3Dhttps%253A%252F%252Fweb.traceint.com%252Fweb%252Findex.html&response_type=code&scope=snsapi_base&state=1#wechat_redirect`；
       - 提示：「登录凭据已更新！请继续在微信中打开上述签到链接，并将复制的签到授权链接发送给机器人：」。
     - 若无需签到凭据或签到凭据有效：清除会话状态，立即触发 Pipeline 执行并将结果回复给用户。
  4. 用户处于 `waiting_checkin_auth` 时回复内容：
     - 换取 `wechatSESS_ID`，持久化回写配置，清除会话状态，立即触发 Pipeline 执行并输出结果。

---

## 7. 前端 UI 与国际化设计 (`frontend/app/(main)/pipeline`)

- **页面入口**：侧边栏「自动化」（路由 `/pipeline`），支持中英双语国际化 (`zh-CN.json` / `en.json`)。
- **视觉风格**：采用与任务管理一致的现代化渐变卡片与精致边框动效。
- **核心组件拆分**：
  - `pipeline-list.tsx`：卡片网格与状态展示。
  - `pipeline-card.tsx`：单个卡片，包含操作菜单与立即执行按钮。
  - `pipeline-dialog.tsx`：新增/编辑卡片弹窗向导。
  - `pipeline-auth-modal.tsx`：执行时凭据失效的就地快速扫码与粘贴链接弹窗。
  - `pipeline-result-dialog.tsx`：执行结果明细与步骤进度展示。

---

## 8. 测试与质量保证

- **单元与集成测试**：
  - `pipeline_dao_test.go`：测试卡片增删改查、ID 唯一性约束与多用户隔离。
  - `pipeline_service_test.go`：测试座位可用性校验拦截、占座及签到完整调用链及 Mock 驱动。
  - `bot_command_test.go`：测试 `/help`, `/show`, `/run`, 多轮交互状态机状态切换与超时。
- **代码质量门禁**：开发完成后依次通过 `make code-check` 与 `make format`，前端通过 `bun run lint` 与 `bun run format`。
