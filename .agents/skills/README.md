# Wavelet Agent Skills 索引与维护指南

本文档维护当前项目仓库（`Wavelet`）中所有已安装 Agent Skills 的**官方上游仓库索引**、**功能作用说明**以及**非 Git 方式（离线下载拷贝）的更新操作方法**。

---

## 一、技能总览与上游分类

本项目采用**目录源码直入（非 Git Submodule）**方式引入外部技能，确保仓库轻量、自包含且不受子模块指针冲突影响。当前共收录 **45 个技能**，来源分为以下五大类：

| 来源分类 | 技能数量 | 官方仓库地址 | 仓库内路径 |
| :--- | :---: | :--- | :--- |
| **ECC 社区体系** | 25 | `https://github.com/affaan-m/ECC` | `skills/<skill-name>` |
| **Superpowers 开发流程** | 8 | `https://github.com/obra/superpowers` | `skills/<skill-name>` |
| **DeepSeek 决策沉淀** | 1 | `https://github.com/czm15053/write-notes-like-deepseek` | 根目录全包 |
| **自动化目标循环** | 1 | `https://github.com/dave1010/autoresearch` | 根目录全包 |
| **Wavelet 自研/内核业务** | 10 | `https://github.com/Rain-kl/Wavelet` (当前项目) | `.agents/skills/<skill-name>` |

---

## 二、全量 Skills 详细索引表

### 1. Superpowers 开发流程治理体系 (8 个)

