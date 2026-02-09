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
| Database | PostgreSQL 18+ (Docker: postgres:18-alpine) |
| Cache | Redis 7+ (go-redis v9) |
| Storage | AWS S3 SDK v2 / Aliyun OSS SDK |
| Config | Viper v1.21.0 |
| Logger | Zap |
| Docs | Swagger (swaggo) |

## Requirements

- Go 1.25+
- PostgreSQL 18+ (or 14+ for older versions)
- Redis 6+
- AWS S3 / Aliyun OSS / MinIO (external cluster)

## Requirements

- Go 1.24+
- PostgreSQL 14+
- Redis 6+
- AWS S3 / Aliyun OSS / MinIO (optional, can use local storage)

## Quick Start with Docker (Recommended)

The fastest way to start AssetHub is using Docker Compose, which includes all dependencies:

### 1. Start All Services

```bash
# Start PostgreSQL, Redis, MinIO, and AssetHub
docker-compose up -d

# View logs
docker-compose logs -f api
```

This will start:
- **PostgreSQL 18** at `localhost:5432` (latest stable version)
- **Redis 7** at `localhost:6379`
- **AssetHub API** at `localhost:8080`

**Note**: Storage backend (AWS S3 or external MinIO) must be configured separately.

### 2. Verify Services

```bash
# Health check
curl http://localhost:8080/health

# Swagger UI
open http://localhost:8080/swagger/index.html
```

### ⚠️ PostgreSQL 18 重要变更

**Docker 卷挂载路径已更改**：

PostgreSQL 18 的 Docker 镜像改变了数据目录结构：
- **旧路径（17 及更早）**: `/var/lib/postgresql/data`
- **新路径（18+）**: `/var/lib/postgresql`

**原因**：PostgreSQL 18 内部使用版本特定的 PGDATA 路径（`/var/lib/postgresql/18/docker`），而 `data` 目录现在是符号链接。

**影响**：如果继续使用旧路径会导致容器启动失败。docker-compose.yaml 已正确配置为 `/var/lib/postgresql`。

