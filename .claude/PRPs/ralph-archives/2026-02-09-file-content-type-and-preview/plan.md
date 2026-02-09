# Feature: 自动推导 Content-Type 并支持浏览器预览

## 摘要

文件上传时自动检测 Content-Type，生成的预签名 URL 能在浏览器中直接预览（图片、视频、PDF等），而不是触发下载。

---

## User Story

**作为** 前端开发者  
**我想要** 上传文件时自动推导正确的 Content-Type  
**以便** 生成的预签名 URL 能在浏览器中直接预览，而不是下载

**验收标准**:
- 上传图片/视频时，即使客户端未提供 Content-Type，系统也能自动检测
- 预签名下载 URL 在浏览器中打开时，图片/视频/PDF 直接预览
- 不可预览的文件（如 .zip）触发下载行为
- 客户端提供的 Content-Type 会被验证，防止伪造

---

## Problem Statement

**当前问题**:
1. Content-Type 完全依赖客户端输入，无验证（`internal/handlers/file_handler.go:154-206`）
2. 预签名 URL 未设置响应头，浏览器默认下载而非预览（`pkg/storage/oss.go:144-155`, `pkg/storage/s3.go:170-185`）
3. 客户端可以伪造 Content-Type，导致安全风险

**影响**:
- 用户体验差：图片/视频无法直接预览
- 安全风险：恶意文件可以伪装成图片
- 数据不一致：数据库中的 Content-Type 不可信

## Solution Statement

**解决方案**:
1. 在 `UploadDirect` 中使用 `http.DetectContentType()` 检测文件头（前 512 字节）
2. 修改 `Storage` 接口，添加 `PresignOptions` 参数支持响应头设置
3. 在 OSS/S3 实现中设置 `ResponseContentDisposition` 和 `ResponseContentType`
4. 根据 MIME 类型判断使用 `inline`（预览）还是 `attachment`（下载）

**技术选型**:
- 使用 Go 标准库 `http.DetectContentType()`（已足够，无需引入 `mimetype` 库）
- OSS SDK: `oss.GetObjectRequest.ResponseContentType`
- S3 SDK: `s3.GetObjectInput.ResponseContentType`

---

## Metadata

```yaml
type: ENHANCEMENT
complexity: MEDIUM
estimated_time: 4-6 hours
affected_components:
  - pkg/storage (interface + implementations)
  - internal/services/file_service.go
  - internal/handlers/file_handler.go
dependencies:
  - github.com/aliyun/alibabacloud-oss-go-sdk-v2 v1.4.0
  - github.com/aws/aws-sdk-go-v2/service/s3 v1.71.1
```

---

## UX Design

### Before (当前行为)

```
用户上传 image.png
  ↓
客户端发送: Content-Type: "application/octet-stream" (错误)
  ↓
服务端直接存储: file.content_type = "application/octet-stream"
  ↓
生成预签名 URL (无响应头设置)
  ↓
浏览器打开 URL → 触发下载 ❌
```

### After (期望行为)

```
用户上传 image.png
  ↓
客户端发送: Content-Type: "application/octet-stream" (错误)
  ↓
服务端检测文件头: 前 512 字节 → "image/png" ✅
  ↓
存储到数据库: file.content_type = "image/png"
  ↓
生成预签名 URL:
  - ResponseContentType: "image/png"
  - ResponseContentDisposition: "inline" (因为是图片)
  ↓
浏览器打开 URL → 直接预览 ✅
```

---

## Mandatory Reading

**必读文件**（按顺序）:

1. **`pkg/storage/storage.go`** (line 1-50)
   - 理解 `Storage` 接口定义
   - 当前 `GeneratePresignedDownloadURL` 签名

2. **`pkg/storage/oss.go`** (line 144-155)
   - OSS 预签名实现
   - 需要添加 `ResponseContentType` 和 `ResponseContentDisposition`

3. **`pkg/storage/s3.go`** (line 170-185)
   - S3 预签名实现
   - 需要添加响应头参数

4. **`internal/services/file_service.go`** (line 88-136, 139-169)
   - `UploadDirect` 和 `InitPresignedUpload` 实现
   - 需要添加 Content-Type 检测逻辑

5. **`internal/handlers/file_handler.go`** (line 154-206, 219-245)
   - Handler 层请求处理
   - 需要验证客户端提供的 Content-Type

