# Feature: 添加阿里云 OSS 存储后端

## Summary

为 AssetHub 添加阿里云 OSS 作为可选存储后端，实现 Storage 接口的 8 个方法，支持小文件上传、大文件分片上传、预签名 URL 生成和文件删除。

## User Story

As a 系统管理员
I want to 选择使用阿里云 OSS 作为文件存储后端
So that 可以在中国区域获得更好的性能和成本优势

## Problem Statement

当前项目只支持 S3 存储后端，在中国区域使用时存在网络延迟和成本问题。需要添加阿里云 OSS 支持，作为可选的存储后端。

## Solution Statement

在现有的 Storage 接口抽象基础上，添加 OSS 实现。使用阿里云 OSS Go SDK V2，实现与 S3Storage 相同的 8 个接口方法。通过配置文件的 `storage.type` 字段切换存储后端。

## Metadata

| Field            | Value                                             |
| ---------------- | ------------------------------------------------- |
| Type             | NEW_CAPABILITY                                    |
| Complexity       | MEDIUM                                            |
| Systems Affected | 存储层、配置管理                                   |
| Dependencies     | github.com/aliyun/alibabacloud-oss-go-sdk-v2      |
| Estimated Tasks  | 5                                                 |

---

## UX Design

### Before State
```
配置文件:
  storage:
    type: "s3"  # 只能选择 s3 或 local（未实现）

系统行为:
  - 只支持 AWS S3 或 S3 兼容服务（MinIO）
  - 中国区域访问 S3 延迟高、成本高
```

### After State
```
配置文件:
  storage:
    type: "oss"  # 新增选项
    oss:
      endpoint: "oss-cn-hangzhou.aliyuncs.com"
      bucket: "assethub-files"
      access_key_id: "xxx"
      access_key_secret: "xxx"

系统行为:
  - 支持阿里云 OSS 作为存储后端
  - 中国区域用户获得更好的性能和成本
  - 所有文件操作（上传、下载、删除）透明切换
```

### Interaction Changes
| Location | Before | After | User Impact |
|----------|--------|-------|-------------|
| 配置文件 | 只能选 s3/local | 可选 oss | 支持阿里云 OSS |
| 文件上传 API | 只能传到 S3 | 可传到 OSS | 中国区域更快 |
| 预签名 URL | S3 URL | OSS URL | 前端直传到 OSS |

---

## Mandatory Reading

**CRITICAL: 实现前必须读这些文件:**

| Priority | File | Lines | Why Read This |
|----------|------|-------|---------------|
| P0 | `pkg/storage/s3.go` | 1-199 | 完全镜像这个结构和模式 |
| P0 | `pkg/storage/storage.go` | 18-94 | Storage 接口定义 |
| P1 | `internal/config/config.go` | 45-62 | 配置结构模式 |
| P1 | `configs/config.example.yaml` | 30-40 | 配置文件格式 |

