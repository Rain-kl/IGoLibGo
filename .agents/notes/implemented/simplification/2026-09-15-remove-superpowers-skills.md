# Agent Note: 裁撤 Superpowers 流程治理技能体系

Status: implemented

## Problem

此前 Wavelet 项目在 `.agents/skills/` 引入了来自 `obra/superpowers` 的 8 个流程治理类技能（包括 `using-superpowers`、`brainstorming`、`systematic-debugging`、`writing-plans`、`executing-plans`、`verification-before-completion`、`receiving-code-review`、`using-git-worktrees`）。

在实际开发与协作演进中，暴露出以下问题：
1. **元技能与本地环境规范重叠冗余**：诸如计划拟定、完成前验证、代码审查等原则已在项目顶层根规则 `AGENTS.md`、Antigravity 平台自带工作流（如 Planning Mode、Verification Checks）以及团队工程规范中深度固化，外部流程技能带来了多层指令重叠与 Prompt 膨胀；
2. **上下文开销与认知负担**：过多的流程控制技能占用 Agent 技能检索空间，稀释了垂直业务技能（如 `wv-*` 核心框架契约、`clickhouse-io`、`golang-patterns` 等）的注意力权重；
3. **维护成本与外部依赖漂移**：流程规范应由项目统一维护，外部通用的流程技能容易与 Wavelet 的 Cordis 微内核开发规范及 Git 上下游守则产生认知冲突。

## Decision

彻底裁撤本地仓库中 8 个 Superpowers 流程治理技能，并精简相关运维与索引配置：
1. **清理物理目录**：移除 `.agents/skills/` 下的 `using-superpowers`、`brainstorming`、`systematic-debugging`、`writing-plans`、`executing-plans`、`verification-before-completion`、`receiving-code-review`、`using-git-worktrees` 目录；
2. **更新项目治理规范**：从 `AGENTS.md` 的技能索引表中剔除上述 8 个条目，流程治理全面交由 `AGENTS.md` 原生 Guardrails 与平台上下文规范；
3. **重构技能索引与工具脚本**：
   - 更新 `.agents/skills/README.md`，技能总数由 45 个调整为 37 个，分类精简为四大类（ECC 社区体系 25 个、DeepSeek 决策沉淀 1 个、Autoresearch 1 个、Wavelet 自研业务 10 个）；
   - 更新 `scripts/update_skills.sh`，移除 `SP_REPO`、`SP_SKILLS`、`--all-sp` 参数及对应更新逻辑，保留 ECC 与其余生态工具链。

## Alternatives considered

- **方案 A：保留技能目录但从 `AGENTS.md` 中解绑**：放弃。保留未被声明和维护的孤立技能目录会增加仓库冗余度，且后续执行离线更新脚本时仍会拉取无用内容。
- **方案 B：重命名为 `wv-` 前缀并改造内部 Prompt**：放弃。Wavelet 的流程规范已直接融入 `AGENTS.md` 顶层指令中，无需额外包装成独立的 Cordis 技能，直接精简是维护成本最低、执行确定性最高的方式。

## Consequences

- **收益**：
  - Agent 技能空间显著净化，聚焦于业务架构、微内核扩展与工程实现；
  - 减少 Prompt 上下文占用，消除流程指令冲突；
  - 简化了 `scripts/update_skills.sh` 的更新链路与测试维护工作量。
- **代价**：
  - 外部开发者若依赖 `obra/superpowers` 的特定命令/提示词模板，需参考项目根目录 `AGENTS.md` 与历史提交规范。