6. **`internal/models/file.go`** (line 18)
   - `ContentType` 字段定义（varchar(100)）

---

## Patterns to Mirror

### Pattern 1: 接口扩展（添加可选参数）

**参考**: `pkg/storage/storage.go`

当前接口:
```go
GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration) (string, error)
```

**扩展模式**（添加 options 参数）:
```go
// PresignOptions 预签名选项
type PresignOptions struct {
    ContentType        string // 响应 Content-Type
    ContentDisposition string // 响应 Content-Disposition (inline/attachment)
}

// 修改后的接口
GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration, opts *PresignOptions) (string, error)
```

**Linus 点评**: 好。用结构体封装可选参数，而不是无限增加函数参数。这是 Go 的惯用法。

---

### Pattern 2: Content-Type 检测

**使用 Go 标准库**:
```go
import "net/http"

// 读取文件前 512 字节
buffer := make([]byte, 512)
n, _ := file.Read(buffer)

// 自动检测
detectedType := http.DetectContentType(buffer[:n])
// 返回: "image/png", "video/mp4", "application/pdf" 等
```

**Linus 点评**: 完美。用标准库，不要引入第三方依赖。512 字节足够检测 99% 的文件类型。

---

### Pattern 3: 可预览类型判断

```go
import "strings"

func isPreviewable(contentType string) bool {
    previewable := []string{
        "image/",
        "video/",
        "audio/",
        "application/pdf",
        "text/",
    }
    
    for _, prefix := range previewable {
        if strings.HasPrefix(contentType, prefix) {
            return true
        }
    }
    return false
}

func getContentDisposition(contentType, filename string) string {
    if isPreviewable(contentType) {
        return "inline"
    }
    return fmt.Sprintf("attachment; filename=%q", filename)
}
```

**Linus 点评**: 简洁。但是 `filename` 需要 URL 编码，防止注入。用 `mime.FormatMediaType()`。

---

### Pattern 4: OSS SDK 响应头设置

**参考**: 阿里云文档 + `pkg/storage/oss.go:144-155`

当前实现:
```go
result, err := s.presignClient.PresignGetObject(ctx, &oss.GetObjectRequest{
    Bucket: oss.Ptr(s.bucket),
    Key:    oss.Ptr(key),
}, oss.PresignExpires(expiration))
```

**修改后**:
```go
req := &oss.GetObjectRequest{
    Bucket: oss.Ptr(s.bucket),
    Key:    oss.Ptr(key),
}

// 设置响应头（如果提供）
if opts != nil {
    if opts.ContentType != "" {
        req.ResponseContentType = oss.Ptr(opts.ContentType)
    }
    if opts.ContentDisposition != "" {
        req.ResponseContentDisposition = oss.Ptr(opts.ContentDisposition)
    }
}

result, err := s.presignClient.PresignGetObject(ctx, req, oss.PresignExpires(expiration))
```

---

### Pattern 5: S3 SDK 响应头设置

**参考**: IBM COS 文档 + `pkg/storage/s3.go:170-185`

当前实现:
```go
req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
    Bucket: aws.String(s.bucket),
    Key:    aws.String(key),
}, s3.WithPresignExpires(expiration))
```

**修改后**:
```go
input := &s3.GetObjectInput{
    Bucket: aws.String(s.bucket),
    Key:    aws.String(key),
}

// 设置响应头（如果提供）
if opts != nil {
    if opts.ContentType != "" {
        input.ResponseContentType = aws.String(opts.ContentType)
    }
    if opts.ContentDisposition != "" {
        input.ResponseContentDisposition = aws.String(opts.ContentDisposition)
    }
}

req, err := s.presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expiration))
```

---

## Files to Change

### 核心文件（必改）

1. **`pkg/storage/storage.go`**
   - 添加 `PresignOptions` 结构体
   - 修改 `GeneratePresignedDownloadURL` 接口签名

2. **`pkg/storage/oss.go`**
   - 实现 OSS 响应头设置
   - 修改 `GeneratePresignedDownloadURL` 方法

3. **`pkg/storage/s3.go`**
   - 实现 S3 响应头设置
   - 修改 `GeneratePresignedDownloadURL` 方法

4. **`internal/services/file_service.go`**
   - 在 `UploadDirect` 中添加 Content-Type 检测
   - 修改 `GetPresignedDownloadURL` 调用，传递响应头参数

