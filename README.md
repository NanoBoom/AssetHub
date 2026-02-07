# AssetHub

Enterprise-grade file storage microservice with unified API for multiple storage backends.

[中文文档](README_ZH.md)

## Features

- ✅ **Unified Storage Interface** - Abstract S3/OSS/Local storage with single API
- ✅ **Direct Upload** - Backend proxy upload for small files
- ✅ **Presigned Upload** - Frontend direct upload with presigned URLs
- ✅ **Multipart Upload** - Chunked upload for large files (GB-scale videos)
- ✅ **Metadata Management** - Complete file information storage and query
- ✅ **RESTful API** - Standard HTTP interface with Swagger docs
- ✅ **Health Check** - Database and Redis connectivity monitoring

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Web Framework | Gin v1.11.0 |
| ORM | GORM v1.31.1 |
| Database | PostgreSQL 14+ |
| Cache | Redis 6+ (go-redis v9) |
| Storage | AWS S3 SDK v2 / Aliyun OSS SDK |
| Config | Viper v1.21.0 |
| Logger | Zap |
| Docs | Swagger (swaggo) |

## Architecture

```
AssetHub/
├── cmd/api/main.go              # Application entry
├── internal/
│   ├── config/                  # Config loading (Viper)
│   ├── database/                # PostgreSQL connection
│   ├── cache/                   # Redis client
│   ├── handlers/                # HTTP handlers
│   │   ├── health.go            # Health check
│   │   └── file_handler.go      # File operations
│   ├── services/                # Business logic
│   │   └── file_service.go      # File service
│   ├── repositories/            # Data access layer
│   │   └── file_repository.go   # File repository
│   ├── models/                  # Data models
│   │   ├── base.go              # Base model (ID, timestamps)
│   │   └── file.go              # File model
│   ├── middleware/              # Middleware
│   │   ├── cors.go              # CORS
│   │   ├── error.go             # Error handler
│   │   ├── logger.go            # Request logger
│   │   └── recovery.go          # Panic recovery
│   ├── errors/                  # Custom errors
│   └── logger/                  # Logger initialization
├── pkg/
│   ├── response/                # Unified response format
│   └── storage/                 # Storage abstraction
│       ├── interface.go         # Storage interface
│       ├── s3.go                # S3 implementation
│       ├── oss.go               # Aliyun OSS implementation
│       └── local.go             # Local filesystem
├── configs/
│   └── config.yaml              # Default config (no secrets)
├── .env.example                 # Environment variables template
├── Makefile                     # Build commands
└── go.mod
```

## Requirements

- Go 1.24+
- PostgreSQL 14+
- Redis 6+
- AWS S3 / Aliyun OSS / MinIO (optional, can use local storage)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/NanoBoom/asethub.git
cd asethub
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Setup Database

```bash
# Create database
psql -h localhost -U postgres -c "CREATE DATABASE assethub;"

# Or use Docker
make db-create
```

### 4. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your credentials
vim .env
```

**Configuration Priority** (high to low):
1. Environment variables (`.env` file)
2. `configs/config.yaml` (default values)
3. Viper defaults (hardcoded)

**Storage Backend Selection**:

```bash
# Use Aliyun OSS (recommended for production)
STORAGE_TYPE=oss
OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
OSS_BUCKET=your-bucket
OSS_ACCESS_KEY_ID=your-key-id
OSS_ACCESS_KEY_SECRET=your-secret

# Use AWS S3
STORAGE_TYPE=s3
S3_REGION=us-east-1
S3_BUCKET=your-bucket
S3_ACCESS_KEY_ID=your-key-id
S3_SECRET_ACCESS_KEY=your-secret

# Use Local Storage (development only)
STORAGE_TYPE=local
LOCAL_BASE_PATH=./storage
```

### 5. Run Application

```bash
# Development mode (with hot reload)
make dev

# Or run directly
make run
```

Server starts at `http://localhost:8003` (configurable via `APP_PORT`).

### 6. Verify

```bash
# Health check
curl http://localhost:8003/health

# Swagger UI
open http://localhost:8003/swagger/index.html
```

## API Endpoints

### Health Check

- `GET /health` - Check database and Redis connectivity

### File Upload

- `POST /files/upload/direct` - Direct upload (backend proxy)
- `POST /files/upload/presigned/init` - Initialize presigned upload
- `POST /files/upload/presigned/confirm` - Confirm presigned upload
- `POST /files/upload/multipart/init` - Initialize multipart upload
- `POST /files/upload/multipart/part-url` - Generate part upload URL
- `POST /files/upload/multipart/complete` - Complete multipart upload

### File Management

- `GET /files/:id/download-url` - Get download URL (presigned)
- `GET /files/:id` - Get file metadata
- `DELETE /files/:id` - Delete file

Full API documentation: `http://localhost:8003/swagger/index.html`

## Development

### Makefile Commands

```bash
make help       # Show all commands
make build      # Build binary to bin/api
make run        # Run application
make dev        # Run with hot reload (requires air)
make test       # Run tests
make lint       # Run golangci-lint
make clean      # Clean build artifacts
make swag-init  # Generate Swagger docs
```

### Layered Architecture

Request flow:

```
HTTP Request → Middleware → Handler → Service → Repository → Database
                                         ↓
                                      Storage (S3/OSS/Local)
                                         ↓
                                      Cache (Redis)
```

- **Handler**: HTTP request handling, parameter validation
- **Service**: Business logic, transaction management
- **Repository**: Data access, database operations
- **Storage**: File storage abstraction

### Adding New Features

1. Define model in `internal/models/`
2. Implement repository in `internal/repositories/`
3. Implement service in `internal/services/`
4. Implement handler in `internal/handlers/`
5. Register route in `cmd/api/main.go`
6. Add Swagger annotations
7. Run `make swag-init`

### Configuration

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `app.name` | `APP_NAME` | AssetHub | Application name |
| `app.port` | `APP_PORT` | 8080 | HTTP port |
| `app.env` | `APP_ENV` | development | Environment |
| `database.host` | `DB_HOST` | localhost | PostgreSQL host |
| `database.port` | `DB_PORT` | 5432 | PostgreSQL port |
| `database.user` | `DB_USER` | postgres | Database user |
| `database.password` | `DB_PASSWORD` | - | Database password |
| `database.dbname` | `DB_NAME` | assethub | Database name |
| `redis.host` | `REDIS_HOST` | localhost | Redis host |
| `redis.port` | `REDIS_PORT` | 6379 | Redis port |
| `redis.db` | `REDIS_DB` | 2 | Redis database |
| `storage.type` | `STORAGE_TYPE` | oss | Storage backend (s3/oss/local) |

## Connection Strings

| Service | Connection String |
|---------|------------------|
| PostgreSQL | `postgresql://postgres:postgres@localhost:5432/assethub` |
| Redis | `redis://localhost:6379/2` |

## License

MIT
