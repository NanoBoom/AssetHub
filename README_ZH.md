# AssetHub

企业级文件存储微服务，提供统一的多存储后端 API。

[English Documentation](README.md)

## 功能特性

- ✅ **统一存储接口** - 抽象 S3/OSS/本地存储，单一 API 适配多种后端
- ✅ **直接上传** - 小文件后端代理上传
- ✅ **预签名上传** - 前端直传，使用预签名 URL
- ✅ **分片上传** - 大文件分片上传（支持 GB 级视频）
- ✅ **元数据管理** - 完整的文件信息存储和查询
- ✅ **RESTful API** - 标准 HTTP 接口，带 Swagger 文档
- ✅ **健康检查** - 数据库和 Redis 连接监控

## 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | Gin v1.11.0 |
| ORM | GORM v1.31.1 |
| 数据库 | PostgreSQL 14+ |
| 缓存 | Redis 6+ (go-redis v9) |
| 存储 | AWS S3 SDK v2 / 阿里云 OSS SDK |
| 配置 | Viper v1.21.0 |
| 日志 | Zap |
| 文档 | Swagger (swaggo) |

## 项目架构

```
AssetHub/
├── cmd/api/main.go              # 应用入口
├── internal/
│   ├── config/                  # 配置加载 (Viper)
│   ├── database/                # PostgreSQL 连接
│   ├── cache/                   # Redis 客户端
│   ├── handlers/                # HTTP 处理器
│   │   ├── health.go            # 健康检查
│   │   └── file_handler.go      # 文件操作
│   ├── services/                # 业务逻辑层
│   │   └── file_service.go      # 文件服务
│   ├── repositories/            # 数据访问层
│   │   └── file_repository.go   # 文件仓储
│   ├── models/                  # 数据模型
│   │   ├── base.go              # 基础模型 (ID, 时间戳)
│   │   └── file.go              # 文件模型
│   ├── middleware/              # 中间件
│   │   ├── cors.go              # CORS
│   │   ├── error.go             # 错误处理
│   │   ├── logger.go            # 请求日志
│   │   └── recovery.go          # Panic 恢复
│   ├── errors/                  # 自定义错误
│   └── logger/                  # 日志初始化
├── pkg/
│   ├── response/                # 统一响应格式
│   └── storage/                 # 存储抽象层
│       ├── interface.go         # 存储接口
│       ├── s3.go                # S3 实现
│       ├── oss.go               # 阿里云 OSS 实现
│       └── local.go             # 本地文件系统
├── configs/
│   └── config.yaml              # 默认配置（不含敏感信息）
├── .env.example                 # 环境变量模板
├── Makefile                     # 构建命令
└── go.mod
```

## 环境要求

- Go 1.24+
- PostgreSQL 14+
- Redis 6+
- AWS S3 / 阿里云 OSS / MinIO（可选，可使用本地存储）

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/NanoBoom/asethub.git
cd asethub
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置数据库

```bash
# 创建数据库
psql -h localhost -U postgres -c "CREATE DATABASE assethub;"

# 或使用 Docker
make db-create
```

### 4. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 填入你的凭证
vim .env
```

**配置优先级**（从高到低）：
1. 环境变量（`.env` 文件）
2. `configs/config.yaml`（默认值）
3. Viper 默认值（硬编码）

**存储后端选择**：

```bash
# 使用阿里云 OSS（生产环境推荐）
STORAGE_TYPE=oss
OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
OSS_BUCKET=your-bucket
OSS_ACCESS_KEY_ID=your-key-id
OSS_ACCESS_KEY_SECRET=your-secret

# 使用 AWS S3
STORAGE_TYPE=s3
S3_REGION=us-east-1
S3_BUCKET=your-bucket
S3_ACCESS_KEY_ID=your-key-id
S3_SECRET_ACCESS_KEY=your-secret

# 使用本地存储（仅开发环境）
STORAGE_TYPE=local
LOCAL_BASE_PATH=./storage
```

### 5. 运行应用

```bash
# 开发模式（热重载）
make dev

# 或直接运行
make run
```

服务启动在 `http://localhost:8003`（可通过 `APP_PORT` 配置）。

### 6. 验证

```bash
# 健康检查
curl http://localhost:8003/health

# Swagger UI
open http://localhost:8003/swagger/index.html
```

## API 端点

### 健康检查

- `GET /health` - 检查数据库和 Redis 连接状态

### 文件上传

- `POST /files/upload/direct` - 直接上传（后端代理）
- `POST /files/upload/presigned/init` - 初始化预签名上传
- `POST /files/upload/presigned/confirm` - 确认预签名上传
- `POST /files/upload/multipart/init` - 初始化分片上传
- `POST /files/upload/multipart/part-url` - 生成分片上传 URL
- `POST /files/upload/multipart/complete` - 完成分片上传

### 文件管理

- `GET /files/:id/download-url` - 获取下载 URL（预签名）
- `GET /files/:id` - 获取文件元数据
- `DELETE /files/:id` - 删除文件

完整 API 文档：`http://localhost:8003/swagger/index.html`

## 开发指南

### Makefile 命令

```bash
make help       # 显示所有命令
make build      # 构建二进制到 bin/api
make run        # 运行应用
make dev        # 热重载运行（需要 air）
make test       # 运行测试
make lint       # 运行 golangci-lint
make clean      # 清理构建产物
make swag-init  # 生成 Swagger 文档
```

### 分层架构

请求流程：

```
HTTP 请求 → 中间件 → Handler → Service → Repository → 数据库
                                  ↓
                               Storage (S3/OSS/Local)
                                  ↓
                               Cache (Redis)
```

- **Handler**：HTTP 请求处理，参数验证
- **Service**：业务逻辑，事务管理
- **Repository**：数据访问，数据库操作
- **Storage**：文件存储抽象

### 添加新功能

1. 在 `internal/models/` 定义数据模型
2. 在 `internal/repositories/` 实现数据访问层
3. 在 `internal/services/` 实现业务逻辑
4. 在 `internal/handlers/` 实现 HTTP 处理器
5. 在 `cmd/api/main.go` 注册路由
6. 添加 Swagger 注释
7. 运行 `make swag-init`

### 配置说明

| 配置键 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `app.name` | `APP_NAME` | AssetHub | 应用名称 |
| `app.port` | `APP_PORT` | 8080 | HTTP 端口 |
| `app.env` | `APP_ENV` | development | 运行环境 |
| `database.host` | `DB_HOST` | localhost | PostgreSQL 主机 |
| `database.port` | `DB_PORT` | 5432 | PostgreSQL 端口 |
| `database.user` | `DB_USER` | postgres | 数据库用户 |
| `database.password` | `DB_PASSWORD` | - | 数据库密码 |
| `database.dbname` | `DB_NAME` | assethub | 数据库名 |
| `redis.host` | `REDIS_HOST` | localhost | Redis 主机 |
| `redis.port` | `REDIS_PORT` | 6379 | Redis 端口 |
| `redis.db` | `REDIS_DB` | 2 | Redis 数据库 |
| `storage.type` | `STORAGE_TYPE` | oss | 存储后端 (s3/oss/local) |

## 连接字符串

| 服务 | 连接字符串 |
|------|-----------|
| PostgreSQL | `postgresql://postgres:postgres@localhost:5432/assethub` |
| Redis | `redis://localhost:6379/2` |

## 许可证

MIT
