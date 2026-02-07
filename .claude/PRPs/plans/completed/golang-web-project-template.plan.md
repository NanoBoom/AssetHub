# Feature: Golang Web 项目基础结构模板

## Summary

为 AssetHub（图片、视频资产管理系统）构建一个基于 Gin + GORM + Redis 的 Golang Web 项目基础结构模板。采用分层架构（Handler → Service → Repository），集成配置管理、日志、错误处理、中间件、Redis 缓存等核心功能模块，遵循 golang-standards/project-layout 社区规范。

## User Story

As a Golang 开发者
I want to 有一个标准化的 Web 项目结构模板
So that 可以快速启动 AssetHub 项目开发，遵循最佳实践，减少重复配置工作

## Problem Statement

当前项目是空项目，需要从零开始搭建完整的 Golang Web 项目基础架构，包括目录结构、核心依赖、配置管理、日志系统、错误处理、中间件、Redis 缓存等基础设施。

## Solution Statement

采用 Gin + GORM + Redis + Viper + Zap 技术栈，构建一个分层架构的 Web 项目模板：
- **入口层**: cmd/api/main.go 应用启动入口
- **接口层**: internal/handlers HTTP 请求处理
- **业务层**: internal/services 业务逻辑
- **数据层**: internal/repositories 数据访问
- **缓存层**: internal/cache Redis 缓存
- **基础设施**: internal/config, internal/logger, internal/middleware

## Metadata

| Field            | Value                                                        |
| ---------------- | ------------------------------------------------------------ |
| Type             | NEW_CAPABILITY                                               |
| Complexity       | MEDIUM                                                       |
| Systems Affected | 整个项目基础架构                                              |
| Dependencies     | gin v1.11.0, gorm v1.31.1, go-redis v9, viper v1.21.0, zap  |
| Estimated Tasks  | 25                                                           |

## Infrastructure

**已部署的服务 (Docker):**

| Service    | Connection String                                    |
| ---------- | ---------------------------------------------------- |
| PostgreSQL | `postgresql://postgres:postgres@localhost:5432/assethub` |
| Redis      | `redis://localhost:6379/2`                           |

**注意**: 需要创建 `assethub` 数据库

---

## UX Design

### Before State
```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                              BEFORE STATE                                      ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║   ┌─────────────┐                                                             ║
║   │   空项目    │                                                             ║
║   │  AssetHub/  │                                                             ║
║   │  ├─ .claude │                                                             ║
║   │  ├─ .git    │                                                             ║
║   │  └─ README  │                                                             ║
║   └─────────────┘                                                             ║
║                                                                               ║
║   PAIN_POINT: 无项目结构，无法开始开发                                         ║
║   DATA_FLOW:  无                                                              ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝
```

### After State
```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                               AFTER STATE                                      ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║                                                                               ║
║   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌────────────┐ ║
║   │   HTTP      │────►│   Handler   │────►│   Service   │────►│ Repository │ ║
║   │   Request   │     │    Layer    │     │    Layer    │     │   Layer    │ ║
║   └─────────────┘     └─────────────┘     └─────────────┘     └────────────┘ ║
║         │                   │                   │                   │        ║
║         ▼                   ▼                   ▼                   ▼        ║
║   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌────────────┐ ║
║   │ Middleware  │     │   Logger    │     │   Cache     │     │   GORM     │ ║
║   │ (Recovery,  │     │   (Zap)     │     │  (Redis)    │     │ (Postgres) │ ║
║   │  CORS, etc) │     │             │     │             │     │            │ ║
║   └─────────────┘     └─────────────┘     └─────────────┘     └────────────┘ ║
║                                                                               ║
║   USER_FLOW: HTTP Request → Middleware → Handler → Service → Cache/DB        ║
║   VALUE_ADD: 标准化分层架构 + Redis 缓存，可维护、可测试、可扩展               ║
║   DATA_FLOW: Request → Validation → Business Logic → Cache → Database        ║
║                                                                               ║
╚═══════════════════════════════════════════════════════════════════════════════╝
```

