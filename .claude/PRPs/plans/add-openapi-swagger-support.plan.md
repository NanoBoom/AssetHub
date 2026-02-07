# Feature: 添加 OpenAPI 和 Swagger UI 支持

## Summary

为 AssetHub Go Web 应用添加 OpenAPI 3.0 规范支持和交互式 Swagger UI 文档界面。使用 swaggo/swag 工具自动从代码注释生成 API 文档，通过 gin-swagger 中间件在 `/swagger/index.html` 提供可视化文档访问。

## User Story

作为 API 开发者和使用者
我想要自动生成的、可交互的 API 文档
以便快速理解 API 端点、测试接口、减少沟通成本

## Problem Statement

当前 AssetHub 项目缺少 API 文档，开发者需要阅读源代码才能理解接口定义。这导致：
- 新成员上手困难
- 前后端协作效率低
- API 变更难以追踪
- 无法快速测试接口

## Solution Statement

集成 swaggo/swag 工具链，通过代码注释自动生成 OpenAPI 规范文档，并使用 gin-swagger 提供 Swagger UI 界面。文档与代码同步维护，支持在线测试。

## Metadata

| Field            | Value                                             |
| ---------------- | ------------------------------------------------- |
| Type             | NEW_CAPABILITY                                    |
| Complexity       | MEDIUM                                            |
| Systems Affected | HTTP 路由、Handler 层、配置系统                   |
| Dependencies     | swaggo/swag v1.16+, gin-swagger v1.6+, files v1.0+ |
| Estimated Tasks  | 7                                                 |

---

## UX Design

### Before State
```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                              BEFORE STATE                                      ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║   开发者                                                                       ║
║     │                                                                         ║
║     ├─► 想了解 API 端点                                                       ║
║     │     └─► 阅读源代码 (internal/handlers/*.go)                             ║
║     │         └─► 查找路由定义 (cmd/api/main.go)                              ║
║     │             └─► 推测请求/响应格式                                        ║
║     │                                                                         ║
║     ├─► 想测试 API                                                            ║
║     │     └─► 使用 curl/Postman 手动构造请求                                  ║
║     │         └─► 反复试错参数格式                                            ║
║     │                                                                         ║
║     └─► 想分享 API 给前端                                                     ║
║           └─► 口头描述或写文档（容易过期）                                    ║
║                                                                               ║
║   PAIN_POINT:                                                                 ║
║   - 无文档，学习成本高                                                         ║
║   - 手动测试效率低                                                            ║
║   - 文档与代码不同步                                                          ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════════╗
║                               AFTER STATE                                      ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║   开发者                                                                       ║
║     │                                                                         ║
║     ├─► 访问 http://localhost:8003/swagger/index.html                         ║
║     │     │                                                                   ║
║     │     ├─► 查看所有 API 端点列表（按 Tag 分组）                             ║
║     │     │     └─► 展开查看详细参数、响应格式                                 ║
║     │     │                                                                   ║
║     │     ├─► 点击 "Try it out" 在线测试                                      ║
║     │     │     └─► 填写参数 → Execute → 查看实时响应                         ║
║     │     │                                                                   ║
║     │     └─► 下载 OpenAPI JSON/YAML 规范                                     ║
║     │           └─► 导入到 Postman/Insomnia 等工具                            ║
║     │                                                                         ║
║     └─► 分享文档链接给团队                                                    ║
║           └─► 文档自动与代码同步（swag init）                                 ║
║                                                                               ║
║   VALUE_ADD:                                                                  ║
║   - 零学习成本：可视化文档                                                     ║
║   - 在线测试：无需额外工具                                                     ║
║   - 自动同步：注释即文档                                                       ║
║                                                                               ║
║   DATA_FLOW:                                                                  ║
║   代码注释 → swag init → docs/swagger.json → gin-swagger → Swagger UI         ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝
```

