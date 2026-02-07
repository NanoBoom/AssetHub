# API 设计规则

> 基于代码库分析自动生成 - AssetHub Go + Gin 项目

## 核心模式

### 1. Handler 结构定义

**参考**: `internal/handlers/health.go:14-28`

**规则**:
- Handler 结构体命名：`*Handler` 后缀
- 依赖注入：通过结构体字段（私有）
- 工厂函数：`New*Handler(deps...) *Handler`
- 方法签名：`func (h *Handler) MethodName(c *gin.Context)`

**示例**:
```go
type HealthHandler struct {
    db    *gorm.DB
    redis *cache.RedisClient
}

func NewHealthHandler(db *gorm.DB, redis *cache.RedisClient) *HealthHandler {
    return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
    c.JSON(http.StatusOK, response)
}
```

### 2. 响应结构体定义（面向 OpenAPI）

**参考**: `internal/handlers/health.go:19-24`

**规则**:
- 结构体命名：`*Response` 或 `*Request` 后缀
- 定义位置：与 handler 在同一文件
- 必需 tag：
  - `json:"field_name"` - JSON 字段名（snake_case）
  - `example:"value"` - Swagger 示例值（**关键**）
- 字段可见性：大写（导出）

**反例**（会导致 `additionalProp` 占位符）:
```go
// ❌ 错误：使用 map[string]string
// @Success 200 {object} map[string]string
```

**正例**:
```go
// ✅ 正确：定义专用结构体
type HealthResponse struct {
    Status   string `json:"status" example:"ok"`                // 整体状态
    Database string `json:"database" example:"ok"`              // 数据库状态
    Redis    string `json:"redis" example:"ok"`                 // Redis状态
}

// @Success 200 {object} HealthResponse
```

### 3. Swagger 注释规范

**参考**: `internal/handlers/health.go:30-37`

**规则**:
- 第一行：`// MethodName godoc`（固定格式）
- 必需标签：
  - `@Summary` - 简短描述（一行）
  - `@Description` - 详细描述
  - `@Tags` - API 分组
  - `@Accept` - 接受的 Content-Type（通常 `json`）
  - `@Produce` - 返回的 Content-Type（通常 `json`）
  - `@Success` - 成功响应：`状态码 {object} 结构体名`
  - `@Router` - 路由：`/path [method]`
- 可选标签：
  - `@Param` - 参数：`name location type required "description"`
  - `@Failure` - 失败响应

**模板**:
```go
// CreateAsset godoc
// @Summary      创建资产
// @Description  上传并创建新的资产记录
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        body body CreateAssetRequest true "资产信息"
// @Success      200 {object} response.Response{data=AssetResponse}
// @Failure      400 {object} response.Response
// @Router       /assets [post]
func (h *AssetHandler) CreateAsset(c *gin.Context) { ... }
```

### 4. 路由注册模式

**参考**: `cmd/api/main.go:112-113`

**规则**:
- 位置：`setupRouter()` 函数中
- 步骤：
  1. 使用工厂函数创建 handler：`handler := handlers.NewXxxHandler(deps...)`
  2. 注册路由：`router.METHOD("/path", handler.Method)`
- 中间件顺序（已固定）：Recovery → Logger → ErrorHandler → CORS

**示例**:
```go
func setupRouter(db *gorm.DB, redis *cache.RedisClient, logger *zap.Logger) *gin.Engine {
    router := gin.New()

    // 中间件（固定顺序）
    router.Use(middleware.Recovery(logger))
    router.Use(middleware.Logger(logger))
    router.Use(middleware.ErrorHandler())
    router.Use(middleware.CORS())

    // 路由注册
    healthHandler := handlers.NewHealthHandler(db, redis)
    router.GET("/health", healthHandler.Check)

    assetHandler := handlers.NewAssetHandler(db)
    router.POST("/assets", assetHandler.Create)
    router.GET("/assets/:id", assetHandler.Get)

    return router
}
```

## 通用响应封装

**参考**: `pkg/response/response.go:9-29`

**可选使用**（当前 `/health` 未使用）:

```go
// 成功响应
response.Success(c, data)  // 返回 {"code": 0, "message": "success", "data": {...}}

// 错误响应
response.Error(c, http.StatusBadRequest, "invalid input")
```

**Swagger 注释**（使用通用封装时）:
```go
// @Success 200 {object} response.Response{data=AssetResponse}
```

## 错误处理

**参考**: `internal/errors/errors.go:18-28`, `internal/middleware/error.go:11-34`

**规则**:
- 使用自定义 `AppError` 类型
- Handler 中调用 `c.Error(errors.NewXxxError(...))`
- 由 `ErrorHandler` 中间件统一处理

**工厂函数**:
```go
errors.NewNotFoundError("resource not found")
errors.NewBadRequestError("invalid input", err)
errors.NewInternalError(err)
```

## 参数验证

**规则**（基于 Gin 标准）:
- 请求结构体使用 `binding` tag
- 调用 `c.ShouldBindJSON(&req)` 验证

**示例**:
```go
type CreateAssetRequest struct {
    Name     string `json:"name" binding:"required" example:"logo.png"`
    Category string `json:"category" binding:"required,oneof=image video" example:"image"`
}

func (h *AssetHandler) Create(c *gin.Context) {
    var req CreateAssetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(errors.NewBadRequestError("invalid request", err))
        return
    }
    // 处理请求
}
```

## 全局 Swagger 配置

**参考**: `cmd/api/main.go:29-43`

**位置**: `cmd/api/main.go` 文件顶部，`package main` 之后

**必需标签**:
```go
// @title           AssetHub API
// @version         1.0
// @description     资产管理系统 API
// @host            localhost:8003
// @BasePath        /
```

**生成命令**:
```bash
make swag-init  # 或 swag init -g cmd/api/main.go -o docs
```

## 命名约定

| 类型 | 约定 | 示例 |
|------|------|------|
| Handler 结构体 | `*Handler` | `AssetHandler`, `UserHandler` |
| Handler 工厂函数 | `New*Handler` | `NewAssetHandler` |
| 请求结构体 | `*Request` | `CreateAssetRequest` |
| 响应结构体 | `*Response` | `AssetResponse`, `HealthResponse` |
| JSON 字段 | `snake_case` | `created_at`, `user_id` |
| 路由路径 | `/lowercase` | `/assets`, `/health` |

## 关键检查清单

**新增 API 端点时**:
- [ ] 定义专用的 Request/Response 结构体（不使用 `map[string]interface{}`）
- [ ] 所有字段包含 `json` 和 `example` tag
- [ ] 添加完整的 Swagger 注释（`@Summary`, `@Router` 等）
- [ ] 使用工厂函数创建 handler
- [ ] 在 `setupRouter()` 中注册路由
- [ ] 运行 `make swag-init` 重新生成文档
- [ ] 验证 Swagger UI 显示正确的字段（不是 `additionalProp`）
