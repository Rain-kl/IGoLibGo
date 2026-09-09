# Agent Note: 升级 Go 1.27 基线与跨版本现代语法糖全量重构

Status: implemented

## Problem

随着 Go 1.27 的正式发布，Wavelet 面临着进一步消除历史包袱、简化框架使用体验的架构升级契机：
1. **微内核方法级泛型受限（Go < 1.27 历史硬伤）**：在 Go 1.26 及之前，方法不支持类型参数（Generic Methods），导致 Cordis 框架只能将依赖注入定义为顶层包函数 `core.Provide[T](ctx, svc)` 与 `core.Inject[T](ctx)`。这与 Cordis “以 Context 为中心总线”的流畅风格（如 `ctx.Router()`、`ctx.DB()`、`ctx.Cache()`）格格不入，且强迫所有业务插件无谓导入 `Wavelet/core`；
2. **外部 UUID 依赖冗余**：Session 管理与 OAuth 状态生成长期依赖第三方 `github.com/google/uuid`，未能利用 Go 1.27 原生内置的 RFC 9562 标准库 `"uuid"`；
3. **历史现代语法糖未彻底普及**：全库部分测试与业务模块仍残留传统 C 风格循环 `for i := 0; i < n; i++`、等待组样板代码 `wg.Add(1)` + `go func() { defer wg.Done() }()` 以及反射切片排序 `sort.Slice`。

## Decision

将全仓库 Go 基线升级至 **Go 1.27.0**，并在全量模块（微内核、通用基础库、全部 7 个业务领域插件、驱动与基础设施层、下游模板）中全面落地现代 Go 语法糖：

1. **全面采纳 Go 1.27 结构体泛型方法（Generic Methods）**：
   - 在 `*core.Container` 与 `*core.Context` 上原生实现泛型方法：`ctx.Provide[T](svc)`、`ctx.Inject[T]()`、`ctx.MustInject[T]()`、`ctx.Has[T]()`、`ctx.When[T](fn)`、`ctx.Bind[T](fn)`、`ctx.Using[T](fn)`；
   - 原包级顶层函数保留为 100% 透明转发别名，无破坏性兼容存量调用；
   - 将全库（`auth`、`user`、`upload`、`admin`、`system`、`risk_control`、`msg_gateway`、`infra`、`drivers` 及 `downstream/custom_example`）的依赖注册与获取全面重构为原生流畅写法；
2. **全面迁移至 Go 1.27 标准库 `"uuid"` 并修剪外部依赖**：
   - 在 `pkg/util/uuid.go`、`plugins/domain/auth/service/session_service.go` 与 `plugins/domain/auth/controller/oauth.go` 中彻底替换 `github.com/google/uuid` 为标准库 `"uuid"`（`uuid.New().String()`）；
   - 执行 `go mod tidy` 将 `github.com/google/uuid` 从直接依赖清单中清除；
3. **普及 Go 1.25 `sync.WaitGroup.Go` 与标准库并发糖**：
   - 在 `driver_asynq_worker/meta_test.go` 与 `infra/database/sqlite_concurrency_test.go` 等高并发场景下，直接使用 `wg.Go(fn)` 驱动并发任务；
4. **流式集合与切片反向迭代（Go 1.23 / 1.24）**：
   - 在 `pkg/logger/ringbuffer.go` 引入 `slices.Backward` 反向迭代器；
   - 在 `msg_gateway` 引入 `slices.Clone` 与 `slices.Contains`；
   - 在 `risk_control` 淘汰 `sort.Slice` 反射排序，切换至高内聚的 `slices.SortFunc` + `cmp.Compare`；
5. **整型范围迭代 `for range n`（Go 1.22）**：
   - 梳理全库计数循环（如 `snowflake.go`、`user/service.go`、`meta_test.go` 等），全面改写为 `for range n` / `for i := range n`；
6. **多子代理协同分治机制**：
   - 编制模块规范白皮书 [docs/go127-refactor-plan.md](file:///Users/ryan/Code/Go/Wavelet/docs/go127-refactor-plan.md)，派发 9 个专职子代理对号认领模块推进，确保每个插件职责清晰、物理隔离。

## Alternatives considered

- **方案 A：仅升级 `backend/go.mod` 声明版本至 1.27，暂不重构 `ctx.Provide` / `ctx.Inject`**：放弃。Go 1.27 最核心的语言进化正是结构体泛型方法，Cordis 框架此前被迫使用顶层函数的历史包袱若不趁机拔除，会持续增加业务插件的导入冗余与心智负担。
- **方案 B：直接删除包级 `core.Provide` / `core.Inject` 函数，强制仅保留 `ctx.Provide`**：放弃。激进删除会导致外部插件与第三方下游扩展编译中断。通过双轨支持 + 内部透明转发，既实现了内部代码现代化，又保证了 100% 向下兼容。
- **方案 C：单线程逐个文件手动修改**：放弃。全库涉及 9 个独立模块、60+ 文件，采用子代理按架构分层并发推进，大幅提升了交付效率并保证了每个模块单元测试的就地闭环。

## Consequences

- **收益**：
  - 代码行数显著精简，彻底消除了倒装的依赖注入函数调用，API 风格与 Cordis 微内核高度自洽；
  - 移除了外部 UUID 库的直接依赖，工程更纯净、编译更轻量；
  - 运行时广泛受益于 Go 1.27 的小对象分配优化以及 `slices.SortFunc` / `clear` 的底层性能红利；
  - 全量架构防线检查 `scripts/check_cordis_architecture.sh` 0 违规，全量单测 `go test ./...` 100% 通过，`golangci-lint run` 0 issue。
- **代价**：
  - 研发与 CI/CD 构建环境需将 Go 编译器基线提升至 `>= 1.27.0`。
