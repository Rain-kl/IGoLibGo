---
name: "wv-new-api"
description: "Wavelet 项目专用：当新增或修改业务 API、Handler、服务层逻辑、插件路由注册时必须使用。本技能指导基于 Cordis 插件的 API 架构、ctx.Router() 声明式路由注册、Handler/Service 分层、RESTful api-design 响应规范、Swagger 与质量门禁。"
metadata:
  origin: Wavelet
---

# 新增业务 API 开发与路由注册规范 (Cordis 插件化架构)

本技能是 Wavelet 在 Cordis 微内核与插件化架构下，进行 HTTP API 接口开发、路由注册以及遵循统一 RESTful `api-design` 规范的唯一指导。

---

## 1. 核心架构哲学：插件自包含 (Self-Contained Plugins)

在 Cordis 架构中，所有业务能力均封装为**高内聚、自包含的插件 (Plugin)**。每个插件自主管理自身的路由声明、中间件挂载、服务逻辑、数据模型与迁移脚本。

### 插件目录标准结构 (`backend/downstream/plugins/<name>/` 或 `backend/plugins/domain/<name>/`)

所有标准插件与下游定制插件，**统一以 [`backend/downstream/plugins/custom_example`](file:///backend/downstream/plugins/custom_example) 为基准模板**，严格采用物理子包隔离的分层架构：

```text
backend/downstream/plugins/custom_example/ (或 backend/plugins/domain/order/)
├── plugin.go           # 插件根入口：实现 core.Plugin，装配各子包并向 Cordis 注册
│
├── consts/             # package consts：常量、配置键名与业务常量定义
│   └── consts.go
│
├── controller/         # package controller：HTTP 控制器与路由声明 (参数绑定、会话获取、RESTful 响应)
│   └── hello/          # 业务分组/实体子包
│       └── hello.go    # 接口处理 Handler（直接以业务命名，禁止平铺 handlers_*.go）
│
├── service/            # package service：业务逻辑层（用例编排、事务控制、事件发布，严禁依赖 *gin.Context）
│   └── order.go        # 订单业务用例实现
│
├── dao/                # package dao：数据访问持久化层 DAL (GORM CRUD、SQL 转义防注入)
│   └── order.go        # 订单数据访问实现
│
├── model/              # package model：纯数据实体与 DTO（无外部依赖）
│   ├── entity/         # 数据库映射实体 (TableName() 带插件专属前缀)
│   │   └── order.go
│   └── dto/            # 请求 Request DTO 与响应 Response DTO、领域对象
│       └── order.go
│
└── migrations/         # 专属嵌入式 Goose SQL 双方言迁移脚本 (//go:embed)
    ├── postgres/       # PostgreSQL 迁移脚本
    └── sqlite/         # SQLite 迁移脚本
```
> ⚠️ **严禁**：严禁在插件根目录平铺 `handlers_*.go`、`service_*.go`、`dao_*.go` 等文件，必须使用物理子包，且子包内直接按业务实体命名。严格约束 `controller -> service -> dao -> model` 单向依赖。

---

## 2. 插件契约与路由注册流程

### 步骤 1：定义插件结构并实现 `core.Plugin`

```go
package order

import (
	"Wavelet/core"
	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/order/controller"
	"Wavelet/plugins/domain/order/service"
)

type Plugin struct {
	svc *service.OrderService
}

func (p *Plugin) Name() string {
	return "domain.order"
}

func (p *Plugin) Apply(ctx *core.Context) error {
	// 1. 初始化业务 Service
	p.svc = service.NewOrderService(ctx)

	// 2. 如果需要对外暴露服务，注入 IoC 容器供其他插件消费
	// core.Provide[contracts.OrderService](ctx, p.svc)

	// 3. 注册 HTTP 路由与中间件
	p.registerRoutes(ctx)

	return nil
}
```

### 步骤 2：通过 `ctx.Router()` 挂载路由组与中间件

通过微内核扩展点 `ctx.Router()` 声明式挂载语义化路由：

```go
func (p *Plugin) registerRoutes(ctx *core.Context) {
	authSvc, _ := core.Inject[contracts.AuthService](ctx)
	
	group := ctx.Router().Group("/api/v1/orders")
	if authSvc != nil {
		group.Use(authSvc.RequireAuthMiddleware())
	}

	ctrl := controller.NewOrderController(p.svc)
	group.GET("", ctrl.ListOrders)
	group.POST("", ctrl.CreateOrder)
	group.GET("/:id", ctrl.GetOrderDetail)
	group.PUT("/:id/cancel", ctrl.CancelOrder)
	group.DELETE("/:id", ctrl.DeleteOrder)
}
```

### 步骤 3：公开接口与白名单注册 (`RegisterWhitelist`)

如果插件包含**无需登录**的公开端点（如登录、注册、人机校验、Webhooks、公开状态查询），必须在 `Apply` 中主动注册到白名单：

```go
func (p *Plugin) Apply(ctx *core.Context) error {
	ctx.Router().RegisterWhitelist(
		"/api/v1/public/ping",
		"/api/v1/public/webhook/*",
	)
	...
}
```
> 💡 **防线机制**：注册到白名单的路由在经过 `RequireAuthMiddleware()` 时将自动放行，彻底消除全局/组级鉴权中间件引起的 401 Unauthorized 误拦截。

---

## 3. HTTP 响应与状态码标准 (`api-design` 规范)

项目全面采纳通用技能 **`api-design`** 的 RESTful 规范，统一通过 `Wavelet/pkg/response` 输出：

### 3.1 状态码与响应语义

| HTTP 方法 | 场景 | HTTP 状态码 | 响应方式 |
| :--- | :--- | :--- | :--- |
| **GET** | 查询单条或列表 | `200 OK` | `c.JSON(http.StatusOK, response.OK(data))` |
| **GET** (分页) | 分页列表查询 | `200 OK` | `c.JSON(http.StatusOK, response.Paged(items, meta))` |
| **POST** | 新建资源成功 | `201 Created` | `response.Created(c, "/api/v1/orders/"+id, order)` (附带 Location 头) |
| **PUT / PATCH** | 更新资源成功返回内容 | `200 OK` | `c.JSON(http.StatusOK, response.OK(order))` |
| **DELETE** | 删除成功无返回体 | `204 No Content` | `response.NoContent(c)` |
| **任意** | 客户端参数验证失败 | `400 Bad Request` | `response.AbortBadRequestWithCode(c, "validation_error", "参数校验失败")` |
| **任意** | 未登录或认证过期 | `401 Unauthorized` | `response.AbortUnauthorized(c, "未登录")` |
| **任意** | 资源未找到 | `404 Not Found` | `response.AbortNotFound(c, "订单不存在")` |
| **任意** | 服务端异常 | `500 Internal Error` | `response.AbortInternal(c, "内部系统错误")` |

### 3.2 统一响应信封格式

* **成功数据信封**：
  ```json
  {
    "data": { "id": "123", "status": "pending" }
  }
  ```
* **分页数据信封**：
  ```json
  {
    "data": [{ "id": "123" }],
    "meta": { "total": 100, "page": 1, "per_page": 20, "total_pages": 5 }
  }
  ```
* **错误信封**：
  ```json
  {
    "error": {
      "code": "validation_error",
      "message": "请求参数校验失败",
      "details": [{ "field": "email", "issue": "邮箱格式不正确" }]
    }
  }
  ```

---

## 4. Handler 与 Service 职责划分

### Handler 规范 (`controller/<entity>.go`)
1. **参数绑定与校验**：使用 `c.ShouldBindJSON` 或 `c.ShouldBindQuery`。
2. **提取身份**：从 `c` 提取当前用户。
3. **调用 Service**：将 `c.Request.Context()` 传递给 Service，获取结果。
4. **错误处理**：统一使用 `response.Abort*` 系列函数中断请求，禁止裸写 `c.JSON(status, gin.H{"error": ...})`。
5. **Swagger 注释**：使用规范的 `@Success` 与 `@Failure` 注释。

```go
// CreateOrder 创建订单
// @Summary 创建订单
// @Tags Order
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "创建订单参数"
// @Success 201 {object} response.Any{data=dto.OrderDTO} "创建成功"
// @Failure 400 {object} response.AnyError "参数错误"
// @Router /api/v1/orders [post]
func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortBadRequestWithCode(c, "validation_error", "参数绑定失败")
		return
	}

	order, err := ctrl.svc.CreateOrder(c.Request.Context(), req)
	if err != nil {
		response.AbortInternal(c, "创建订单失败")
		return
	}

	response.Created(c, "/api/v1/orders/"+order.ID, order)
}
```

### Service 规范 (`service/<entity>.go`)
1. 纯 Go 逻辑，第一参数为 `ctx context.Context`，返回 `(result, error)`。
2. **严禁依赖 `*gin.Context`** 或调用 `c.JSON`/`Abort*`。
3. 数据库操作通过 `ctx.DB()` 或受 Trace 保护的 DB 实例完成。
4. 日志记录统一调用 `logger.ErrorF(ctx, format, args...)`。

---

## 5. 质量验证门禁

在完成 API 开发后，必须依次运行以下命令：
```bash
make swagger        # 重新生成 Swagger 文档
make format         # Biome + golangci-lint 自动格式化
make code-check     # 静态代码质量检查 (架构规则 + golangci-lint + 前端类型检查)
cd backend && go test ./...  # 运行单元测试
```