5. **`internal/handlers/file_handler.go`**
   - 验证客户端提供的 Content-Type（可选）
   - 更新 Swagger 注释

### 辅助文件（可选）

6. **`pkg/utils/mime.go`** (新建)
   - 封装 Content-Type 检测逻辑
   - 封装可预览类型判断逻辑

7. **`pkg/storage/oss_test.go`** (新建或修改)
   - 测试 OSS 响应头设置

8. **`pkg/storage/s3_test.go`** (新建或修改)
   - 测试 S3 响应头设置

---

## Step-by-Step Tasks

### Task 1: 修改 Storage 接口，添加响应头参数支持

**ACTION**: 扩展 `Storage` 接口，添加 `PresignOptions` 结构体

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/pkg/storage/storage.go`

```go
// 在 Storage 接口定义之前添加
type PresignOptions struct {
    ContentType        string // 响应 Content-Type
    ContentDisposition string // 响应 Content-Disposition (inline/attachment)
}

// 修改接口方法签名
type Storage interface {
    // ... 其他方法保持不变
    
    // GeneratePresignedDownloadURL 生成预签名下载 URL
    // opts 为 nil 时使用默认行为（不设置响应头）
    GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration, opts *PresignOptions) (string, error)
}
```

**MIRROR**: Pattern 1 - 接口扩展模式

**IMPORTS**: 无新增

**GOTCHA**:
- `opts *PresignOptions` 使用指针，允许传 `nil`（向后兼容）
- 不要使用 `...PresignOptions` 可变参数，那是垃圾设计

**VALIDATE**:
```bash
go build ./pkg/storage/...
```

---

### Task 2: 实现 OSS 存储的响应头设置

**ACTION**: 修改 `pkg/storage/oss.go` 的 `GeneratePresignedDownloadURL` 方法

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/pkg/storage/oss.go`

定位到 line 144-155，替换整个方法:

```go
func (s *OSSStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration, opts *PresignOptions) (string, error) {
    req := &oss.GetObjectRequest{
        Bucket: oss.Ptr(s.bucket),
        Key:    oss.Ptr(key),
    }

    // 设置响应头（如果提供）
    if opts != nil {
        if opts.ContentType != "" {
            req.ResponseContentType = oss.Ptr(opts.ContentType)
        }
        if opts.ContentDisposition != "" {
            req.ResponseContentDisposition = oss.Ptr(opts.ContentDisposition)
        }
    }

    result, err := s.presignClient.PresignGetObject(ctx, req, oss.PresignExpires(expiration))
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }

    return result.URL, nil
}
```

**MIRROR**: Pattern 4 - OSS SDK 响应头设置

**IMPORTS**: 无新增（已有 `oss` 包）

**GOTCHA**:
- `oss.Ptr()` 是必需的，OSS SDK 使用指针字段
- 不要在 `opts == nil` 时设置默认值，保持原有行为

**VALIDATE**:
```bash
go build ./pkg/storage/...
go test ./pkg/storage -run TestOSSStorage
```

---

### Task 3: 实现 S3 存储的响应头设置

**ACTION**: 修改 `pkg/storage/s3.go` 的 `GeneratePresignedDownloadURL` 方法

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/pkg/storage/s3.go`

定位到 line 170-185，替换整个方法:

```go
func (s *S3Storage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration, opts *PresignOptions) (string, error) {
    input := &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    }

    // 设置响应头（如果提供）
    if opts != nil {
        if opts.ContentType != "" {
            input.ResponseContentType = aws.String(opts.ContentType)
        }
        if opts.ContentDisposition != "" {
            input.ResponseContentDisposition = aws.String(opts.ContentDisposition)
        }
    }

    req, err := s.presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expiration))
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }

    return req.URL, nil
}
```

**MIRROR**: Pattern 5 - S3 SDK 响应头设置

**IMPORTS**: 无新增（已有 `aws` 包）

**GOTCHA**:
- `aws.String()` 是必需的，S3 SDK 使用指针字段
- S3 和 OSS 的字段名完全一致，保持代码对称性

**VALIDATE**:
```bash
go build ./pkg/storage/...
go test ./pkg/storage -run TestS3Storage
```

---

### Task 4: 在 Service 层添加 Content-Type 自动检测

**ACTION**: 修改 `file_service.go` 的 `UploadDirect` 方法，添加文件头检测

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/internal/services/file_service.go`

