# 统一文件存储服务 - 接入文档

## 概述

统一文件存储服务提供了企业级的文件管理能力，支持小文件直接上传和大文件分片上传，通过预签名 URL 确保文件访问安全。

## 快速开始

### 1. 配置 S3

在 `.env` 文件中配置 S3 凭证：

```bash
# S3 Storage
S3_ACCESS_KEY_ID=your_access_key_id
S3_SECRET_ACCESS_KEY=your_secret_access_key
S3_BUCKET=your_bucket_name
S3_REGION=us-east-1
```

### 2. 运行数据库迁移

```bash
./scripts/migrate.sh up
```

### 3. 启动服务

```bash
go run cmd/api/main.go
```

服务将在 `http://localhost:8003` 启动。

### 4. 访问 Swagger 文档

打开浏览器访问：`http://localhost:8003/swagger/index.html`

## API 使用指南

### 小文件上传（< 100MB）

#### 方式 1：后端代理上传

```bash
curl -X POST http://localhost:8003/api/v1/files/upload \
  -F "name=example.txt" \
  -F "content_type=text/plain" \
  -F "file=@/path/to/file.txt"
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 1,
    "name": "example.txt",
    "size": 1024,
    "storage_key": "files/1234567890/example.txt",
    "status": "completed",
    "download_url": "https://s3.amazonaws.com/..."
  }
}
```

#### 方式 2：前端直传

**步骤 1：获取预签名 URL**

```bash
curl -X POST http://localhost:8003/api/v1/files/upload/presigned \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.txt",
    "content_type": "text/plain",
    "size": 1024
  }'
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 1,
    "upload_url": "https://s3.amazonaws.com/...",
    "storage_key": "files/1234567890/example.txt",
    "expires_in": 3600
  }
}
```

**步骤 2：前端直接上传到 S3**

```javascript
const response = await fetch(uploadUrl, {
  method: 'PUT',
  body: file,
  headers: {
    'Content-Type': file.type
  }
});
```

**步骤 3：确认上传完成**

```bash
curl -X POST http://localhost:8003/api/v1/files/upload/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": 1
  }'
```

### 大文件分片上传（>= 100MB）

#### 步骤 1：初始化分片上传

```bash
curl -X POST http://localhost:8003/api/v1/files/upload/multipart/init \
  -H "Content-Type: application/json" \
  -d '{
    "name": "large-video.mp4",
    "content_type": "video/mp4",
    "size": 104857600
  }'
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 1,
    "upload_id": "upload-id-123",
    "storage_key": "files/1234567890/large-video.mp4"
  }
}
```

#### 步骤 2：为每个分片生成预签名 URL

```bash
curl -X POST http://localhost:8003/api/v1/files/upload/multipart/part-url \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": 1,
    "part_number": 1
  }'
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "part_number": 1,
    "upload_url": "https://s3.amazonaws.com/...",
    "expires_in": 3600
  }
}
```

#### 步骤 3：上传每个分片到 S3

```javascript
const response = await fetch(partUrl, {
  method: 'PUT',
  body: partData
});

// 保存 ETag（从响应头中获取）
const etag = response.headers.get('ETag');
```

#### 步骤 4：完成分片上传

```bash
curl -X POST http://localhost:8003/api/v1/files/upload/multipart/complete \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": 1,
    "parts": [
      {"part_number": 1, "etag": "\"abc123\""},
      {"part_number": 2, "etag": "\"def456\""}
    ]
  }'
```

### 文件下载

#### 获取下载 URL

```bash
curl -X GET http://localhost:8003/api/v1/files/1/download-url
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": 1,
    "download_url": "https://s3.amazonaws.com/...",
    "expires_in": 900
  }
}
```

#### 下载文件

```bash
curl -o downloaded-file.txt "https://s3.amazonaws.com/..."
```

### 文件管理

#### 获取文件信息

```bash
curl -X GET http://localhost:8003/api/v1/files/1
```

#### 删除文件

```bash
curl -X DELETE http://localhost:8003/api/v1/files/1
```

## 在业务服务中集成

### Go 语言集成示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const storageServiceURL = "http://localhost:8003"