### Project Structure After
```
AssetHub/
├── cmd/
│   └── api/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载
│   ├── database/
│   │   └── database.go          # PostgreSQL 数据库连接
│   ├── cache/
│   │   └── redis.go             # Redis 缓存客户端
│   ├── handlers/
│   │   ├── handler.go           # Handler 基础结构
│   │   └── health.go            # 健康检查
│   ├── middleware/
│   │   ├── cors.go              # CORS 中间件
│   │   ├── error.go             # 错误处理中间件
│   │   ├── logger.go            # 请求日志中间件
│   │   └── recovery.go          # Panic 恢复中间件
│   ├── models/
│   │   └── base.go              # 基础模型定义
│   ├── repositories/
│   │   └── repository.go        # Repository 基础接口
│   ├── services/
│   │   └── service.go           # Service 基础接口
│   ├── errors/
│   │   └── errors.go            # 自定义错误类型
│   └── logger/
│       └── logger.go            # 日志初始化
├── pkg/
│   └── response/
│       └── response.go          # 统一响应格式
├── configs/
│   ├── config.yaml              # 默认配置
│   └── config.example.yaml      # 配置示例
├── scripts/
│   ├── setup.sh                 # 初始化脚本
│   └── init_db.sql              # 数据库初始化 SQL
├── .env.example                 # 环境变量示例
├── .gitignore                   # Git 忽略文件
├── go.mod                       # Go 模块定义
├── go.sum                       # 依赖锁定
├── Makefile                     # 构建命令
└── README.md                    # 项目说明
```

---

## Mandatory Reading

**CRITICAL: Implementation agent MUST read these files before starting any task:**

| Priority | File | Why Read This |
|----------|------|---------------|
| P0 | `.claude/PRPs/ai_docs/golang_web_stack_guide.md` | 完整的技术栈参考指南，包含所有代码模式 |

