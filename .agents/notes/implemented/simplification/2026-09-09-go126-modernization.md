# Agent Note: 深度应用 Go 1.26 现代语言特性与语法糖重构

Status: implemented

## Problem

Wavelet 后端模块一直停留在较低版本的语法习惯中，导致多处存在样板代码与低效模式：
1. **错误类型解包繁琐**：多处采用传统的两步式解包（`var target *T; if errors.As(err, &target)`），不仅声明冗长，而且变量容易污染外层作用域；
2. **并发与等待组样板多**：goroutine 协同逻辑需要重复书写 `wg.Add(1)`、`go func() { defer wg.Done(); ... }`；
3. **Map 集合键值提取低效繁琐**：多处手动声明切片并通过 `for k := range m` 进行 `append`，产生多行样板代码；
4. **内存重置产生无谓 GC 压力**：高频接口与内部缓存（如 `Container.interfaceCache`、`disk.Cache.items`）清空时直接使用 `make(map...)` 重新分配，废弃既有哈希桶，增加堆内存开销；
5. **固定循环书写陈旧**：广泛使用 C 风格的三段式 `for i := 0; i < n; i++`，未利用现代 Go 的 `for range n`。

## Decision

将 `backend/go.mod` 基线升级为 `go 1.26.0`，并在后端核心架构与领域层充分释放现代 Go（Go 1.22 ~ 1.26）的语言特性与语法糖：
1. **全面采纳 Go 1.26 泛型单步解包 `errors.AsType[T](err)`**：
   - 在统一响应中间件 `pkg/response/middleware.go`、配置加载 `plugins/infra/config/source.go`、管理台异常 `admin/errs`、文件服务 `upload/filesrv` 及所有配套测试中，全部将 `errors.As` 重构为 `errors.AsType[*T](err)`；
2. **利用 Go 1.25/1.26 的 `sync.WaitGroup.Go(func())` 并发糖**：
   - 彻底改造 `backend/core/events.go` 事件总线并行广播机制，直接调用 `wg.Go(func() { ... })` 驱动监听器并发执行，消除手动计数与闭包传递；
   - 现代化改造 `core/context_test.go`、`core/events_test.go`、`pkg/limiter/memory_test.go` 等高并发测试套件；
3. **标准库集合流式糖 `slices.Collect(maps.Keys(m))`**：
   - 在 `plugins/drivers/driver_asynq_worker/meta.go` 与 `pkg/testhelper/test_helper.go` 中，以一行声明式替代以往 5 行的切片声明与遍历追加；
4. **内置 `clear(m)` 消除哈希表重分配**：
   - 在 `backend/core/container.go`（服务容器接口解析缓存失效）与 `backend/pkg/cache/disk/cache.go`（磁盘缓存清空）中，将重新 `make` 升级为 `clear(m)`，保留底层容量，彻底规避缓存高频抖动下的 GC 开销；
5. **基础函数式语法糖扩展**：
   - 在底层 `backend/pkg/util/slice.go` 引入高频通用的泛型函数式语法糖 `Map[T, U any]` 与 `Filter[T any]` 并补齐单元测试；
6. **整数迭代语法糖 `for range n` 与 `for i := range n`**：
   - 梳理并重写全流程并发测试中的传统计步循环。

## Alternatives considered

- **方案 A：仅升级 `go.mod` 声明版本，不改动现有代码**：放弃。仅更新版本号无法为业务开发与代码可读性带来实质收益，也未解决重复模式的维护负担。
- **方案 B：使用第三方函数式库（如 `samber/lo`）**：放弃。Wavelet 秉持保持底层基础设施与基础库纯洁轻量的架构原则，直接利用 Go 1.26 原生标准库（`slices`, `maps`, `errors`, `sync`）与轻量受管的 `pkg/util` 泛型实现，杜绝引入重型第三方依赖。

## Consequences

- **收益**：
  - 代码行数显著压缩，语义直观易懂，完全消除了错误断言与并发计数的模板样板；
  - 减少了锁内重新分配 map 带来的堆分配与 GC 暂停时间；
  - 为整个 Wavelet 后端代码库树立了 Go 1.26 时代的现代最佳实践基线；
  - 100% 保持既有 Cordis 架构防线纯净度，全量测试 `go test ./...` 与 `make code-check` 0 error / 0 violation 通过。
- **代价**：
  - 开发和构建环境需保证 Go 编译器版本 >= 1.26。
