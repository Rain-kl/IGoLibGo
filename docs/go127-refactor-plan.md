# Wavelet Go 1.27 现代化与语法糖重构执行计划指南

本文档为 Wavelet 项目升级至 **Go 1.27** 并全面重构现代语法糖（涵盖 Go 1.22 ~ 1.27）的标准指南与任务分配白皮书。
所有被派发的子代理（Subagent）必须首先阅读本文档，并依照本文档中定义的重构标准领取并执行各自所属模块的任务。

---

## 1. 重构目标与架构防线

1. **基线升级**：`backend/go.mod` 基线升级为 `go 1.27.0`；
2. **微内核能力飞跃**：全面利用 Go 1.27 突破性的**结构体泛型方法（Generic Methods）**，在 `*core.Context` 与 `*core.Container` 上原生支持 `ctx.Provide[T](svc)`、`ctx.Inject[T]()`、`ctx.When[T](...)`、`ctx.Using[T](...)`；
3. **依赖彻底纯净化**：全面以 Go 1.27 原生内置的 RFC 9562 标准库 `"uuid"` 替代 `github.com/google/uuid`，并在业务迁移完成后通过 `go mod tidy` 从工程依赖树中彻底剪除外部 uuid 依赖；
4. **历史语法糖全量普及**：全库淘汰历史遗留陈旧写法，补齐 Go 1.22 ~ 1.26 引入的高性能、低模板语法糖；
5. **绝对遵守 Cordis 架构守则**：
   - 保持微内核 `backend/core/` 纯净，严禁业务侵入；
   - 保持 `backend/pkg/` 通用库纯净，严禁反向依赖 `core/` 或 `plugins/`；
   - 插件间严禁跨包私有调用，跨插件通信仅面向 `backend/core/contracts` 编程；
   - 单表单一所有者原则：严禁跨插件 DDL。

---

## 2. 现代 Go 语法糖与重构规范清单 (Go 1.22 ~ 1.27)

子代理在重构所负责模块时，必须逐条对照并应用以下语法糖规范：

### 规范 1：结构体泛型方法 (Go 1.27 Generic Methods)
- **旧模式**：必须调用包级顶层函数：
  ```go
  core.Provide[contracts.AuthService](ctx, svc)
  svc, err := core.Inject[contracts.UserService](ctx)
  core.When[contracts.DBService](ctx, func(db contracts.DBService) { ... })
  ```
- **新模式**：直接调用上下文/容器的方法：
  ```go
  ctx.Provide[contracts.AuthService](svc)
  svc, err := ctx.Inject[contracts.UserService]()
  svc := ctx.MustInject[contracts.UserService]()
  ctx.When[contracts.DBService](func(db contracts.DBService) { ... })
  ctx.Using[contracts.CacheService](func(cache contracts.CacheService) { ... })
  ```

### 规范 2：标准库 RFC 9562 `uuid` (Go 1.27 `"uuid"`)
- **旧模式**：
  ```go
  import "github.com/google/uuid"
  token := uuid.NewString()
  ```
- **新模式**：
  ```go
  import "uuid"
  token := uuid.New().String()
  // 或 uuid.NewV7().String()、uuid.Parse(str)
  ```

### 规范 3：单步泛型错误解包 (Go 1.26 `errors.AsType[T]`)
- **旧模式**：
  ```go
  var target *CustomError
  if errors.As(err, &target) { ... }
  ```
- **新模式**：
  ```go
  if target, ok := errors.AsType[*CustomError](err); ok { ... }
  ```

### 规范 4：并发协同等待组 (Go 1.25 `sync.WaitGroup.Go`)
- **旧模式**：
  ```go
  wg.Add(1)
  go func() {
      defer wg.Done()
      doWork()
  }()
  ```
- **新模式**：
  ```go
  wg.Go(func() {
      doWork()
  })
  ```

### 规范 5：流式集合提取 (Go 1.24 `slices.Collect`)
- **旧模式**：
  ```go
  keys := make([]string, 0, len(m))
  for k := range m {
      keys = append(keys, k)
  }
  ```
- **新模式**：
  ```go
  keys := slices.Collect(maps.Keys(m))
  vals := slices.Collect(maps.Values(m))
  ```

### 规范 6：基准测试迭代器 (Go 1.24 `testing.B.Loop()`)
- **旧模式**：`for i := 0; i < b.N; i++`
- **新模式**：`for b.Loop() { ... }`

### 规范 7：反向切片迭代 (Go 1.23 `slices.Backward`)
- **旧模式**：`for i := len(s) - 1; i >= 0; i-- { ... }`
- **新模式**：`for i, v := range slices.Backward(s) { ... }`

### 规范 8：整型计数循环 (Go 1.22 `for range n`)
- **旧模式**：`for i := 0; i < n; i++` 或 `for j := 0; j < iterations; j++`
- **新模式**：`for range n` 或 `for i := range n`

### 规范 9：内存原地重置 (Go 1.21 `clear(m)` / `clear(s)`)
- **旧模式**：`m = make(map[K]V)` 废弃旧 map 造成 GC 堆抖动
- **新模式**：`clear(m)` 原地清空键值对，复用已分配的哈希桶容量

---

## 3. 子代理模块任务分配大盘 (Task Board)

每个专职子代理对号入座，认领并执行以下模块任务：