**External Documentation:**
| Source | Section | Why Needed |
|--------|---------|------------|
| [Gin Docs](https://gin-gonic.com/en/docs/) | Quickstart, Examples | 框架基础用法 |
| [GORM Docs](https://gorm.io/docs/) | Connecting, CRUD | ORM 操作模式 |
| [go-redis Docs](https://redis.uptrace.dev/guide/) | Getting Started | Redis 客户端用法 |
| [Viper GitHub](https://github.com/spf13/viper) | Usage | 配置管理 |
| [Zap Docs](https://pkg.go.dev/go.uber.org/zap) | Logger | 日志系统 |

---

## Patterns to Mirror

**PROJECT_STRUCTURE:**
```
// 遵循 golang-standards/project-layout 社区规范
// /cmd - 应用入口，保持精简
// /internal - 私有代码，Go 强制禁止外部导入
// /pkg - 可复用的公共库
// /configs - 配置文件模板
```

**NAMING_CONVENTION:**
```go
// 文件名: snake_case (user_handler.go)
// 包名: lowercase (handlers, services)
// 接口名: 以 er 结尾或描述性名称 (UserService, Repository)
// 构造函数: New + 类型名 (NewUserHandler)
```

**ERROR_HANDLING:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:586-626
// COPY THIS PATTERN:
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

**LOGGING_PATTERN:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:525-564
// COPY THIS PATTERN:
func InitLogger(env string) (*zap.Logger, error) {
    var config zap.Config
    if env == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    return config.Build()
}
```

**REDIS_PATTERN:**
```go
// Redis 客户端初始化模式
import "github.com/redis/go-redis/v9"

func NewRedisClient(cfg *RedisConfig) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %w", err)
    }

    return client, nil
}
```

**HANDLER_PATTERN:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:944-996
// COPY THIS PATTERN:
type UserHandler struct {
    service services.UserService
    logger  *zap.Logger
}

func NewUserHandler(service services.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) Get(c *gin.Context) {
    id := c.Param("id")
    user, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }
    c.JSON(http.StatusOK, user)
}
```

**SERVICE_PATTERN:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:998-1045
// COPY THIS PATTERN:
type UserService interface {
    GetByID(ctx context.Context, id string) (*models.User, error)
}

type userService struct {
    repo  repositories.UserRepository
    cache *redis.Client  // 添加 Redis 缓存
}

func NewUserService(repo repositories.UserRepository, cache *redis.Client) UserService {
    return &userService{repo: repo, cache: cache}
}
```

**REPOSITORY_PATTERN:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:1047-1091
// COPY THIS PATTERN:
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*models.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return &user, nil
}
```

**MIDDLEWARE_PATTERN:**
```go
// SOURCE: .claude/PRPs/ai_docs/golang_web_stack_guide.md:670-808
// 中间件顺序: Recovery → Logger → Error Handler → CORS → Business Logic
```

---

## Files to Change

| File                                | Action | Justification                       |
| ----------------------------------- | ------ | ----------------------------------- |
| `go.mod`                            | CREATE | Go 模块定义和依赖管理                |
| `Makefile`                          | CREATE | 构建和开发命令                       |
| `.gitignore`                        | CREATE | Git 忽略规则                         |
| `.env.example`                      | CREATE | 环境变量模板                         |
| `configs/config.yaml`               | CREATE | 默认配置文件                         |
| `configs/config.example.yaml`       | CREATE | 配置示例说明                         |
| `scripts/init_db.sql`               | CREATE | 数据库初始化 SQL                     |
| `internal/config/config.go`         | CREATE | 配置加载逻辑                         |
| `internal/logger/logger.go`         | CREATE | Zap 日志初始化                       |
| `internal/errors/errors.go`         | CREATE | 自定义错误类型                       |
| `internal/database/database.go`     | CREATE | GORM PostgreSQL 数据库连接           |
| `internal/cache/redis.go`           | CREATE | Redis 缓存客户端                     |
| `internal/models/base.go`           | CREATE | 基础模型定义                         |
| `internal/middleware/recovery.go`   | CREATE | Panic 恢复中间件                     |
| `internal/middleware/logger.go`     | CREATE | 请求日志中间件                       |
| `internal/middleware/error.go`      | CREATE | 错误处理中间件                       |
| `internal/middleware/cors.go`       | CREATE | CORS 中间件                          |
| `internal/handlers/health.go`       | CREATE | 健康检查 Handler                     |
| `internal/repositories/repository.go` | CREATE | Repository 基础接口                |
| `internal/services/service.go`      | CREATE | Service 基础接口                     |
| `pkg/response/response.go`          | CREATE | 统一响应格式                         |
| `cmd/api/main.go`                   | CREATE | 应用入口                             |
| `scripts/setup.sh`                  | CREATE | 初始化脚本                           |

---

## NOT Building (Scope Limits)

Explicit exclusions to prevent scope creep:

- **用户认证模块** - 不在本次基础模板范围内，后续需求单独实现
- **具体业务实体** - 只提供示例结构，不实现具体的 Asset 业务逻辑
- **数据库迁移工具** - 使用 GORM AutoMigrate，不集成专门的迁移工具
- **API 文档生成** - 不集成 Swagger，后续按需添加
- **容器化配置** - 不包含 Dockerfile 和 docker-compose
- **CI/CD 配置** - 不包含 GitHub Actions 等 CI 配置
- **前端资源** - 不包含 web/static 和 web/templates

---

## Step-by-Step Tasks

Execute in order. Each task is atomic and independently verifiable.

### Task 0: CREATE `assethub` 数据库

- **ACTION**: 在 PostgreSQL 中创建 assethub 数据库
- **IMPLEMENT**:
  ```bash
  docker exec -it <postgres_container> psql -U postgres -c "CREATE DATABASE assethub;"
  ```
  或使用 psql:
  ```bash
  psql -h localhost -U postgres -c "CREATE DATABASE assethub;"
  ```
- **VALIDATE**: 数据库创建成功，可以连接

### Task 1: CREATE `go.mod`

- **ACTION**: CREATE Go 模块定义文件
- **IMPLEMENT**:
  - 模块名: `github.com/NanoBoom/asethub`
  - Go 版本: 1.23+
- **VALIDATE**: `go mod tidy` 能正常执行

```go
module github.com/NanoBoom/asethub

go 1.23
```

### Task 2: CREATE `.gitignore`

- **ACTION**: CREATE Git 忽略文件
- **IMPLEMENT**: 忽略二进制、IDE、环境变量等文件
- **VALIDATE**: 文件存在

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary
*.test

# Output of go coverage tool
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo

# Environment
.env
*.local

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
```

### Task 3: CREATE `.env.example`

- **ACTION**: CREATE 环境变量模板
- **IMPLEMENT**: 定义所有需要的环境变量，包含 Redis 配置
- **VALIDATE**: 文件存在

```bash
# Application
APP_ENV=development
APP_PORT=8080
APP_NAME=AssetHub

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=assethub
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=2

# Logging
LOG_LEVEL=debug
```

### Task 4: CREATE `configs/config.yaml`

- **ACTION**: CREATE 默认配置文件
- **IMPLEMENT**: YAML 格式配置，包含 app、database、redis、log 配置段
- **VALIDATE**: YAML 语法正确

```yaml
app:
  name: "AssetHub"
  port: 8080
  env: "development"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "assethub"
  sslmode: "disable"
  max_open_conns: 10
  max_idle_conns: 5

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 2
  pool_size: 10

log:
  level: "debug"
  format: "console"
```

### Task 5: CREATE `configs/config.example.yaml`

- **ACTION**: CREATE 配置说明文件
- **IMPLEMENT**: 包含所有配置项的注释说明
- **VALIDATE**: 文件存在

### Task 6: CREATE `scripts/init_db.sql`

- **ACTION**: CREATE 数据库初始化 SQL
- **IMPLEMENT**: 创建数据库和初始扩展
- **VALIDATE**: 文件存在

```sql
-- 创建数据库 (如果不存在)
-- 注意: 此 SQL 需要在 postgres 数据库中执行
-- CREATE DATABASE assethub;

-- 连接到 assethub 数据库后执行以下内容
-- 启用 UUID 扩展 (可选)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 启用 pg_trgm 扩展用于模糊搜索 (可选)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

### Task 7: CREATE `internal/config/config.go`

- **ACTION**: CREATE 配置加载模块
- **IMPLEMENT**:
  - 使用 Viper 加载配置
  - 支持环境变量覆盖
  - 定义 Config 结构体，包含 Redis 配置
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:350-475`
- **IMPORTS**: `github.com/spf13/viper`
- **VALIDATE**: `go build ./internal/config/...`

```go
// Config 结构体需包含 Redis 配置
type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
    PoolSize int    `mapstructure:"pool_size"`
}
```

### Task 8: CREATE `internal/logger/logger.go`

- **ACTION**: CREATE 日志初始化模块
- **IMPLEMENT**:
  - 使用 Zap 日志库
  - 开发环境彩色输出
  - 生产环境 JSON 格式
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:525-564`
- **IMPORTS**: `go.uber.org/zap`, `go.uber.org/zap/zapcore`
- **VALIDATE**: `go build ./internal/logger/...`

### Task 9: CREATE `internal/errors/errors.go`

- **ACTION**: CREATE 自定义错误类型
- **IMPLEMENT**:
  - AppError 结构体
  - 常用错误构造函数 (NotFound, BadRequest, Internal)
  - 实现 error 接口
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:586-626`
- **VALIDATE**: `go build ./internal/errors/...`

### Task 10: CREATE `internal/database/database.go`

- **ACTION**: CREATE 数据库连接模块
- **IMPLEMENT**:
  - GORM PostgreSQL 连接
  - 连接池配置
  - 健康检查方法
  - DSN: `postgresql://postgres:postgres@localhost:5432/assethub`
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:209-225`
- **IMPORTS**: `gorm.io/gorm`, `gorm.io/driver/postgres`
- **GOTCHA**: 使用 GORM v2 导入路径 `gorm.io/gorm`，不是旧的 `github.com/jinzhu/gorm`
- **VALIDATE**: `go build ./internal/database/...`

### Task 11: CREATE `internal/cache/redis.go`

- **ACTION**: CREATE Redis 缓存客户端模块
- **IMPLEMENT**:
  - 使用 go-redis v9
  - 连接配置: `redis://localhost:6379/2`
  - 连接池配置
  - 健康检查方法
  - 封装常用缓存操作 (Get, Set, Delete, Exists)
- **IMPORTS**: `github.com/redis/go-redis/v9`
- **VALIDATE**: `go build ./internal/cache/...`

```go
package cache

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/NanoBoom/asethub/internal/config"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
        PoolSize: cfg.PoolSize,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %w", err)
    }

    return &RedisClient{client: client}, nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
    return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
    return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisClient) Close() error {
    return r.client.Close()
}

func (r *RedisClient) Ping(ctx context.Context) error {
    return r.client.Ping(ctx).Err()
}

// Client 返回原始 redis.Client，用于高级操作
func (r *RedisClient) Client() *redis.Client {
    return r.client
}
```

### Task 12: CREATE `internal/models/base.go`

- **ACTION**: CREATE 基础模型定义
- **IMPLEMENT**:
  - BaseModel 包含 ID, CreatedAt, UpdatedAt, DeletedAt
  - 使用 gorm.Model 或自定义
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:227-244`
- **VALIDATE**: `go build ./internal/models/...`

### Task 13: CREATE `internal/middleware/recovery.go`

- **ACTION**: CREATE Panic 恢复中间件
- **IMPLEMENT**:
  - 捕获 panic
  - 记录错误日志
  - 返回 500 响应
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:717-736`
- **IMPORTS**: `github.com/gin-gonic/gin`, `go.uber.org/zap`
- **VALIDATE**: `go build ./internal/middleware/...`

### Task 14: CREATE `internal/middleware/logger.go`

- **ACTION**: CREATE 请求日志中间件
- **IMPLEMENT**:
  - 记录请求方法、路径、状态码、耗时
  - 使用 Zap 结构化日志
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:740-761`
- **VALIDATE**: `go build ./internal/middleware/...`

### Task 15: CREATE `internal/middleware/error.go`

- **ACTION**: CREATE 错误处理中间件
- **IMPLEMENT**:
  - 处理 c.Errors 中的错误
  - 根据错误类型返回对应 HTTP 状态码
  - 统一错误响应格式
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:673-708`
- **VALIDATE**: `go build ./internal/middleware/...`

### Task 16: CREATE `internal/middleware/cors.go`

- **ACTION**: CREATE CORS 中间件
- **IMPLEMENT**:
  - 设置 Access-Control-Allow-* 头
  - 处理 OPTIONS 预检请求
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:764-782`
- **VALIDATE**: `go build ./internal/middleware/...`

### Task 17: CREATE `internal/handlers/health.go`

- **ACTION**: CREATE 健康检查 Handler
- **IMPLEMENT**:
  - GET /health 端点
  - 返回 {"status": "ok"}
  - 包含数据库和 Redis 连接检查
- **VALIDATE**: `go build ./internal/handlers/...`

```go
// 健康检查应包含 DB 和 Redis 状态
type HealthResponse struct {
    Status   string `json:"status"`
    Database string `json:"database"`
    Redis    string `json:"redis"`
}
```

### Task 18: CREATE `internal/repositories/repository.go`

- **ACTION**: CREATE Repository 基础接口
- **IMPLEMENT**:
  - 定义通用 Repository 接口模式
  - 提供示例注释
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:1047-1091`
- **VALIDATE**: `go build ./internal/repositories/...`

### Task 19: CREATE `internal/services/service.go`

- **ACTION**: CREATE Service 基础接口
- **IMPLEMENT**:
  - 定义通用 Service 接口模式
  - 提供示例注释
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:998-1045`
- **VALIDATE**: `go build ./internal/services/...`

### Task 20: CREATE `pkg/response/response.go`

- **ACTION**: CREATE 统一响应格式
- **IMPLEMENT**:
  - Success 响应结构
  - Error 响应结构
  - 辅助函数
- **VALIDATE**: `go build ./pkg/response/...`

### Task 21: CREATE `cmd/api/main.go`

- **ACTION**: CREATE 应用入口文件
- **IMPLEMENT**:
  - 加载配置
  - 初始化日志
  - 初始化数据库 (PostgreSQL)
  - 初始化 Redis 缓存
  - 设置路由和中间件
  - 启动 HTTP 服务器
  - 优雅关闭 (包括关闭 DB 和 Redis 连接)
- **MIRROR**: `.claude/PRPs/ai_docs/golang_web_stack_guide.md:824-939`
- **VALIDATE**: `go build ./cmd/api/...`

### Task 22: CREATE `Makefile`

- **ACTION**: CREATE 构建命令文件
- **IMPLEMENT**:
  - build: 编译应用
  - run: 运行应用
  - test: 运行测试
  - lint: 代码检查
  - clean: 清理编译产物
  - db-create: 创建数据库
- **VALIDATE**: `make help` 或 `make build`

```makefile
.PHONY: help build run test lint clean db-create

help:
	@echo "Available commands:"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make test      - Run tests"
	@echo "  make lint      - Run linter"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make db-create - Create database"

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

db-create:
	psql -h localhost -U postgres -c "CREATE DATABASE assethub;" || true
```

### Task 23: CREATE `scripts/setup.sh`

- **ACTION**: CREATE 初始化脚本
- **IMPLEMENT**:
  - 检查 Go 版本
  - 下载依赖
  - 创建数据库
  - 复制配置文件
  - 提示下一步操作
- **VALIDATE**: `chmod +x scripts/setup.sh`

### Task 24: 安装依赖并验证

- **ACTION**: RUN 依赖安装和编译验证
- **IMPLEMENT**:
  ```bash
  go mod tidy
  go build ./...
  ```
- **VALIDATE**: 编译无错误

### Task 25: 验证连接

- **ACTION**: RUN 验证数据库和 Redis 连接
- **IMPLEMENT**:
  ```bash
  go run cmd/api/main.go
  curl http://localhost:8080/health
  ```
- **VALIDATE**: 健康检查返回数据库和 Redis 状态正常

---

## Testing Strategy

### Unit Tests to Write (Future)

| Test File                                | Test Cases                 | Validates      |
| ---------------------------------------- | -------------------------- | -------------- |
| `internal/config/config_test.go`         | 配置加载、默认值           | Config 模块    |
| `internal/errors/errors_test.go`         | 错误构造、Error() 方法     | Error 类型     |
| `internal/cache/redis_test.go`           | Redis 连接、CRUD 操作      | Redis 模块     |
| `internal/middleware/*_test.go`          | 中间件行为                 | 中间件        |

### Edge Cases Checklist

- [ ] 配置文件不存在时使用默认值
- [ ] 数据库连接失败时的错误处理
- [ ] Redis 连接失败时的错误处理
- [ ] 环境变量覆盖配置文件值
- [ ] Panic 恢复并返回 500
- [ ] 无效 JSON 请求体处理

---

## Validation Commands

| Level | Command | Expect |
|-------|---------|--------|
| Type Check | `go build ./...` | Exit 0 |
| Vet | `go vet ./...` | Exit 0 |
| Lint | `golangci-lint run` (如已安装) | Exit 0 |
| Test | `go test ./...` | All pass |
| Run | `go run cmd/api/main.go` | Server starts |

### Level 1: STATIC_ANALYSIS

```bash
go build ./...
go vet ./...
```
**EXPECT**: Exit 0, no errors

### Level 2: UNIT_TESTS

```bash
go test ./...
```
**EXPECT**: All tests pass (初始模板可能无测试)

### Level 3: BUILD

```bash
go build -o bin/api cmd/api/main.go
```
**EXPECT**: 生成可执行文件 bin/api

### Level 4: MANUAL_VALIDATION

1. 创建数据库: `psql -h localhost -U postgres -c "CREATE DATABASE assethub;"`
2. 启动服务: `go run cmd/api/main.go`
3. 访问健康检查: `curl http://localhost:8080/health`
4. 预期响应: `{"status":"ok","database":"ok","redis":"ok"}`

---

## Acceptance Criteria

- [ ] 项目结构符合 golang-standards/project-layout 规范
- [ ] `go build ./...` 编译通过
- [ ] `go vet ./...` 无警告
- [ ] 服务器能正常启动
- [ ] /health 端点返回正确响应（包含 DB 和 Redis 状态）
- [ ] 日志按环境配置正确输出
- [ ] 配置能从文件和环境变量加载
- [ ] 中间件正确处理请求
- [ ] PostgreSQL 连接正常
- [ ] Redis 连接正常

---

## Completion Checklist

- [ ] Task 0: 创建 assethub 数据库
- [ ] Task 1-6: 项目配置文件 (go.mod, .gitignore, configs, init_db.sql)
- [ ] Task 7-12: 核心基础设施 (config, logger, errors, database, cache, models)
- [ ] Task 13-16: 中间件层 (recovery, logger, error, cors)
- [ ] Task 17-20: 业务层骨架 (handlers, repositories, services, response)
- [ ] Task 21-23: 入口和构建 (main.go, Makefile, setup.sh)
- [ ] Task 24-25: 依赖安装和连接验证
- [ ] Level 1: 静态分析通过
- [ ] Level 3: 构建成功
- [ ] Level 4: 手动验证通过
- [ ] 所有验收标准满足

---

## Risks and Mitigations

| Risk               | Likelihood   | Impact       | Mitigation                              |
| ------------------ | ------------ | ------------ | --------------------------------------- |
| Go 版本不兼容      | LOW          | HIGH         | 在 go.mod 中明确要求 go 1.23+           |
| 依赖版本冲突       | LOW          | MEDIUM       | 使用 go mod tidy 自动解决               |
| PostgreSQL 连接失败 | LOW         | HIGH         | Docker 已部署，提供清晰的连接信息       |
| Redis 连接失败     | LOW          | MEDIUM       | Docker 已部署，健康检查会提示状态       |
| golangci-lint 未装 | MEDIUM       | LOW          | Makefile 中 lint 命令设为可选           |

---

## Notes

### 技术选型理由

1. **Gin vs Echo/Fiber/Chi**: Gin 生态最成熟，社区支持最好，性能优秀
2. **GORM vs sqlx/Ent**: GORM 功能全面，学习曲线平缓，适合快速开发
3. **go-redis v9**: Redis 官方推荐的 Go 客户端，性能优秀，API 简洁
4. **Zap vs Zerolog**: Zap 自定义能力更强，支持多种输出格式
5. **Viper**: Go 生态中最流行的配置管理库，功能全面

### 连接信息

- **PostgreSQL**: `postgresql://postgres:postgres@localhost:5432/assethub`
- **Redis**: `redis://localhost:6379/2`

### 后续扩展建议

1. **认证模块**: 添加 JWT 中间件和用户认证
2. **API 文档**: 集成 swaggo/swag 生成 Swagger 文档
3. **数据库迁移**: 考虑 golang-migrate/migrate 进行版本化迁移
4. **容器化**: 添加 Dockerfile 和 docker-compose.yaml
5. **CI/CD**: 添加 GitHub Actions 工作流
6. **分布式缓存**: Redis 集群支持
7. **消息队列**: 添加 Redis Pub/Sub 或专用 MQ

### 目录结构说明

- `/internal` 是 Go 特殊目录，外部包无法导入，保护私有代码
- `/internal/cache` 新增的 Redis 缓存模块
- `/pkg` 是可复用的公共代码，可被外部项目导入
- `/cmd` 保持精简，只做依赖组装，业务逻辑放 `/internal`