定位到 line 88-136 的 `UploadDirect` 方法，在上传到 OSS 之前添加检测逻辑:

```go
func (s *FileService) UploadDirect(ctx context.Context, req *UploadDirectRequest) (*UploadDirectResponse, error) {
    // ... 现有的验证逻辑 ...

    // 【新增】检测 Content-Type（读取前 512 字节）
    buffer := make([]byte, 512)
    n, err := req.File.Read(buffer)
    if err != nil && err != io.EOF {
        return nil, errors.NewInternalError(err)
    }
    
    // 重置文件指针到开头（重要！）
    if seeker, ok := req.File.(io.Seeker); ok {
        if _, err := seeker.Seek(0, io.SeekStart); err != nil {
            return nil, errors.NewInternalError(err)
        }
    }

    // 自动检测 Content-Type
    detectedType := http.DetectContentType(buffer[:n])
    
    // 如果客户端提供了 Content-Type，验证其合理性
    // 这里简单处理：优先使用检测结果
    contentType := detectedType
    if req.ContentType != "" && req.ContentType != "application/octet-stream" {
        // 客户端提供了非默认值，记录日志但仍使用检测结果
        s.logger.Info("client provided content-type",
            zap.String("provided", req.ContentType),
            zap.String("detected", detectedType),
        )
    }

    // 上传到存储
    uploadKey := fmt.Sprintf("%s/%s", req.Path, req.Filename)
    if err := s.storage.Upload(ctx, uploadKey, req.File, contentType); err != nil {
        return nil, errors.NewInternalError(err)
    }

    // ... 后续逻辑保持不变，使用 contentType 存储到数据库 ...
}
```

**MIRROR**: Pattern 2 - Content-Type 检测

**IMPORTS**:
```go
import (
    "io"
    "net/http"
    // ... 其他已有的 import
)
```

**GOTCHA**:
- **必须** 在读取 512 字节后重置文件指针，否则上传的文件会缺少开头
- `http.DetectContentType()` 对于未知类型返回 `"application/octet-stream"`
- 不要信任客户端提供的 Content-Type，始终使用检测结果

**VALIDATE**:
```bash
go build ./internal/services/...
go test ./internal/services -run TestFileService_UploadDirect
```

---

### Task 5: 修改 Service 层的预签名下载 URL 生成逻辑

**ACTION**: 修改 `file_service.go` 的 `GetPresignedDownloadURL` 方法，传递响应头参数

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/internal/services/file_service.go`

找到 `GetPresignedDownloadURL` 方法（如果不存在则新建），修改调用 `storage.GeneratePresignedDownloadURL` 的部分:

```go
func (s *FileService) GetPresignedDownloadURL(ctx context.Context, fileID uint, expiration time.Duration) (string, error) {
    // 查询文件记录
    var file models.File
    if err := s.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return "", errors.NewNotFoundError("file not found")
        }
        return "", errors.NewInternalError(err)
    }

    // 判断是否可预览
    disposition := "attachment"
    if isPreviewable(file.ContentType) {
        disposition = "inline"
    }

    // 构造响应头选项
    opts := &storage.PresignOptions{
        ContentType:        file.ContentType,
        ContentDisposition: disposition,
    }

    // 生成预签名 URL
    url, err := s.storage.GeneratePresignedDownloadURL(ctx, file.StorageKey, expiration, opts)
    if err != nil {
        return "", errors.NewInternalError(err)
    }

    return url, nil
}

// isPreviewable 判断文件类型是否可在浏览器中预览
func isPreviewable(contentType string) bool {
    previewable := []string{
        "image/",
        "video/",
        "audio/",
        "application/pdf",
        "text/",
    }
    
    for _, prefix := range previewable {
        if strings.HasPrefix(contentType, prefix) {
            return true
        }
    }
    return false
}
```

**MIRROR**: Pattern 3 - 可预览类型判断

**IMPORTS**:
```go
import (
    "strings"
    // ... 其他已有的 import
)
```

**GOTCHA**:
- `isPreviewable()` 可以提取到 `pkg/utils/mime.go`，但先在这里实现（YAGNI 原则）
- 不要使用 `fmt.Sprintf("attachment; filename=%q", file.Filename)`，文件名已经在存储中，不需要重复
- 对于不可预览的文件，使用简单的 `"attachment"` 即可

**VALIDATE**:
```bash
go build ./internal/services/...
go test ./internal/services -run TestFileService_GetPresignedDownloadURL
```

---

### Task 6: 修改 Handler 层支持新的 Service 方法

**ACTION**: 修改 `file_handler.go`，调用新的 `GetPresignedDownloadURL` 方法

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/internal/handlers/file_handler.go`