// UploadFile 上传文件到存储服务
func UploadFile(filename string, content []byte) (uint, error) {
    // 创建 multipart 请求
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 添加文件字段
    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return 0, err
    }
    io.Copy(part, bytes.NewReader(content))

    // 添加其他字段
    writer.WriteField("name", filename)
    writer.WriteField("content_type", "application/octet-stream")
    writer.Close()

    // 发送请求
    req, _ := http.NewRequest("POST", storageServiceURL+"/api/v1/files/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    // 解析响应
    var result struct {
        Code int `json:"code"`
        Data struct {
            FileID uint `json:"file_id"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Data.FileID, nil
}

// GetDownloadURL 获取文件下载 URL
func GetDownloadURL(fileID uint) (string, error) {
    url := fmt.Sprintf("%s/api/v1/files/%d/download-url", storageServiceURL, fileID)

    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Code int `json:"code"`
        Data struct {
            DownloadURL string `json:"download_url"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Data.DownloadURL, nil
}
```

### JavaScript/TypeScript 集成示例

```typescript
const STORAGE_SERVICE_URL = 'http://localhost:8003';

// 小文件上传（前端直传）
async function uploadSmallFile(file: File): Promise<number> {
  // 1. 获取预签名 URL
  const initResponse = await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/upload/presigned`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: file.name,
      content_type: file.type,
      size: file.size
    })
  });

  const { data } = await initResponse.json();
  const { file_id, upload_url } = data;

  // 2. 上传到 S3
  await fetch(upload_url, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': file.type }
  });

  // 3. 确认上传
  await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/upload/confirm`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ file_id })
  });

  return file_id;
}

// 大文件分片上传
async function uploadLargeFile(file: File): Promise<number> {
  const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk

  // 1. 初始化分片上传
  const initResponse = await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/upload/multipart/init`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: file.name,
      content_type: file.type,
      size: file.size
    })
  });

  const { data: initData } = await initResponse.json();
  const { file_id, upload_id } = initData;

  // 2. 分片上传
  const parts = [];
  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

  for (let i = 0; i < totalChunks; i++) {
    const start = i * CHUNK_SIZE;
    const end = Math.min(start + CHUNK_SIZE, file.size);
    const chunk = file.slice(start, end);
    const partNumber = i + 1;

    // 获取分片预签名 URL
    const urlResponse = await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/upload/multipart/part-url`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ file_id, part_number: partNumber })
    });

    const { data: urlData } = await urlResponse.json();

    // 上传分片
    const uploadResponse = await fetch(urlData.upload_url, {
      method: 'PUT',
      body: chunk
    });

    const etag = uploadResponse.headers.get('ETag');
    parts.push({ part_number: partNumber, etag });
  }

  // 3. 完成分片上传
  await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/upload/multipart/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ file_id, parts })
  });

  return file_id;
}

// 获取下载 URL
async function getDownloadURL(fileId: number): Promise<string> {
  const response = await fetch(`${STORAGE_SERVICE_URL}/api/v1/files/${fileId}/download-url`);
  const { data } = await response.json();
  return data.download_url;
}
```

## 最佳实践

### 1. 文件大小阈值

- **小文件（< 100MB）**：使用直接上传或预签名 URL 上传
- **大文件（>= 100MB）**：使用分片上传

### 2. 分片大小建议

- 最小分片大小：5MB
- 推荐分片大小：5MB - 100MB
- 最大分片数量：10,000

### 3. 错误处理

所有 API 返回统一的错误格式：

```json
{
  "code": 400,
  "message": "invalid request",
  "data": null
}
```

常见错误码：
- `400`：请求参数错误
- `404`：文件不存在
- `500`：服务器内部错误

### 4. 安全建议

- 预签名 URL 有时效性（上传 1 小时，下载 15 分钟）
- 不要将预签名 URL 暴露给未授权用户
- 定期轮换 S3 访问密钥

### 5. 性能优化

- 使用前端直传减少服务器带宽消耗
- 大文件使用分片上传提高成功率
- 合理设置分片大小平衡性能和可靠性

## 故障排查

### 问题 1：上传失败

**可能原因**：
- S3 凭证配置错误
- 存储桶不存在或无权限
- 网络连接问题

**解决方法**：
1. 检查 `.env` 文件中的 S3 配置
2. 验证 S3 存储桶权限
3. 检查服务日志

### 问题 2：预签名 URL 过期

**可能原因**：
- URL 生成后超过有效期未使用

**解决方法**：
- 重新获取预签名 URL
- 调整客户端上传逻辑，减少延迟

### 问题 3：分片上传失败

**可能原因**：
- 分片顺序错误
- ETag 不匹配
- 部分分片上传失败

**解决方法**：
- 确保分片按顺序上传
- 正确保存每个分片的 ETag
- 实现重试机制

## 支持

如有问题，请联系开发团队或查看项目文档。