### 任务 1：微内核核心（Core）
- **领任务角色**：`Subagent Core`
- **目标路径**：`backend/core/`
- **交付内容**：
  1. 在 `Container` 与 `Context` 上实现原生泛型方法（`Provide`、`Inject`、`MustInject`、`Has`、`When`、`Using` 等）；
  2. 保留原顶层函数 `core.Provide` / `core.Inject` 作为透明转发；
  3. 运行 `go test ./core/...` 确保测试全部绿灯。

### 任务 2：通用基础库（Pkg）
- **领任务角色**：`Subagent Pkg`
- **目标路径**：`backend/pkg/`
- **交付内容**：
  1. 保证 `backend/pkg/util/uuid.go` 采用 Go 1.27 标准库 `"uuid"`；
  2. 梳理测试与工具包，将 `goroutine_test.go`、`batchwriter/writer_test.go` 等处的 `wg.Add(1)` + `go` 重构成 `wg.Go(...)`；
  3. 运行 `go test ./pkg/...` 确保全部通过。

### 任务 3：认证域插件（Domain Auth）
- **领任务角色**：`Subagent Domain Auth`
- **目标路径**：`backend/plugins/domain/auth/`
- **交付内容**：
  1. `service/session_service.go`：替换 `github.com/google/uuid` 为标准库 `"uuid"`；
  2. `controller/oauth.go`：替换 `github.com/google/uuid` 为标准库 `"uuid"`；
  3. `plugin.go`：重构为 `ctx.Provide[contracts.AuthService](svc)`；
  4. 运行 `go test ./plugins/domain/auth/...`。

### 任务 4：用户域插件（Domain User）
- **领任务角色**：`Subagent Domain User`
- **目标路径**：`backend/plugins/domain/user/`
- **交付内容**：
  1. `plugin.go`：重构为 `ctx.Provide[contracts.UserService](svc)`；
  2. `handlers.go`、`task.go` 等：将 `core.Inject` 升级为 `ctx.Inject`；
  3. 检查并清理遗留 `for i := 0` 循环；
  4. 运行 `go test ./plugins/domain/user/...`。

### 任务 5：上传与存储域插件（Domain Upload）
- **领任务角色**：`Subagent Domain Upload`
- **目标路径**：`backend/plugins/domain/upload/`
- **交付内容**：
  1. `plugin.go`：重构为 `ctx.Provide[contracts.UploadService](svc)`；
  2. `shared/context_services.go`、`routers.go`：将 `core.Inject` 升级为 `ctx.Inject`；
  3. 运行 `go test ./plugins/domain/upload/...`。

### 任务 6：管理后台域插件（Domain Admin）
- **领任务角色**：`Subagent Domain Admin`
- **目标路径**：`backend/plugins/domain/admin/`
- **交付内容**：
  1. `plugin.go`：重构为 `ctx.Provide`；
  2. `service/`、`repository/`、`handler/`：将 `core.Inject` 升级为 `ctx.Inject`；
  3. 运行 `go test ./plugins/domain/admin/...`。

### 任务 7：系统域插件（Domain System）
- **领任务角色**：`Subagent Domain System`
- **目标路径**：`backend/plugins/domain/system/`
- **交付内容**：
  1. `plugin.go` 及相关测试：升级为 `ctx.Provide` / `ctx.Inject`；
  2. 运行 `go test ./plugins/domain/system/...`。

### 任务 8：风控域插件（Domain RiskControl）
- **领任务角色**：`Subagent Domain RiskControl`
- **目标路径**：`backend/plugins/domain/risk_control/`
- **交付内容**：
  1. `plugin.go`、`logstore/db_helper.go`：升级为 `ctx.Provide` / `ctx.Inject`；
  2. 运行 `go test ./plugins/domain/risk_control/...`。

### 任务 9：消息网关域插件（Domain MsgGateway）
- **领任务角色**：`Subagent Domain MsgGateway`
- **目标路径**：`backend/plugins/domain/msg_gateway/`
- **交付内容**：
  1. `plugin.go`、`service/service.go`、`dao/dao.go`：升级为 `ctx.Provide` / `ctx.Inject`；
  2. 运行 `go test ./plugins/domain/msg_gateway/...`。

### 任务 10：基础设施、驱动与下游模板（Infra, Drivers & Downstream）
- **领任务角色**：`Subagent Infra & Drivers`
- **目标路径**：`backend/plugins/infra/`、`backend/plugins/drivers/`、`backend/downstream/`
- **交付内容**：
  1. `plugins/infra/`（storage、cache、database、logger）：升级为 `ctx.Provide`；
  2. `plugins/drivers/`（driver_http、driver_asynq_worker、driver_inproc_worker）：升级为 `ctx.Inject`；
  3. `driver_asynq_worker/meta_test.go`：将 `for i := 0; i < workers; i++` 与 `for j := 0` 改为 `for range`，将 `wg.Add(1)` + `go` 改为 `wg.Go`；
  4. `infra/database/sqlite_concurrency_test.go`：将 `wg.Add(1)` + `go` 改为 `wg.Go`；
  5. `downstream/plugins/custom_example/plugin.go`：示范现代泛型方法；
  6. 运行 `go test ./plugins/infra/... ./plugins/drivers/...`。

---

## 4. 验证与门禁命令

各子代理在各自工作目录下必须使用以下环境变量执行测试（确保使用 Go 1.27.1 编译器）：

```bash
unset GOROOT && export PATH="/Users/ryan/.local/share/mise/installs/go/1.27.1/bin:$PATH" && go test ./...
```