找到或新建 `GetPresignedDownloadURL` handler 方法:

```go
// GetPresignedDownloadURL godoc
// @Summary      获取预签名下载 URL
// @Description  生成文件的预签名下载 URL，支持浏览器预览
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        id path int true "文件 ID"
// @Param        expiration query int false "过期时间（秒）" default(3600)
// @Success      200 {object} response.Response{data=PresignedURLResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /files/{id}/download-url [get]
func (h *FileHandler) GetPresignedDownloadURL(c *gin.Context) {
    // 解析文件 ID
    fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.Error(errors.NewBadRequestError("invalid file ID", err))
        return
    }

    // 解析过期时间（默认 1 小时）
    expiration := 3600
    if exp := c.Query("expiration"); exp != "" {
        if parsed, err := strconv.Atoi(exp); err == nil && parsed > 0 {
            expiration = parsed
        }
    }

    // 调用 Service 层
    url, err := h.fileService.GetPresignedDownloadURL(c.Request.Context(), uint(fileID), time.Duration(expiration)*time.Second)
    if err != nil {
        c.Error(err)
        return
    }

    // 返回响应
    response.Success(c, PresignedURLResponse{
        URL:       url,
        ExpiresIn: expiration,
    })
}

// PresignedURLResponse 预签名 URL 响应
type PresignedURLResponse struct {
    URL       string `json:"url" example:"https://bucket.oss-cn-hangzhou.aliyuncs.com/path/file.png?signature=xxx"`
    ExpiresIn int    `json:"expires_in" example:"3600"`
}
```

**MIRROR**: API 设计规则 - Handler 结构定义 + Swagger 注释规范

**IMPORTS**:
```go
import (
    "strconv"
    "time"
    // ... 其他已有的 import
)
```

**GOTCHA**:
- 使用 `c.Param("id")` 获取路径参数，不是 `c.Query("id")`
- 过期时间单位是秒，需要转换为 `time.Duration`
- 必须定义 `PresignedURLResponse` 结构体，不要使用 `map[string]interface{}`

**VALIDATE**:
```bash
go build ./internal/handlers/...
make swag-init
```

---

### Task 7: 注册新的路由

**ACTION**: 在 `cmd/api/main.go` 中注册新的路由

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/cmd/api/main.go`

定位到 `setupRouter()` 函数，在文件相关路由部分添加:

```go
func setupRouter(db *gorm.DB, redis *cache.RedisClient, logger *zap.Logger, storage storage.Storage) *gin.Engine {
    // ... 现有的中间件和路由 ...

    // 文件管理路由
    fileHandler := handlers.NewFileHandler(db, storage, logger)
    router.POST("/files/upload", fileHandler.UploadDirect)
    router.GET("/files/:id/download-url", fileHandler.GetPresignedDownloadURL) // 【新增】

    // ... 其他路由 ...
}
```

**MIRROR**: API 设计规则 - 路由注册模式

**IMPORTS**: 无新增

**GOTCHA**:
- 路由路径使用 `:id` 占位符，不是 `{id}`
- 确保 `fileHandler` 已经创建，不要重复创建

**VALIDATE**:
```bash
go build ./cmd/api/...
make run
# 测试路由: curl http://localhost:8003/files/1/download-url
```

---

### Task 8: 添加单元测试

**ACTION**: 为新功能添加单元测试

**IMPLEMENT**:

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/pkg/storage/oss_test.go` (新建或修改)

```go
func TestOSSStorage_GeneratePresignedDownloadURL_WithOptions(t *testing.T) {
    // 跳过集成测试（需要真实的 OSS 凭证）
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    storage := setupOSSStorage(t)
    ctx := context.Background()

    // 测试用例 1: 不传 opts（向后兼容）
    url1, err := storage.GeneratePresignedDownloadURL(ctx, "test.png", time.Hour, nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, url1)

    // 测试用例 2: 传递 Content-Type
    opts := &storage.PresignOptions{
        ContentType:        "image/png",
        ContentDisposition: "inline",
    }
    url2, err := storage.GeneratePresignedDownloadURL(ctx, "test.png", time.Hour, opts)
    assert.NoError(t, err)
    assert.NotEmpty(t, url2)
    assert.Contains(t, url2, "response-content-type=image%2Fpng")
    assert.Contains(t, url2, "response-content-disposition=inline")
}
```

