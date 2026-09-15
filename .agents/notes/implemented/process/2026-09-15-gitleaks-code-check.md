# Agent Note: code-check 接入 gitleaks 密钥扫描

Status: implemented

## Problem

真实凭据一旦写进测试或提交进 Git，会随仓库与 CI 日志扩散，轮换成本高。仅靠人工审查拦不住「看起来像 mock、其实是线上 Token」的写法。gitleaks 默认的 `telegram-bot-api-token` 规则还要求附近出现 `telegr` 标识符，因此 `{"token": "<bot_id>:<secret>"}` 这种测试赋值会漏报。

## Decision

`make code-check` 在架构检查之后、lint 之前调用 `scripts/security_check.sh`，对工作区做 `gitleaks dir`（扫当前树，不扫 Git 历史）。配置在 `.gitleaks.toml`：`useDefault = true`，并额外注册 `telegram-bot-api-token-raw`（`[0-9]{5,16}:A[A-Za-z0-9_-]{34}`，不要求附近标识符）。脚本启动时用合成 Token 做一次自检，该规则失效则门禁直接失败。测试与配置里的真实密钥由 `AGENTS.md` 明文禁止；文档截断示例（末尾 `...`）与技能文档里的短 dummy Stripe key 走 allowlist。

## Alternatives considered

- **只用 gitleaks 默认规则、不补 raw Telegram 规则** — 默认规则漏报无 `telegr` 上下文的裸 Bot Token，而测试里常见的就是 `Credentials["token"] = "..."`，放弃。
- **TruffleHog / git-secrets** — TruffleHog 更重、默认走 Git 历史与验证器；git-secrets 规则面窄。仓库已是 Go 工具链，gitleaks 单二进制、可钉版本、配置可扩展，选它。
- **`gitleaks git` 扫全历史** — 历史里一旦出现过密钥，门禁会永远红直到改写历史。code-check 拦的是「当前要提交的树」；历史审计另做。
- **不做门禁、只写 AGENTS.md** — 规范拦不住已经写进 `_test.go` 的真实 Token。

## Consequences

- **收益**：当前工作区里的 Telegram Bot Token、私钥、云厂商 Key 等会在 lint 之前失败；默认规则漏掉的裸 Bot Token 也能拦住。
- **代价与已知上限**：开发机/CI 需要 gitleaks（脚本可按 v8.30.1 下载到用户 cache）。文档假阳性靠 allowlist 或 `gitleaks:allow`。不扫 Git 历史，已入库但已从 HEAD 删掉的密钥不会被这道门禁报出。规则或默认配置漏报时，先补 `.gitleaks.toml` 再放行。

## Verification

- `scripts/security_check.sh` 的 canary 必须对 `token: "1234567890:A" + 34 位` 报 exit 1。
- 对含 `{"token": "<8-16 位数字>:A + 34 位"}` 的 `_test.go`，`gitleaks dir -c .gitleaks.toml` 必须打出 `telegram-bot-api-token-raw`。