### Interaction Changes
| Location                  | Before                     | After                                  | User Impact                    |
| ------------------------- | -------------------------- | -------------------------------------- | ------------------------------ |
| `/swagger/index.html`     | 404 Not Found              | Swagger UI 界面                        | 可访问交互式文档               |
| `/swagger/doc.json`       | 不存在                     | OpenAPI JSON 规范                      | 可下载规范文件                 |
| Handler 函数              | 无注释或简单注释           | 结构化 Swagger 注释                    | 文档自动生成                   |
| 开发流程                  | 代码 → 手动写文档          | 代码注释 → swag init → 自动生成文档    | 文档与代码同步                 |

---

## Mandatory Reading

**CRITICAL: 实施前必须阅读这些文件以理解现有模式：**

| Priority | File                                                      | Lines  | Why Read This                                  |
| -------- | --------------------------------------------------------- | ------ | ---------------------------------------------- |
| P0       | `cmd/api/main.go`                                         | 81-97  | 路由注册模式 - 必须在此处添加 Swagger 路由    |
| P0       | `internal/handlers/health.go`                             | 14-47  | Handler 结构体模式 - 需要为其添加 Swagger 注释 |
| P1       | `internal/middleware/logger.go`                           | 10-27  | 中间件模式 - 理解中间件链顺序                  |
| P1       | `pkg/response/response.go`                                | 9-28   | 统一响应格式 - Swagger 响应模型定义            |
| P2       | `internal/errors/errors.go`                               | 5-28   | 错误类型定义 - Swagger 错误响应模型            |
| P2       | `internal/config/config.go`                               | 7-42   | 配置结构 - 可选扩展 Swagger 配置               |