文件: `/Users/fanlz/Projects/doodleEsc/AssetHub/internal/services/file_service_test.go` (新建或修改)

```go
func TestFileService_UploadDirect_DetectContentType(t *testing.T) {
    // Mock 依赖
    db := setupTestDB(t)
    storage := &mockStorage{}
    logger := zap.NewNop()
    service := NewFileService(db, storage, logger)

    // 创建测试文件（PNG 文件头）
    pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
    file := bytes.NewReader(pngHeader)

    req := &UploadDirectRequest{
        File:        file,
        Filename:    "test.png",
        ContentType: "application/octet-stream", // 客户端提供错误的类型
        Path:        "uploads",
    }

    resp, err := service.UploadDirect(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, "image/png", resp.ContentType) // 应该检测为 PNG
}
```

**MIRROR**: Go 标准测试模式

**IMPORTS**:
```go
import (
    "bytes"
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
)
```

**GOTCHA**:
- 集成测试需要真实的 OSS/S3 凭证，使用 `testing.Short()` 跳过
- 单元测试使用 mock 对象，不依赖外部服务
- PNG 文件头: `89 50 4E 47 0D 0A 1A 0A`（8 字节）

**VALIDATE**:
```bash
go test ./pkg/storage/... -short
go test ./internal/services/... -v
```

---

### Task 9: 更新 Swagger 文档

**ACTION**: 重新生成 Swagger 文档

**IMPLEMENT**:

```bash
# 重新生成文档
make swag-init

# 验证生成结果
ls -lh docs/swagger.json docs/swagger.yaml

# 启动服务，访问 Swagger UI
make run
# 浏览器打开: http://localhost:8003/swagger/index.html
```

**MIRROR**: API 设计规则 - 全局 Swagger 配置

**IMPORTS**: 无

**GOTCHA**:
- 必须在所有 handler 修改完成后执行
- 如果 Swagger 注释格式错误，`swag init` 会报错
- 检查 `PresignedURLResponse` 是否正确显示字段（不是 `additionalProp`）

**VALIDATE**:
```bash
# 检查 Swagger JSON 中是否包含新路由
jq '.paths."/files/{id}/download-url"' docs/swagger.json

# 检查响应结构体定义
jq '.definitions.PresignedURLResponse' docs/swagger.json
```

---

## Testing Strategy

### 单元测试

**测试文件**:
- `pkg/storage/oss_test.go` - OSS 响应头设置
- `pkg/storage/s3_test.go` - S3 响应头设置
- `internal/services/file_service_test.go` - Content-Type 检测

**测试用例**:
1. **向后兼容性**: `opts = nil` 时不设置响应头
2. **Content-Type 检测**: 
   - PNG 文件头 → `image/png`
   - JPEG 文件头 → `image/jpeg`
   - PDF 文件头 → `application/pdf`
   - 未知文件 → `application/octet-stream`
3. **可预览类型判断**:
   - `image/png` → `inline`
   - `video/mp4` → `inline`
   - `application/zip` → `attachment`

### 集成测试

**手动测试步骤**:

1. **上传图片文件**:
```bash
curl -X POST http://localhost:8003/files/upload \
  -F "file=@test.png" \
  -F "path=uploads"
```

2. **获取预签名 URL**:
```bash
curl http://localhost:8003/files/1/download-url
```

3. **验证响应头**:
```bash
# 复制上一步返回的 URL
curl -I "https://bucket.oss-cn-hangzhou.aliyuncs.com/path/test.png?signature=xxx"

# 检查响应头:
# Content-Type: image/png
# Content-Disposition: inline
```

4. **浏览器测试**:
   - 在浏览器中打开预签名 URL
   - 图片应该直接显示，而不是下载

### 边缘案例