上游仓库统一为：[`https://github.com/obra/superpowers`](https://github.com/obra/superpowers)

| Skill 名称 | 上游路径 | 核心作用与职责 | 更新说明 |
| :--- | :--- | :--- | :--- |
| `using-superpowers` | `skills/using-superpowers` | **总开关与元技能**：强制 Agent 治愈懒惰与自作主张，在开始任何任务前优先检索、激活并规范执行匹配的 Skills。 | 覆盖更新 `SKILL.md` 与 `references/` |
| `brainstorming` | `skills/brainstorming` | 需求与方案头脑风暴：在动手编写代码前，与开发者互动探索深层意图、多方案对比与架构权衡（契合 Think Before Coding）。 | 覆盖更新 `SKILL.md` 与子目录 |
| `systematic-debugging` | `skills/systematic-debugging` | 系统化科学排错：严禁病急乱投医乱改代码，强制按“①最小复现 → ②根因链诊断 → ③假说证伪 → ④验证修复”执行。 | 覆盖更新 `SKILL.md` 与脚本文档 |
| `writing-plans` | `skills/writing-plans` | 编写落地计划：面对多步骤复杂任务时，在动工前产出带检查点（Checkpoints）、可验证步骤的结构化执行计划。 | 覆盖更新 `SKILL.md` 与评审提示 |
| `executing-plans` | `skills/executing-plans` | 执行实现计划：分步推进任务，按计划执行并在关键检查点暂停与用户对齐确认。 | 覆盖更新 `SKILL.md` |
| `verification-before-completion` | `skills/verification-before-completion` | **完成前硬性实证验证**：声称修复/完成或 Git 提交前，必须在终端执行测试/验证命令，用控制台真实输出作为证据。 | 覆盖更新 `SKILL.md` |
| `receiving-code-review` | `skills/receiving-code-review` | 理性响应代码审查反馈：严禁无脑附和或机械盲改，要求经过技术推导与实证验证后有理有据地采纳或反驳。 | 覆盖更新 `SKILL.md` |
| `using-git-worktrees` | `skills/using-git-worktrees` | Git Worktree 环境隔离：为新特性或计划执行创建独立的物理目录分支，避免污染主工作树。 | 覆盖更新 `SKILL.md` |

---

### 2. ECC 社区技能体系 (24 个)

上游仓库统一为：[`https://github.com/affaan-m/ECC`](https://github.com/affaan-m/ECC)

| Skill 名称 | 上游路径 | 核心作用与职责 | 更新说明 |
| :--- | :--- | :--- | :--- |
| `golang-patterns` | `skills/golang-patterns` | Go 惯用设计模式：并发安全控制（goroutine/channel/sync）、接口抽象、错误封装、Context 传递与零内存分配。 | 覆盖更新 `SKILL.md` |
| `golang-testing` | `skills/golang-testing` | Go 测试工程规范：表驱动测试、Benchmark 基准测试、Fuzzing 模糊测试与测试覆盖率要求。 | 覆盖更新 `SKILL.md` |
| `api-design` | `skills/api-design` | RESTful API 架构规范：端点资源命名、HTTP 状态码映射、游标/偏移分页、Envelope 错误响应与版本控制。 | 覆盖更新 `SKILL.md` |
| `hexagonal-architecture` | `skills/hexagonal-architecture` | 六边形架构（Ports & Adapters）：领域模型边界划分、依赖反转与用例编排（契合 Cordis 微内核机制）。 | 覆盖更新 `SKILL.md` |
| `frontend-patterns` | `skills/frontend-patterns` | Next.js / React 前端开发模式：组件分层骨架、状态管理、渲染性能与 UI 架构最佳实践。 | 覆盖更新 `SKILL.md` |
| `react-performance` | `skills/react-performance` | **基于 Vercel 70+ 性能法则**：消除请求瀑布、阻止不必要重渲染、优化 Bundle 打包体积与服务端流式传输。 | 覆盖更新 `SKILL.md` |
| `react-patterns` | `skills/react-patterns` | 现代 React 进阶模式：复合组件（Compound Components）、自定义 Hooks 抽象、Suspense 与错误边界。 | 覆盖更新 `SKILL.md` |
| `react-testing` | `skills/react-testing` | React / Next.js 测试套件：React Testing Library、Vitest/Jest 单元测试、MSW 网络 Mock 与 a11y 断言。 | 覆盖更新 `SKILL.md` |
| `nextjs-turbopack` | `skills/nextjs-turbopack` | Next.js Turbopack 编译优化：增量打包配置、文件系统缓存加速与 Turbopack/Webpack 选型指南。 | 覆盖更新 `SKILL.md` |
| `design-system` | `skills/design-system` | 设计系统一致性审计：Design Tokens 规范、语义色彩变量（WCAG 对比度）、组件变体（Variant Props）设计。 | 覆盖更新 `SKILL.md` |
| `accessibility` | `skills/accessibility` | 前端无障碍标准（WCAG 2.2 AA）：键盘焦点导航、颜色对比度、屏幕阅读器语义及 ARIA 属性最佳实践。 | 覆盖更新 `SKILL.md` |
| `motion-ui` | `skills/motion-ui` | Next.js / React UI 动效系统：生产级转场、微交互弹簧动画与减少动态效果（reduced-motion）无障碍适配。 | 覆盖更新 `SKILL.md` |
| `motion-patterns` | `skills/motion-patterns` | 常见动效标准化实现：弹窗（Modal）、抽屉（Drawer）、Toast 消息提示、列表交错（Stagger）与手势动画。 | 覆盖更新 `SKILL.md` |
| `clickhouse-io` | `skills/clickhouse-io` | ClickHouse 分析库工程规范：ReplacingMergeTree 等引擎 DDL 设计、分区/主键优化、高吞吐批量写入与聚合查询。 | 覆盖更新 `SKILL.md` |
| `data-throughput-accelerator` | `skills/data-throughput-accelerator` | 高吞吐量数据管道加速：大数据摄取、数据回填（Backfill）、异步 ETL、Manifest 追赶与批量写入吞吐调优。 | 覆盖更新 `SKILL.md` |
| `content-hash-cache-pattern` | `skills/content-hash-cache-pattern` | 内容哈希缓存模式：基于 SHA-256 算法实现与物理路径解耦的自动失效缓存，用于耗时文件计算。 | 覆盖更新 `SKILL.md` |
| `security-review` | `skills/security-review` | 综合安全审查：敏感用户输入过滤/转义、Secrets 密钥防护、鉴权授权漏洞排查与 OWASP Top 10 防范。 | 覆盖更新 `SKILL.md` 及相关文档 |
| `security-bounty-hunter` | `skills/security-bounty-hunter` | 高危远程漏洞挖掘：排查越权访问、远程代码执行（RCE）、SQL/命令注入漏洞与认证缺陷。 | 覆盖更新 `SKILL.md` |
| `production-audit` | `skills/production-audit` | 生产就绪度审计（Production Readiness）：单点故障排查、探针健康检查、超时/熔断机制与上线前检查。 | 覆盖更新 `SKILL.md` |
| `docker-patterns` | `skills/docker-patterns` | Docker 与 Compose 最佳实践：多阶段构建、精简安全基础镜像、网络卷配置与非 root 权限运行。 | 覆盖更新 `SKILL.md` |
| `deployment-patterns` | `skills/deployment-patterns` | 持续部署工作流：CI/CD Pipeline 模式、金丝雀/灰度发布、健康探针检测与回滚策略设计。 | 覆盖更新 `SKILL.md` |
| `e2e-testing` | `skills/e2e-testing` | Playwright 端到端自动化测试：Page Object Model 设计、网络请求 Mock、CI 运行集成与 Flaky 测试治理。 | 覆盖更新 `SKILL.md` |
| `tdd-workflow` | `skills/tdd-workflow` | 测试驱动开发工作流：强制 Red-Green-Refactor 循环，确保新功能与 Bug 修复达到 80%+ 测试覆盖率。 | 覆盖更新 `SKILL.md` |
| `code-review-skill` | `skills/code-review-skill` | 全栈代码审查指南：跨语言代码质量、坏味道排查、性能与架构审计及建设性评审反馈。 | 覆盖更新 `SKILL.md` 及子目录 |
| `shadcn` | `skills/shadcn` | shadcn/ui 组件库开发管理：组件检索、样式与主题定制、Preset 预设应用与 CLI 命令标准。 | 覆盖更新 `SKILL.md` 及规则目录 |

---

### 3. 独立开源 Agent 生态技能 (2 个)

| Skill 名称 | 官方仓库地址 | 仓库内路径 | 核心作用与职责 | 更新说明 |
| :--- | :--- | :--- | :--- | :--- |
| `write-notes-like-deepseek` | [`https://github.com/czm15053/write-notes-like-deepseek`](https://github.com/czm15053/write-notes-like-deepseek) | 根目录 | 架构决策治理与防撞护栏：要求在开发前沉淀 Agent Notes（为什么做、放弃了什么），解决跨会话失忆与破坏性重构。 | 全量覆盖：包含 `SKILL.md`、`templates/`、`scripts/`、`references/`、`board.html` |
| `autoresearch` | [`https://github.com/dave1010/autoresearch`](https://github.com/dave1010/autoresearch) | 根目录 | 自动化目标驱动循环：围绕 Metric 指标持续执行 modify → verify → keep/discard 的自动化评估与修复闭环。 | 全量覆盖：包含 `SKILL.md`、各子命令 markdown 文档及编排脚本 |

---

### 4. Wavelet 自研/核心业务与框架技能 (10 个)

上游仓库统一为当前开源主仓库：[`https://github.com/Rain-kl/Wavelet`](https://github.com/Rain-kl/Wavelet)  
存放路径：`.agents/skills/<skill-name>`  
> 注：专门针对本项目架构与业务流程的开发教程类技能，统一以 `wv-` 为命名前缀。

| Skill 名称 | 核心作用与职责 | 维护与更新原则 |
| :--- | :--- | :--- |
| `wv-new-api` | Cordis 插件 HTTP API 开发：通过 `ctx.Router()` 声明路由、挂载中间件、Handler/Service 分层与 Swagger 生成。 | 随 Cordis 微内核路由与上下文机制演进同步就地更新。 |
| `wv-new-async-task` | 异步任务与定时调度：通过 `ctx.Task()` 与 `ctx.Schedule()` 注册 Asynq 异步任务与 Cron 调度。 | 随异步 Worker 架构和调度器驱动调整同步更新。 |
| `wv-new-setting` | 配置与系统设置声明：通过 `ctx.Settings()` 声明热加载设置 Schema，绑定 YAML 启动配置。 | 随配置系统与管理台热更新模型演进同步更新。 |
| `wv-database-guide` | 数据库开发与迁移规范：插件自包含 Goose SQL 迁移（PG/SQLite 双方言）、ClickHouse 分析表与 batchwriter 接入。 | 随着双数据库方言策略或 ClickHouse 批量写入框架变更而维护。 |
| `wv-cache-framework` | 三层缓存访问规范：基于 `ctx.Cache()` 与 `contracts.CacheService`（RAM L1 + Redis L2 + Pub/Sub 同步）。 | 随着缓存失效广播与多级缓存抽象变更而维护。 |
| `wv-logstore` | 日志与分析库架构：日志/分析用途表划分、`plugins/domain/risk_control/logstore` 接入与 PG/SQLite 回落。 | 随着审计时序日志与主分析库路由策略变动维护。 |
| `wv-file-upload` | 统一文件摄取服务：通过 `upload.Ingest` / `contracts.StorageService` 进行文件存储，严禁旁路写表。 | 随着存储服务抽象和文件接入端点调整同步更新。 |
| `wv-push-notification` | 系统通知推送机制：通过统一触发器投递消息，支持跨域解耦与动态推送配置。 | 随着通知服务与网关契约改动同步更新。 |
| `wv-logging` | 结构化日志与链路追踪：基于 `backend/pkg/logger`（Zap + otelzap + 5000 行环形缓冲区，支持 Admin WebSocket 日志流）及 `contracts.LoggerService`。 | 随着后端日志系统与 Trace 上下文调整维护。 |
| `wv-release-guide` | 版本发布规范：从 Git 历史整理规范的 Version Bump 提交信息以自动触发双语 GitHub Release。 | 随着 CI/CD 自动化发版脚本逻辑调整维护。 |

---

## 三、非 Git 方式的 Skills 更新操作方法

由于当前项目未采用 Git Submodule，而是将 Skill 文件直接纳入主仓库版本管理，更新时**采用拉取最新上游源码并按需覆盖拷贝**的流程。

### 核心更新原则
1. **纯净覆盖**：更新时先将对应 Skill 目录清空或直接覆盖，避免上游废弃的文件遗留在本地。
2. **禁止带入 `.git`**：严禁将外部仓库的 `.git` 目录拷贝进项目，否则会导致嵌套子仓库异常。
3. **原子提交**：更新技能后，运行架构与代码检查，然后使用标准 Conventional Commits 进行本地提交（禁止直接推送）。

---

### 方法 A：使用内置一键更新脚本（自动化推荐）

本项目在 `scripts/update_skills.sh` 提供了自动化更新脚本。

#### 常用命令
```bash
# 查看帮助与说明
bash scripts/update_skills.sh --help

# 查看所有技能及其对应仓库
bash scripts/update_skills.sh --list

# 更新单个 Superpowers 技能（例：systematic-debugging）
bash scripts/update_skills.sh systematic-debugging

# 一键更新全部 8 个 Superpowers 技能
bash scripts/update_skills.sh --all-sp

# 更新单个 ECC 技能（例：golang-patterns）
bash scripts/update_skills.sh golang-patterns

# 一键更新全部 25 个 ECC 技能
bash scripts/update_skills.sh --all-ecc

# 一键更新所有外部依赖技能 (Superpowers + ECC + DeepSeek + Autoresearch)
bash scripts/update_skills.sh --all
```

---

### 方法 B：单项手动更新命令（Shell 快速执行）

#### 场景 1：更新 Superpowers 仓库中的技能
```bash
git clone --depth 1 https://github.com/obra/superpowers /tmp/sp-temp

# 拷贝指定技能
cp -R /tmp/sp-temp/skills/systematic-debugging .agents/skills/
cp -R /tmp/sp-temp/skills/verification-before-completion .agents/skills/

rm -rf /tmp/sp-temp
```

#### 场景 2：更新 ECC 仓库中的技能
```bash
git clone --depth 1 https://github.com/affaan-m/ECC /tmp/ECC-temp

cp -R /tmp/ECC-temp/skills/golang-patterns .agents/skills/
cp -R /tmp/ECC-temp/skills/react-performance .agents/skills/

rm -rf /tmp/ECC-temp
```

---

## 四、更新后的验证与提交标准

完成任何 Skill 的拷贝更新后，必须执行以下流程：

1. **检查 Git 变动**：
   ```bash
   git diff --stat .agents/skills/
   ```
2. **运行项目架构合规检查**：
   ```bash
   scripts/check_cordis_architecture.sh
   ```
3. **本地 Git 规范提交**（严格遵守 Conventional Commits，严禁 push 远程）：
   ```bash
   git add .agents/skills/
   git commit -m "chore(skills): update <skill-name> from upstream"
   ```