详情参考：[Upgrading to PostgreSQL 18 in Docker](https://medepros.ca/blog/upgrading-to-postgresql-18-in-docker-a-crucial-change-you-need-to-be-aware-of/)

### 3. Configure Storage Backend

Before starting services, configure your S3 storage:

**Option A: AWS S3 (Recommended for production)**

```bash
# Create .env.docker file
cp .env.docker.example .env.docker

# Edit with your AWS credentials
vim .env.docker
```

```bash
# .env.docker
S3_REGION=ap-southeast-1
S3_BUCKET=your-production-bucket
S3_ACCESS_KEY_ID=your-access-key-id
S3_SECRET_ACCESS_KEY=your-secret-access-key
S3_ENDPOINT=  # Leave empty for AWS S3
S3_USE_PATH_STYLE=false
```

**Option B: External MinIO Cluster (Development)**

```bash
# .env.docker
S3_REGION=us-east-1
S3_BUCKET=assethub-files
S3_ACCESS_KEY_ID=your-minio-access-key
S3_SECRET_ACCESS_KEY=your-minio-secret-key
S3_ENDPOINT=http://localhost:9000  # Your MinIO endpoint
S3_USE_PATH_STYLE=true
```

### 4. Start Services

```bash
# Start with custom environment
docker-compose --env-file .env.docker up -d

# Or use default (requires S3 credentials configured)
docker-compose up -d
```

### 5. Verify Services

```bash
# Health check
curl http://localhost:8080/health

# Swagger UI
open http://localhost:8080/swagger/index.html
```

### 6. Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ deletes all data)
docker-compose down -v
```

### ⚠️ Security Warning

**This configuration is for development only!**

The docker-compose.yaml uses hardcoded default credentials:
- PostgreSQL: `postgres/postgres`

**For production deployment:**
1. Use environment variable files (`.env.docker`) to manage S3 credentials
2. Change PostgreSQL default password
3. Use Docker secrets or Kubernetes secrets for sensitive data
4. Restrict port bindings to `127.0.0.1` where possible
5. Enable SSL/TLS for database connections (`DB_SSLMODE=require`)
6. Use strong AWS IAM policies instead of access keys when possible
7. Rotate S3 access keys regularly

Example production environment:
```bash
# .env.production
POSTGRES_PASSWORD=your_strong_random_password_here
S3_ACCESS_KEY_ID=your-access-key-id
S3_SECRET_ACCESS_KEY=your-secret-access-key
DB_SSLMODE=require
APP_ENV=production
LOG_LEVEL=info
```

### Docker Compose Services

| Service | Description | Ports |
|---------|-------------|-------|
| `api` | AssetHub application | 8080 |
| `postgres` | PostgreSQL database | 5432 |
| `redis` | Redis cache | 6379 |

**Note**: Storage backend (S3/OSS) must be configured externally via environment variables.

### Customization

Copy `.env.docker.example` to `.env.docker` and modify:

```bash
cp .env.docker.example .env.docker
# Edit .env.docker with your settings
docker-compose --env-file .env.docker up -d
```

**Key Environment Variables**:

- `STORAGE_TYPE`: Storage backend (`s3`, `oss`, `local`)
- `S3_REGION`: AWS S3 region (e.g., `us-east-1`, `ap-southeast-1`)
- `S3_BUCKET`: S3 bucket name
- `S3_ACCESS_KEY_ID`: AWS access key ID
- `S3_SECRET_ACCESS_KEY`: AWS secret access key
- `S3_ENDPOINT`: S3 endpoint (leave empty for AWS S3, set to MinIO URL for local cluster)
- `S3_USE_PATH_STYLE`: Set to `true` for MinIO, `false` for AWS S3
- `LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`)

### Production Deployment

For production, use the Dockerfile with your orchestration platform:

```bash
# Build image
docker build -t assethub:latest .

# Run with external database/storage
docker run -d \
  --name assethub \
  -p 8080:8080 \
  -e DB_HOST=your-postgres-host \
  -e REDIS_HOST=your-redis-host \
  -e STORAGE_TYPE=s3 \
  -e S3_BUCKET=your-production-bucket \
  assethub:latest
```

See [Configuration](#configuration) for all environment variables.

## Quick Start (Local Development)

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

## Upload & Download Workflows

### 1. Presigned Upload (Frontend Direct Upload)

```mermaid
sequenceDiagram
    participant Client
    participant AssetHub
    participant S3/OSS

    Client->>AssetHub: POST /files/upload/presigned/init<br/>{name, content_type, size}
    AssetHub->>AssetHub: Create file record (status=pending)
    AssetHub->>S3/OSS: Generate presigned PUT URL
    S3/OSS-->>AssetHub: Presigned URL (expires in 1h)
    AssetHub-->>Client: {file_id, upload_url, storage_key}

    Client->>S3/OSS: PUT file to presigned URL<br/>(direct upload, no proxy)
    S3/OSS-->>Client: 200 OK

    Client->>AssetHub: POST /files/upload/presigned/confirm<br/>{file_id}
    AssetHub->>AssetHub: Update file status to completed
    AssetHub-->>Client: {file_id, status=completed}
```

**Use Case**: Small to medium files (< 100MB), reduces backend bandwidth.

### 2. Multipart Upload (Large Files)

```mermaid
sequenceDiagram
    participant Client
    participant AssetHub
    participant S3/OSS

    Client->>AssetHub: POST /files/upload/multipart/init<br/>{name, content_type, size}
    AssetHub->>AssetHub: Create file record (status=uploading)
    AssetHub->>S3/OSS: Initiate multipart upload
    S3/OSS-->>AssetHub: upload_id
    AssetHub-->>Client: {file_id, upload_id, storage_key}

    loop For each part (1 to N)
        Client->>AssetHub: POST /files/upload/multipart/part-url<br/>{file_id, part_number}
        AssetHub->>S3/OSS: Generate presigned URL for part
        S3/OSS-->>AssetHub: Presigned URL
        AssetHub-->>Client: {part_number, upload_url}

        Client->>S3/OSS: PUT part to presigned URL
        S3/OSS-->>Client: 200 OK + ETag
    end

    Client->>AssetHub: POST /files/upload/multipart/complete<br/>{file_id, parts: [{part_number, etag}]}
    AssetHub->>S3/OSS: Complete multipart upload
    S3/OSS-->>AssetHub: 200 OK
    AssetHub->>AssetHub: Update file status to completed
    AssetHub-->>Client: {file_id, status=completed}
```

**Use Case**: Large files (>= 100MB), supports resume, parallel upload.

### 3. Presigned Download

```mermaid
sequenceDiagram
    participant Client
    participant AssetHub
    participant S3/OSS

    Client->>AssetHub: GET /files/:id/download-url
    AssetHub->>AssetHub: Query file metadata
    AssetHub->>S3/OSS: Generate presigned GET URL
    S3/OSS-->>AssetHub: Presigned URL (expires in 15min)
    AssetHub-->>Client: {file_id, download_url, expires_in}

    Client->>S3/OSS: GET file from presigned URL<br/>(direct download, no proxy)
    S3/OSS-->>Client: File content
```

**Use Case**: Secure file access with expiration, reduces backend bandwidth.

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