| 场景 | 输入 | 期望输出 |
|------|------|----------|
| 客户端未提供 Content-Type | `""` | 自动检测 |
| 客户端提供错误的 Content-Type | `"text/plain"` (实际是 PNG) | 使用检测结果 `"image/png"` |
| 文件小于 512 字节 | 100 字节的文本文件 | 正常检测 |
| 空文件 | 0 字节 | `"application/octet-stream"` |
| 不可预览的文件 | `.zip`, `.exe` | `Content-Disposition: attachment` |
| 可预览的文件 | `.png`, `.mp4`, `.pdf` | `Content-Disposition: inline` |

---

## Validation Commands

**使用 Makefile 中的命令**:

```bash
# 1. 编译检查
make build

# 2. 运行测试
make test

# 3. 代码格式化
make fmt

# 4. 代码检查
make lint

# 5. 重新生成 Swagger 文档
make swag-init

# 6. 启动服务
make run

# 7. 查看日志（验证 Content-Type 检测）
make logs
```

**手动验证**:

```bash
# 检查接口签名是否正确
go doc -all pkg/storage | grep GeneratePresignedDownloadURL

# 检查编译错误
go build ./...

# 运行特定测试
go test ./pkg/storage -run TestOSSStorage_GeneratePresignedDownloadURL -v
go test ./internal/services -run TestFileService_UploadDirect -v

# 检查 Swagger 文档
curl http://localhost:8003/swagger/doc.json | jq '.paths'
```

---

## Acceptance Criteria

### 功能验收

- [ ] **AC1**: 上传文件时，系统自动检测 Content-Type（不依赖客户端输入）
  - 验证方法: 上传 PNG 文件，客户端不提供 Content-Type，数据库中存储 `image/png`

- [ ] **AC2**: 客户端提供的 Content-Type 会被验证（使用检测结果）
  - 验证方法: 上传 PNG 文件，客户端提供 `text/plain`，数据库中存储 `image/png`

- [ ] **AC3**: 预签名下载 URL 包含正确的响应头参数
  - 验证方法: 生成预签名 URL，URL 中包含 `response-content-type` 和 `response-content-disposition` 参数

- [ ] **AC4**: 图片/视频/PDF 在浏览器中直接预览
  - 验证方法: 在浏览器中打开预签名 URL，图片直接显示，不触发下载

- [ ] **AC5**: 不可预览的文件（如 .zip）触发下载
  - 验证方法: 在浏览器中打开 .zip 文件的预签名 URL，触发下载对话框

- [ ] **AC6**: 向后兼容性：现有代码调用 `GeneratePresignedDownloadURL(ctx, key, expiration, nil)` 仍然正常工作
  - 验证方法: 不传 `opts` 参数，生成的 URL 不包含响应头参数（保持原有行为）

### 技术验收

- [ ] **TC1**: 所有单元测试通过
  - 验证方法: `make test` 无错误

- [ ] **TC2**: 代码通过 lint 检查
  - 验证方法: `make lint` 无警告

- [ ] **TC3**: Swagger 文档正确生成
  - 验证方法: 访问 `/swagger/index.html`，新路由和响应结构体正确显示

- [ ] **TC4**: 无编译错误
  - 验证方法: `make build` 成功

- [ ] **TC5**: 日志中记录 Content-Type 检测信息
  - 验证方法: 上传文件后，日志中包含 `"detected content-type"` 字段

---

## Risks and Mitigations

### Risk 1: 文件指针未重置导致上传失败

**风险描述**: 读取 512 字节检测 Content-Type 后，如果不重置文件指针，上传的文件会缺少开头部分。

**影响**: 高（数据损坏）

**缓解措施**:
- 在读取后立即调用 `Seek(0, io.SeekStart)`
- 添加单元测试验证文件完整性
- 在日志中记录文件大小，对比上传前后

**检测方法**:
```go
// 上传前记录大小
originalSize := getFileSize(req.File)

// 上传后验证
uploadedSize := getUploadedFileSize(storageKey)
assert.Equal(t, originalSize, uploadedSize)
```

---

### Risk 2: `http.DetectContentType()` 误判

**风险描述**: `http.DetectContentType()` 基于文件头检测，可能误判某些文件类型（如纯文本文件被识别为 `text/plain; charset=utf-8`）。

**影响**: 中（用户体验）

**缓解措施**:
- 对于常见文件扩展名（`.jpg`, `.png`, `.mp4`），优先使用扩展名映射
- 记录检测结果到日志，便于后续分析
- 提供管理接口允许手动修正 Content-Type