**External Documentation:**
| Source | Section | Why Needed |
|--------|---------|------------|
| [阿里云 OSS Go SDK V2](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2) | README | SDK 安装和初始化 |
| [OSS Go SDK V2 - 简单上传](https://www.alibabacloud.com/help/en/oss/developer-reference/v2-simple-upload) | PutObject | 实现 Upload 方法 |
| [OSS Go SDK V2 - 分片上传](https://www.alibabacloud.com/help/en/oss/developer-reference/v2-multipart-upload) | Multipart | 实现分片上传 |
| [OSS Go SDK V2 - 预签名 URL](https://www.alibabacloud.com/help/en/oss/developer-reference/v2-presign-upload) | Presign | 实现预签名方法 |

---

## Patterns to Mirror

**文件结构（完全镜像 s3.go）:**
```go
// SOURCE: pkg/storage/s3.go:16-30
// COPY THIS PATTERN:

// OSSConfig OSS 配置
type OSSConfig struct {
    Endpoint        string // OSS endpoint (如 oss-cn-hangzhou.aliyuncs.com)
    Bucket          string // OSS bucket 名称
    AccessKeyID     string // 阿里云 Access Key ID
    AccessKeySecret string // 阿里云 Access Key Secret
}

// OSSStorage OSS 存储实现
type OSSStorage struct {
    client *oss.Client  // 替换为 OSS 客户端
    bucket string
}
```

**构造函数模式:**
```go
// SOURCE: pkg/storage/s3.go:32-75
// COPY THIS PATTERN:

func NewOSSStorage(ctx context.Context, cfg OSSConfig) (*OSSStorage, error) {
    // 1. 创建 OSS 客户端
    // 2. 验证配置
    // 3. 返回 OSSStorage 实例
}
```

**接口实现模式（8 个方法）:**
```go
// SOURCE: pkg/storage/s3.go:77-198
// COPY THIS PATTERN:

// 每个方法的结构：
// 1. 调用 OSS SDK 对应方法
// 2. 错误处理：fmt.Errorf("failed to xxx: %w", err)
// 3. 返回结果
```

**错误处理模式:**
```go
// SOURCE: pkg/storage/s3.go:86-87
// COPY THIS PATTERN:

if err != nil {
    return fmt.Errorf("failed to upload object: %w", err)
}
```

---

## Files to Change

| File | Action | Justification |
|------|--------|---------------|
| `pkg/storage/oss.go` | CREATE | OSS 存储实现（镜像 s3.go） |
| `pkg/storage/storage.go` | UPDATE | 工厂函数添加 "oss" case |
| `internal/config/config.go` | UPDATE | 添加 OSSConfig 结构体和环境变量绑定 |
| `configs/config.example.yaml` | UPDATE | 添加 OSS 配置示例 |
| `pkg/storage/oss_test.go` | CREATE | 单元测试（镜像 s3_test.go） |

---

## NOT Building (Scope Limits)

- **不实现本地存储**：这是另一个独立任务
- **不修改业务逻辑**：Service 层和 Handler 层无需改动
- **不迁移现有数据**：从 S3 迁移到 OSS 是运维任务，不在此范围
- **不实现 OSS 特有功能**：只实现 Storage 接口定义的 8 个方法

---

## Step-by-Step Tasks

### Task 1: 安装 OSS SDK 依赖

**ACTION**: 添加阿里云 OSS Go SDK V2 依赖

**IMPLEMENT**:
```bash
go get github.com/aliyun/alibabacloud-oss-go-sdk-v2
```

**VALIDATE**:
```bash
go mod tidy && go mod verify
```

**EXPECT**: go.mod 中出现 `github.com/aliyun/alibabacloud-oss-go-sdk-v2` 依赖

---

### Task 2: CREATE `pkg/storage/oss.go`

**ACTION**: 创建 OSS 存储实现文件

**IMPLEMENT**:
1. 定义 `OSSConfig` 结构体（4 个字段：Endpoint, Bucket, AccessKeyID, AccessKeySecret）
2. 定义 `OSSStorage` 结构体（2 个字段：client, bucket）
3. 实现 `NewOSSStorage(ctx context.Context, cfg OSSConfig) (*OSSStorage, error)`
4. 实现 8 个接口方法：
   - `Upload(ctx, key, reader, size) error`
   - `GeneratePresignedUploadURL(ctx, key, expiry) (string, error)`
   - `InitMultipartUpload(ctx, key) (*MultipartUpload, error)`
   - `GeneratePresignedPartURL(ctx, key, uploadID, partNumber, expiry) (string, error)`
   - `CompleteMultipartUpload(ctx, key, uploadID, parts) error`
   - `GeneratePresignedDownloadURL(ctx, key, expiry) (string, error)`
   - `Delete(ctx, key) error`

**MIRROR**: `pkg/storage/s3.go:1-199` - 完全镜像文件结构

**IMPORTS**:
```go
import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
    "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)
```

**GOTCHA**:
- OSS SDK 的客户端初始化方式与 S3 不同，需要使用 `oss.NewClient()`
- 预签名 URL 生成使用 `client.Presign()` 方法
- 分片上传的 UploadID 类型可能不同，需要适配

**VALIDATE**:
```bash
go build ./pkg/storage/...
```

**EXPECT**: 编译通过，无错误

---

### Task 3: UPDATE `pkg/storage/storage.go`

**ACTION**: 在工厂函数中添加 OSS case

**IMPLEMENT**: 在 `NewStorage()` 函数的 switch 语句中添加：
```go
case "oss":
    ossConfig := OSSConfig{
        Endpoint:        cfg.OSS.Endpoint,
        Bucket:          cfg.OSS.Bucket,
        AccessKeyID:     cfg.OSS.AccessKeyID,
        AccessKeySecret: cfg.OSS.AccessKeySecret,
    }
    return NewOSSStorage(ctx, ossConfig)
```

**MIRROR**: `pkg/storage/storage.go:100-109` - 镜像 S3 case 的结构

**LOCATION**: 在第 110 行（"local" case）之前插入

**VALIDATE**:
```bash
go build ./pkg/storage/...
```

**EXPECT**: 编译通过

---

### Task 4: UPDATE `internal/config/config.go`

**ACTION**: 添加 OSS 配置结构体和环境变量绑定

**IMPLEMENT**:

**4.1 添加 OSSConfig 结构体**（在第 58 行后插入）:
```go
type OSSConfig struct {
    Endpoint        string `mapstructure:"endpoint"`
    Bucket          string `mapstructure:"bucket"`
    AccessKeyID     string `mapstructure:"access_key_id"`
    AccessKeySecret string `mapstructure:"access_key_secret"`
}
```

**4.2 更新 StorageConfig 结构体**（修改第 45-49 行）:
```go
type StorageConfig struct {
    Type  string      `mapstructure:"type"`
    S3    S3Config    `mapstructure:"s3"`
    OSS   OSSConfig   `mapstructure:"oss"`  // 新增
    Local LocalConfig `mapstructure:"local"`
}
```

**4.3 添加环境变量绑定**（在第 100 行后插入）:
```go
viper.BindEnv("storage.oss.endpoint", "OSS_ENDPOINT")
viper.BindEnv("storage.oss.bucket", "OSS_BUCKET")
viper.BindEnv("storage.oss.access_key_id", "OSS_ACCESS_KEY_ID")
viper.BindEnv("storage.oss.access_key_secret", "OSS_ACCESS_KEY_SECRET")
```

**MIRROR**: `internal/config/config.go:51-58` (S3Config) 和 `94-100` (环境变量绑定)

**VALIDATE**:
```bash
go build ./internal/config/...
```

**EXPECT**: 编译通过

---

### Task 5: UPDATE `configs/config.example.yaml`

**ACTION**: 添加 OSS 配置示例

**IMPLEMENT**: 在第 38 行后插入：
```yaml
  oss:
    endpoint: "oss-cn-hangzhou.aliyuncs.com"  # OSS endpoint (根据区域选择)
    bucket: "assethub-files"                  # OSS bucket name
    access_key_id: ""                         # Aliyun Access Key ID
    access_key_secret: ""                     # Aliyun Access Key Secret
```

**MIRROR**: `configs/config.example.yaml:32-38` (S3 配置格式)

**VALIDATE**:
```bash
# 验证 YAML 语法
cat configs/config.example.yaml | grep -A 5 "oss:"
```

**EXPECT**: 输出 OSS 配置块，格式正确

---

### Task 6: CREATE `pkg/storage/oss_test.go`

**ACTION**: 创建 OSS 单元测试

**IMPLEMENT**:
1. 创建 `MockOSSStorage` 结构体（实现 Storage 接口）
2. 编写测试用例：
   - `TestOSSUpload`
   - `TestOSSGeneratePresignedUploadURL`
   - `TestOSSMultipartUpload`
   - `TestOSSGeneratePresignedDownloadURL`
   - `TestOSSDelete`

**MIRROR**: `pkg/storage/s3_test.go:1-200` - 完全镜像测试结构

**VALIDATE**:
```bash
go test ./pkg/storage/... -v
```

**EXPECT**: 所有测试通过

---

## Testing Strategy

### Unit Tests to Write

| Test File | Test Cases | Validates |
|-----------|-----------|-----------|
| `pkg/storage/oss_test.go` | Upload, PresignedURL, Multipart, Download, Delete | OSS 接口实现 |

### Edge Cases Checklist

- [ ] 空 AccessKeyID/AccessKeySecret（应该返回错误）
- [ ] 无效的 Endpoint（应该返回错误）
- [ ] 不存在的 Bucket（应该返回错误）
- [ ] 空文件上传（size = 0）
- [ ] 大文件分片上传（测试分片逻辑）
- [ ] 预签名 URL 过期时间（测试不同的 expiry 值）

---

## Validation Commands

### Level 1: STATIC_ANALYSIS
```bash
go build ./... && go vet ./...
```
**EXPECT**: 编译通过，无 vet 警告

### Level 2: UNIT_TESTS
```bash
go test ./pkg/storage/... -v -cover
```
**EXPECT**: 所有测试通过，覆盖率 >= 80%

### Level 3: FULL_SUITE
```bash
go test ./... -v && go build ./cmd/api
```
**EXPECT**: 所有测试通过，API 编译成功

### Level 4: MANUAL_VALIDATION

1. **配置 OSS**:
   - 复制 `config.example.yaml` 到 `config.yaml`
   - 设置 `storage.type: "oss"`
   - 填写真实的 OSS 配置

2. **启动服务**:
   ```bash
   go run cmd/api/main.go
   ```

3. **测试文件上传**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/files/upload \
     -F "file=@test.jpg"
   ```

4. **验证文件存储**:
   - 登录阿里云 OSS 控制台
   - 检查文件是否成功上传到指定 Bucket

---

## Acceptance Criteria

- [ ] OSS SDK 依赖安装成功
- [ ] `pkg/storage/oss.go` 实现所有 8 个接口方法
- [ ] 工厂函数支持 `type: "oss"` 配置
- [ ] 配置文件支持 OSS 配置项
- [ ] 单元测试覆盖率 >= 80%
- [ ] 所有验证命令通过
- [ ] 手动测试文件上传到 OSS 成功
- [ ] 代码风格与 `s3.go` 保持一致

---

## Completion Checklist

- [ ] Task 1: OSS SDK 依赖安装
- [ ] Task 2: `pkg/storage/oss.go` 创建完成
- [ ] Task 3: `pkg/storage/storage.go` 工厂函数更新
- [ ] Task 4: `internal/config/config.go` 配置结构更新
- [ ] Task 5: `configs/config.example.yaml` 配置示例更新
- [ ] Task 6: `pkg/storage/oss_test.go` 测试创建完成
- [ ] Level 1: 静态分析通过
- [ ] Level 2: 单元测试通过
- [ ] Level 3: 完整测试套件通过
- [ ] Level 4: 手动验证通过

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| OSS SDK API 与 S3 差异大 | MEDIUM | HIGH | 仔细阅读 OSS 官方文档，必要时调整接口实现 |
| 预签名 URL 生成方式不同 | MEDIUM | MEDIUM | 参考 OSS SDK 示例代码，测试验证 |
| 分片上传 UploadID 类型不兼容 | LOW | MEDIUM | 使用 string 类型统一处理 |
| 配置项遗漏 | LOW | LOW | 对照 S3Config 逐项检查 |

---

## Notes

**设计决策**:
- 完全镜像 `s3.go` 的结构，保持代码一致性
- 使用阿里云 OSS Go SDK V2（最新版本）
- 配置项命名与 S3 保持对称（endpoint, bucket, access_key_id, access_key_secret）

**未来考虑**:
- 支持 OSS 的 STS 临时凭证
- 支持 OSS 的图片处理功能（缩略图、水印等）
- 支持 OSS 的 CDN 加速

**参考资料**:
- [阿里云 OSS Go SDK V2 GitHub](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2)
- [阿里云 OSS 官方文档](https://www.alibabacloud.com/help/en/oss/developer-reference/quick-start-for-oss-go-sdk-v2)
- [使用 Go SDK V2 生成预签名 URL](https://www.alibabacloud.com/help/en/oss/developer-reference/v2-presign-upload)
- [使用 Go SDK V2 分片上传](https://www.alibabacloud.com/help/en/oss/developer-reference/v2-multipart-upload)
