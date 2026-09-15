# Agent Note: 一条龙自动化 (All-in-One Automation Pipeline) 架构与消息网关交互设计

Status: implemented

## Problem

用户在 IGoLibGo 中使用占座与远程打卡时，需要在多个页面（场馆锁定、抢座/占座、远程签到）分别操作；且如果拥有多个学号/账户（如多用户协同占座），缺乏独立的多账号卡片化管理与一键闭环能力。此外，缺乏外部机器人（Telegram / QQ Bot）多轮对话交互与即时触发执行能力。

## Decision

1. **多账号独立卡片式存储**：
   - 采用独立表 `igo_pipeline_configs`，主键采用全局唯一英文字符串 `id VARCHAR(64)`（正则 `^[a-zA-Z0-9_-]+$`），同时记录所有者 `user_id`，杜绝越权访问。
   - 每个卡片独立保存 TraceInt 登录 Cookie 与过期时间、目标场馆/楼层/座位 Key 与名称、自动签到开关与打卡授权 Token、Beacon 参数。

2. **执行引擎闭环**：
   - 执行前首先探测 TraceInt 登录凭据有效性；若过期返回 `need_auth: "LOGIN"` 与微信 OAuth 授权链接。
   - 若开启自动签到，检查打卡授权 Token 有效性；若过期返回 `need_auth: "CHECKIN"` 与签到授权链接。
   - 查座防冲突：先拉取实时布局核验座位占用状态；若已被他人占用则优雅退出并提示。
   - 占座与签到串行执行：占座成功后自动计算服务器时间并完成 Beacon 模拟打卡。

3. **消息网关与 Bot 交互**：
   - 支持 `/help` / `/start`、`/show`、`/run [配置 ID]` 指令。
   - 当凭据过期时，Bot 自动通过 `contracts.CacheService` 建立 5 分钟交互式补录会话（支持 `/cancel` 退出），用户发送微信链接或 Code 后自动换取凭据并立即无缝触发执行。

4. **前端卡片式大盘与向导**：
   - 路由 `/pipeline`，侧边栏「自动化」。
   - 提供 3 步向导式卡片创建/编辑 Dialog、凭据失效引导重录 Modal、执行结果结果卡片。

## Alternatives considered

- **复用既有单例 `Session` 和 `Venue` 锁定** — 无法支持多卡片绑定不同学号/不同场馆座位。因此必须独立卡片存储。
- **纯粹长轮询等待用户扫码** — 微信 OAuth 在桌面和外部 Bot 下更适合让用户复制重定向链接中的 Code，兼顾安全与跨平台可用性。

## Consequences

- **收益**：多账号独立隔离，一键执行查座、占座、签到全流程；Web 控制台与 IM 机器人双端无缝联动。
- **校验**：全后端单元测试通过，Cordis 架构零依赖违规，前端编译和静态导出 100% 成功。