**External Documentation:**
| Source | Section | Why Needed |
|--------|---------|------------|
| [Swaggo GitHub](https://github.com/swaggo/swag) | README & Examples | 理解注释语法和 CLI 用法 |
| [Gin-Swagger GitHub](https://github.com/swaggo/gin-swagger) | Integration Guide | 理解 Gin 集成方式 |
| [OpenAPI 3.0 Spec](https://swagger.io/specification/) | General Structure | 理解 OpenAPI 规范结构 |

---

## Patterns to Mirror

**ROUTER_REGISTRATION_PATTERN:**
```go
// SOURCE: cmd/api/main.go:81-97
// COPY THIS PATTERN: 在 setupRouter 函数中注册路由
func setupRouter(cfg *config.Config, zapLogger *zap.Logger, db *gorm.DB, redisClient *cache.RedisClient) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 中间件链
	router.Use(middleware.Recovery(zapLogger))
	router.Use(middleware.Logger(zapLogger))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())

	// 路由注册
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	router.GET("/health", healthHandler.Check)

	// 【新增】Swagger 路由将在此处添加

	return router
}
```

**HANDLER_STRUCTURE_PATTERN:**
```go
// SOURCE: internal/handlers/health.go:14-47
// COPY THIS PATTERN: Handler 结构体和方法定义
type HealthHandler struct {
	db    *gorm.DB
	redis *cache.RedisClient
}

func NewHealthHandler(db *gorm.DB, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// 【新增】Swagger 注释将添加在方法上方
func (h *HealthHandler) Check(c *gin.Context) {
	// ... 实现逻辑
	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
```

**RESPONSE_FORMAT_PATTERN:**
```go
// SOURCE: pkg/response/response.go:9-28
// COPY THIS PATTERN: 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
```

**ERROR_HANDLING_PATTERN:**
```go
// SOURCE: internal/errors/errors.go:5-28
// COPY THIS PATTERN: 自定义错误类型
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}
```

---

## Files to Change

| File                                  | Action | Justification                                      |
| ------------------------------------- | ------ | -------------------------------------------------- |
| `go.mod`                              | UPDATE | 添加 swaggo 依赖                                   |
| `cmd/api/main.go`                     | UPDATE | 添加 Swagger 路由和 API 元信息注释                 |
| `internal/handlers/health.go`         | UPDATE | 为 Check 方法添加 Swagger 注释                     |
| `docs/docs.go`                        | CREATE | swag init 自动生成（不手动编辑）                   |
| `docs/swagger.json`                   | CREATE | swag init 自动生成（不手动编辑）                   |
| `docs/swagger.yaml`                   | CREATE | swag init 自动生成（不手动编辑）                   |
| `Makefile`                            | UPDATE | 添加 swag init 命令                                |

---

## NOT Building (Scope Limits)

明确排除以下内容以防止范围蔓延：

- **不添加认证/授权到 Swagger UI** - 当前 API 无认证，Swagger UI 也无需认证。未来如需要，可通过 `@securityDefinitions` 注释添加
- **不为所有 Handler 添加注释** - 仅为现有的 `/health` 端点添加示例注释。其他端点由后续开发时添加
- **不自定义 Swagger UI 主题** - 使用默认主题，避免不必要的复杂性
- **不集成 API 版本管理** - 当前项目无版本需求，保持简单
- **不添加 Swagger 配置到 config.yaml** - 使用代码注释配置，避免配置文件膨胀

---

## Step-by-Step Tasks

按依赖顺序执行。每个任务独立且可验证。

### Task 1: 安装 Swag CLI 工具

- **ACTION**: 安装 swag 命令行工具到 GOPATH/bin
- **IMPLEMENT**: 运行 `go install github.com/swaggo/swag/cmd/swag@latest`
- **VALIDATE**: `swag --version` 输出版本号（预期 v1.16+）
- **GOTCHA**: 确保 `$GOPATH/bin` 在 PATH 中，否则 swag 命令不可用

### Task 2: 添加 Swagger 依赖到 go.mod

- **ACTION**: 添加 gin-swagger 和 files 包
- **IMPLEMENT**:
  ```bash
  go get -u github.com/swaggo/gin-swagger
  go get -u github.com/swaggo/files
  ```
- **IMPORTS**: 这些包将在 main.go 中导入
- **VALIDATE**: `go mod tidy && go build ./cmd/api` 编译成功
- **GOTCHA**: 不要手动编辑 go.mod，使用 `go get` 自动管理版本

### Task 3: 在 main.go 添加 API 元信息注释

- **ACTION**: 在 main.go 顶部添加 Swagger 通用注释
- **IMPLEMENT**: 在 `package main` 下方添加以下注释块：
  ```go
  // @title           AssetHub API
  // @version         1.0
  // @description     资产管理系统 API - 支持图片、视频等资产的上传、管理和检索
  // @termsOfService  http://swagger.io/terms/

  // @contact.name   API Support
  // @contact.email  support@assethub.example.com

  // @license.name  MIT
  // @license.url   https://opensource.org/licenses/MIT

  // @host      localhost:8003
  // @BasePath  /

  // @schemes   http https
  ```
- **MIRROR**: 参考 swaggo 官方示例的注释格式
- **VALIDATE**: 注释格式正确（无语法错误）
- **GOTCHA**:
  - `@host` 必须与实际运行端口一致（从 .env 读取的 APP_PORT=8003）
  - `@BasePath` 当前为 `/`，如果未来添加 `/api/v1` 前缀需修改

### Task 4: 为 health.go 添加 Swagger 注释

- **ACTION**: 为 HealthHandler.Check 方法添加 Swagger 操作注释
- **IMPLEMENT**: 在 `func (h *HealthHandler) Check(c *gin.Context)` 上方添加：
  ```go
  // Check godoc
  // @Summary      健康检查
  // @Description  检查服务、数据库和 Redis 的健康状态
  // @Tags         system
  // @Accept       json
  // @Produce      json
  // @Success      200  {object}  map[string]string  "status: ok/degraded, database: ok/error, redis: ok/error"
  // @Failure      500  {object}  map[string]string  "Internal server error"
  // @Router       /health [get]
  ```
- **MIRROR**: internal/handlers/health.go:23-47 的实际响应格式
- **VALIDATE**: 注释与实际代码行为一致
- **GOTCHA**:
  - `@Success` 的响应类型使用 `map[string]string` 而非自定义结构体（因为代码中使用 `gin.H`）
  - `@Router` 路径必须与 main.go 中注册的路径完全一致

### Task 5: 生成 Swagger 文档

- **ACTION**: 运行 swag init 生成文档
- **IMPLEMENT**: 在项目根目录执行：
  ```bash
  swag init -g cmd/api/mao -o docs
  ```
- **PATTERN**:
  - `-g` 指定包含 API 元信息注释的文件
  - `-o` 指定输出目录（默认 docs/）
- **VALIDATE**:
  - 生成 `docs/docs.go`、`docs/swagger.json`、`docs/swagger.yaml` 三个文件
  - `docs/swagger.json` 包含 `/health` 端点定义
- **GOTCHA**:
  - 如果报错 "cannot find package"，检查 go.mod 中的 module 名称是否正确
  - 每次修改注释后必须重新运行 `swag init`

### Task 6: 在 main.go 注册 Swagger 路由

- **ACTION**: 在 setupRouter 函数中添加 Swagger UI 路由
- **IMPLEMENT**:
  1. 在 import 块添加：
     ```go
     swaggerFiles "github.com/swaggo/files"
     ginSwagger "github.com/swaggo/gin-swagger"
     _ "github.com/NanoBoom/asethub/docs"  // 导入生成的 docs
     ```
  2. 在 `router.GET("/health", healthHandler.Check)` 后添加：
     ```go
     // Swagger 文档路由
     router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
     ```
- **MIRROR**: cmd/api/main.go:81-97 的路由注册模式
- **VALIDATE**: `go build ./cmd/api` 编译成功，无导入错误
- **GOTCHA**:
  - `_ "github.com/NanoBoom/asethub/docs"` 的下划线导入是必需的（初始化 docs 包）
  - 路由路径 `/swagger/*any` 中的 `*any` 是通配符，匹配所有 Swagger UI 资源

### Task 7: 更新 Makefile 添加文档生成命令

- **ACTION**: 在 Makefile 添加 swag 相关命令
- **IMPLEMENT**: 添加以下目标：
  ```makefile
  .PHONY: swag-init
  swag-init:
  	swag init -g cmd/api/main.go -n  .PHONY: swag-fmt
  swag-fmt:
  	swag fmt -g cmd/api/main.go

  .PHONY: docs
  docs: swag-init
  	@echo "Swagger docs generated at docs/"
  ```
- **PATTERN**: 参考 Makefile 中现有的 `.PHONY` 目标格式
- **VALIDATE**:
  - `make swag-init` 成功生成文档
  - `make swag-fmt` 格式化注释
  - `make docs` 执行完整文档生成流程
- **GOTCHA**: Makefile 使用 Tab 缩进，不是空格

---

## Testing Strategy

### Unit Tests to Write

无需为 Swagger 集成编写单元测试。Swagger 是文档工具，通过手动验证即可。

### Edge Cases Checklist

- [ ] Swagger UI 在生产环境是否应该禁用？（当前不禁用，可通过环境变量控制）
- [ ] 如果 docs/ 目录不存在会怎样？（swag init 会自动创建）
- [ ] 如果注释格式错误会怎样？（swag init 会报错，需修复后重新生成）
- [ ] 多个 Handler 使用相同 Tag 会怎样？（正常，Swagger UI 会分组显示）

---

## Validation Commands

### Level 1: STATIC_ANALYSIS

```bash
go mod tidy
go build ./cmd/api
swag init -g cmd/api/main.go -o docs
```

**EXPECT**:
- Exit 0，无编译错误
- docs/ 目录生成三个文件

### Level 2: RUNTIME_VALIDATION

```bash
go run cmd/api/main.go
```

**EXPECT**:
- 服务启动在端口 8003
- 日志显示 "Server starting"

### Level 3: SWAGGER_UI_VALIDATION

**手动测试步骤：**

1. 启动应用：`make run`
2. 浏览器访问：`http://localhost:8003/swagger/index.html`
3. 验证：
   - [ ] Swagger UI 页面正常加载
   - [ ] 显示 "AssetHub API v1.0" 标题
   - [ ] 看到 "system" 标签下的 `/health` 端点
   - [ ] 展开 `/health` 端点，显示详细参数和响应
   - [ ] 点击 "Try it out" → "Execute"
   - [ ] 返回 200 状态码和健康检查 JSON

4. 验证 OpenAPI JSON：
   - 访问：`http://localhost:8003/swagger/doc.json`
   - [ ] 返回有效的 JSON 格式
   - [ ] 包含 `/health` 端点定义

### Level 4: DOCUMENTATION_SYNC_VALIDATION

**验证文档与代码同步：**

1. 修改 health.go 的 Swagger 注释（例如改 Summary）
2. 运行 `make swag-init`
3. 刷新 Swagger UI
4. [ ] 修改立即反映在文档中

---

## Acceptance Criteria

- [ ] Swagger UI 可通过 `/swagger/index.html` 访问
- [ ] OpenAPI JSON 可通过 `/swagger/doc.json` 下载
- [ ] `/health` 端点在 Swagger UI 中正确显示
- [ ] 可在 Swagger UI 中在线测试 `/health` 端点
- [ ] 文档标题、版本、描述与注释一致
- [ ] `make swag-init` 命令可重新生成文档
- [ ] 应用启动无错误，所有现有功能正常

---

## Completion Checklist

- [ ] Task 1: Swag CLI 安装完成
- [ ] Task 2: go.mod 依赖添加完成
- [ ] Task 3: main.go API 元信息注释添加完成
- [ ] Task 4: health.go Swagger 注释添加完成
- [ ] Task 5: docs/ 文件生成完成
- [ ] Task 6: Swagger 路由注册完成
- [ ] Task 7: Makefile 更新完成
- [ ] Level 1: 静态分析通过
- [ ] Level 2: 运行时验证通过
- [ ] Level 3: Swagger UI 验证通过
- [ ] Level 4: 文档同步验证通过
- [ ] 所有验收标准满足

---

## Risks and Mitigations

| Risk                                   | Likelihood | Impact | Mitigation                                                                 |
| -------------------------------------- | ---------- | ------ | -------------------------------------------------------------------------- |
| swag init 与 Go 1.23 不兼容            | LOW        | HIGH   | 使用 @latest 版本，已知兼容性问题在 2026 年已修复                          |
| 忘记运行 swag init 导致文档过期        | MEDIUM     | MEDIUM | 在 Makefile 添加 `make docs` 命令，CI/CD 中自动运行                        |
| Swagger UI 暴露敏感信息                | LOW        | MEDIUM | 当前无敏感端点，未来可通过环境变量在生产环境禁用 Swagger                   |
| 注释格式错误导致文档生成失败           | MEDIUM     | LOW    | swag init 会报错提示，修复后重新生成即可                                   |
| docs/ 目录被误提交到 Git               | LOW        | LOW    | 在 .gitignore 添加 `docs/`（可选，也可提交以便部署时无需重新生成）         |

---

## Notes

### 设计决策

1. **为什么选择 swaggo/swag？**
   - Go 生态最成熟的 Swagger 工具
   - 与 Gin 框架深度集成
   - 注释即文档，无需单独维护
   - 支持 OpenAPI 3.0 标准

2. **为什么不使用 go-swagger？**
   - go-swagger 更适合 contract-first 开发（先写规范再生成代码）
   - swaggo 更适合 code-first 开发（从代码生成文档）
   - 当前项目已有代码，swaggo 更合适

3. **docs/ 目录是否应该提交到 Git？**
   - **建议提交**：部署时无需安装 swag CLI，直接运行即可
   - **不提交**：保持仓库干净，CI/CD 中自动生成
   - **当前选择**：提交（简化部署流程）

4. **生产环境是否应该禁用 Swagger UI？**
   - **当前不禁用**：API 无敏感信息，文档对外公开有助于集成
   - **未来可选**：通过环境变量控制（`if cfg.App.Env != "production"`）

### 未来扩展

- 添加认证支持：使用 `@securityDefinitions` 和 `@Security` 注释
- 添加请求示例：使用 `@Param` 的 `example` 标签
- 添加响应示例：定义结构体并使用 `example` 标签
- 多版本 API：使用 `@BasePath /api/v1` 和 `/api/v2`
- 自定义 Swagger UI：使用 `ginSwagger.Config` 配置主题

### 参考资源

- [Swaggo GitHub](https://github.com/swaggo/swag) - 官方文档和示例
- [Gin-Swagger GitHub](https://github.com/swaggo/gin-swagger) - Gin 集成指南
- [OpenAPI 3.0 规范](https://swagger.io/specification/) - 标准参考
- [Swaggo 注释语法](https://github.com/swaggo/swag#declarative-comments-format) - 完整注释列表
