# 测试指南

## 快速开始

```bash
# 运行所有测试
make test

# 或直接使用 go test
go test -v ./internal/handlers -timeout 60s
```

## 测试类型

### 1. 集成测试（Integration Tests）

**文件**: `internal/handlers/file_handler_integration_test.go`

**特点**:
- 使用 Mock Storage（无需真实 S3）
- 需要 PostgreSQL 数据库
- 测试完整的 HTTP 请求/响应流程

**运行**:
```bash
go test -v ./internal/handlers
```

### 2. 单元测试（Unit Tests）

**文件**: `pkg/storage/s3_test.go`

**特点**:
- 测试 S3 Storage 实现
- 需要真实 S3 或 MinIO

**运行**:
```bash
go test -v ./pkg/storage
```

## 测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

## 环境配置

### 必需服务

1. **PostgreSQL**
   ```bash
   # 使用 Docker 启动
   docker run -d \
     --name postgres-test \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=assethub \
     -p 5432:5432 \
     postgres:15
   ```

2. **配置文件**
   - 确保 `configs/config.yaml` 存在
   - 数据库配置正确

### 可选服务（真实 S3 测试）

如需测试真实 S3 连接，配置环境变量：

```bash
export S3_REGION=us-east-1
export S3_BUCKET=your-bucket
export S3_ACCESS_KEY_ID=your-key
export S3_SECRET_ACCESS_KEY=your-secret
```

## 测试命令

### 运行特定测试

```bash
# 运行单个测试套件
go test -v ./internal/handlers -run TestUploadDirectWithMock

# 运行匹配模式的测试
go test -v ./internal/handlers -run ".*Mock"
```

### 并行测试

```bash
# 使用 4 个并行进程
go test -v -parallel 4 ./...
```

### 详细输出

```bash
# 显示所有日志
go test -v ./internal/handlers -args -test.v

# 显示测试覆盖的代码行
go test -v -cover ./...
```

## 测试数据清理

测试自动清理数据：
- 每个测试套件执行后删除 `test_*` 开头的文件记录
- Mock Storage 在内存中，测试结束自动释放

手动清理数据库：
```sql
DELETE FROM files WHERE name LIKE 'test_%';
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: assethub
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## 常见问题

### 1. 数据库连接失败

**错误**:
```
Failed to connect to database
```

**解决**:
- 检查 PostgreSQL 是否运行
- 验证 `configs/config.yaml` 中的数据库配置
- 确保数据库 `assethub` 已创建

### 2. 约束警告

**警告**:
```
ERROR: constraint "uni_files_storage_key" of relation "files" does not exist
```

**说明**:
- GORM 尝试删除不存在的约束
- 不影响测试执行
- 已在代码中忽略此错误

### 3. 测试超时

**错误**:
```
test timed out after 30s
```

**解决**:
```bash
# 增加超时时间
go test -v ./internal/handlers -timeout 60s
```

## 测试最佳实践

1. **使用 Mock 避免外部依赖**
   - 集成测试使用 Mock Storage
   - 单元测试使用 Mock 接口

2. **清理测试数据**
   - 使用 `defer cleanup()`
   - 测试文件名使用 `test_` 前缀

3. **独立测试**
   - 每个测试独立运行
   - 不依赖其他测试的状态

4. **错误场景覆盖**
   - 测试正常流程
   - 测试错误处理

## 参考文档

- [E2E 测试详细文档](./E2E_TESTING.md)
- [S3 配置文档](./S3_CONFIGURATION.md)
- [API 文档](./swagger.yaml)