**改进方案**（可选）:
```go
// 结合文件扩展名和文件头检测
func detectContentType(filename string, data []byte) string {
    // 1. 先尝试从扩展名推断
    ext := strings.ToLower(filepath.Ext(filename))
    if knownType, ok := extensionMap[ext]; ok {
        return knownType
    }
    
    // 2. 使用文件头检测
    return http.DetectContentType(data)
}
```

---

### Risk 3: OSS/S3 SDK 版本兼容性

**风险描述**: OSS SDK 和 S3 SDK 的响应头参数可能在不同版本中有变化。

**影响**: 中（功能失效）

**缓解措施**:
- 在 `go.mod` 中锁定 SDK 版本
- 添加集成测试验证响应头参数
- 查阅官方文档确认参数名称

**当前版本**:
- OSS SDK: `github.com/aliyun/alibabacloud-oss-go-sdk-v2 v1.4.0`
- S3 SDK: `github.com/aws/aws-sdk-go-v2/service/s3 v1.71.1`

**验证方法**:
```bash
# 检查 SDK 文档
go doc github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss GetObjectRequest
go doc github.com/aws/aws-sdk-go-v2/service/s3 GetObjectInput
```

---

### Risk 4: 向后兼容性破坏

**风险描述**: 修改 `Storage` 接口签名可能导致现有代码编译失败。

**影响**: 高（破坏性变更）

**缓解措施**:
- 使用 `opts *PresignOptions` 指针参数，允许传 `nil`（向后兼容）
- 在所有调用点添加 `nil` 参数（保持原有行为）
- 使用 `grep` 搜索所有调用点，逐一修改

**检查命令**:
```bash
# 搜索所有调用点
rg "GeneratePresignedDownloadURL" --type go

# 编译检查
go build ./...
```

**Linus 点评**: 这是唯一正确的做法。接口变更必须向后兼容，否则就是在破坏用户空间。

---

### Risk 5: 预签名 URL 参数被 URL 编码

**风险描述**: 响应头参数（如 `Content-Disposition: inline`）在 URL 中可能被编码为 `response-content-disposition=inline`，某些 CDN 或代理可能不支持。

**影响**: 低（特定环境）

**缓解措施**:
- 在测试环境验证 URL 格式
- 查阅 OSS/S3 官方文档确认参数格式
- 添加日志记录生成的 URL，便于调试

**验证方法**:
```bash
# 检查生成的 URL 格式
curl -v "https://bucket.oss-cn-hangzhou.aliyuncs.com/test.png?response-content-type=image%2Fpng"

# 验证响应头
curl -I "..." | grep -i content-type
```

---

## Implementation Checklist

**开始实施前**:
- [ ] 阅读所有 Mandatory Reading 文件
- [ ] 理解 Pattern 1-5 的设计模式
- [ ] 确认 OSS/S3 SDK 版本

**实施过程中**:
- [ ] 按照 Task 1-9 的顺序执行
- [ ] 每完成一个 Task，运行对应的 VALIDATE 命令
- [ ] 提交代码前运行 `make test` 和 `make lint`

**实施完成后**:
- [ ] 运行所有 Validation Commands
- [ ] 验证所有 Acceptance Criteria
- [ ] 手动测试 Before/After 场景
- [ ] 更新 CHANGELOG.md（如果有）

---

## Linus 最后的话

**【品味】**: 🟢 好品味 (Good Taste)

这个方案简洁、直接、解决真问题。

**【核心洞察】**:
1. **数据结构**: 用 `PresignOptions` 结构体封装可选参数，而不是无限增加函数参数。这是 Go 的惯用法。
2. **复杂性**: 用标准库 `http.DetectContentType()`，不引入第三方依赖。512 字节足够。
3. **兼容性**: `opts *PresignOptions` 使用指针，允许传 `nil`。向后兼容是神圣不可侵犯的。

**【关键点】**:
- 读取 512 字节后 **必须** 重置文件指针，否则数据损坏。
- 不要信任客户端提供的 Content-Type，始终使用检测结果。
- `inline` vs `attachment` 的判断逻辑要简单，不要搞复杂。

**【最后警告】**:
如果你在实施过程中发现需要超过 3 层缩进，停下来。你的数据结构错了。

滚去实施吧。

---

**END OF PLAN**
