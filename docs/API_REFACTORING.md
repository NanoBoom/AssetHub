# RESTful API 重构文档

## 概述

本次重构将 AssetHub API 从非 RESTful 设计改为符合 RESTful 最佳实践的设计。

**重构日期**：2026-02-11
**影响范围**：所有文件上传相关的 API 端点
**向后兼容性**：**不兼容**（直接替换旧端点）

---

## 核心改进

### 1. URL 中移除动词
- ❌ 旧设计：`/files/upload`、`/files/upload/confirm`、`/files/upload/multipart/init`
- ✅ 新设计：`/files`、`/files/{id}/completion`、`/files/multipart`

### 2. 简化路径层级
- ❌ 旧设计：`/files/upload/multipart/init`（4 层）
- ✅ 新设计：`/files/multipart`（2 层）

### 3. 正确的 HTTP 状态码
- ✅ 创建资源：**201 Created**（原 200 OK）
- ✅ 删除成功：**204 No Content**（原 200 OK + JSON 响应）

### 4. 语义清晰的子资源
- ✅ `/files/{id}/completion` - 表示"完成上传"操作
- ✅ `/files/{id}/link` - 表示"获取下载链接"
- ✅ `/files/{id}/multipart/parts` - 表示"分片资源"

---

## API 端点对比

### 文件资源（Files）

| 旧端点 | 新端点 | 方法 | 状态码变化 | 说明 |
|--------|--------|------|-----------|------|
| `POST /api/v1/files/upload` | `POST /api/v1/files` | POST | 200 → **201** | 直接上传 |
| `POST /api/v1/files/upload/presigned` | `POST /api/v1/files/presigned` | POST | 200 → **201** | 预签名上传 |
| `POST /api/v1/files/upload/confirm` | `POST /api/v1/files/{id}/completion` | POST | 200 | 确认上传完成 |
| `GET /api/v1/files/{id}` | `GET /api/v1/files/{id}` | GET | 200 | 获取文件元数据（无变化） |
| `GET /api/v1/files/{id}/download-url` | `GET /api/v1/files/{id}/link` | GET | 200 | 获取下载 URL |
| `DELETE /api/v1/files/{id}` | `DELETE /api/v1/files/{id}` | DELETE | 200 → **204** | 删除文件 |

### 分片上传（Multipart）

| 旧端点 | 新端点 | 方法 | 状态码变化 | 说明 |
|--------|--------|------|-----------|------|
| `POST /api/v1/files/upload/multipart/init` | `POST /api/v1/files/multipart` | POST | 200 → **201** | 初始化分片上传 |
| `POST /api/v1/files/upload/multipart/part-url` | `POST /api/v1/files/{id}/multipart/parts` | POST | 200 | 获取分片 URL |
| `POST /api/v1/files/upload/multipart/complete` | `POST /api/v1/files/{id}/multipart/completion` | POST | 200 | 完成分片上传 |

---

## 详细变更说明

### 1. 直接上传（小文件）

**旧端点**：
```http
POST /api/v1/files/upload
Content-Type: multipart/form-data

name: example.txt
file: <binary>

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files
Content-Type: multipart/form-data

name: example.txt
file: <binary>

Response: 201 Created
```

**关键变化**：
- ✅ 路径从 `/files/upload` 改为 `/files`
- ✅ 状态码从 200 改为 **201 Created**

---

### 2. 预签名上传（中等文件）

**旧端点**：
```http
POST /api/v1/files/upload/presigned
Content-Type: application/json

{
  "name": "example.txt",
  "content_type": "text/plain",
  "size": 1024
}

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files/presigned
Content-Type: application/json

{
  "name": "example.txt",
  "content_type": "text/plain",
  "size": 1024
}

Response: 201 Created
```

**关键变化**：
- ✅ 路径从 `/files/upload/presigned` 改为 `/files/presigned`
- ✅ 状态码从 200 改为 **201 Created**

---

### 3. 确认上传完成

**旧端点**：
```http
POST /api/v1/files/upload/confirm
Content-Type: application/json

{
  "file_id": "550e8400-e29b-41d4-a716-446655440000"
}

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files/{id}/completion
Content-Type: application/json

{}

Response: 200 OK
```

**关键变化**：
- ✅ 路径从 `/files/upload/confirm` 改为 `/files/{id}/completion`
- ✅ `file_id` 从请求体移到 URL 路径参数
- ✅ 请求体可以为空（或省略）

---

### 4. 获取下载 URL

**旧端点**：
```http
GET /api/v1/files/{id}/download-url

Response: 200 OK
{
  "file_id": "...",
  "download_url": "https://...",
  "expires_in": 900
}
```

**新端点**：
```http
GET /api/v1/files/{id}/link

Response: 200 OK
{
  "file_id": "...",
  "download_url": "https://...",
  "expires_in": 900
}
```

**关键变化**：
- ✅ 路径从 `/files/{id}/download-url` 改为 `/files/{id}/link`
- ✅ 更简洁，语义清晰

---

### 5. 删除文件

**旧端点**：
```http
DELETE /api/v1/files/{id}

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": "...",
    "message": "File deleted successfully"
  }
}
```

**新端点**：
```http
DELETE /api/v1/files/{id}

Response: 204 No Content
(无响应体)
```

**关键变化**：
- ✅ 状态码从 200 改为 **204 No Content**
- ✅ **不返回响应体**（符合 RESTful 标准）

---

### 6. 初始化分片上传

**旧端点**：
```http
POST /api/v1/files/upload/multipart/init
Content-Type: application/json

{
  "name": "large-video.mp4",
  "content_type": "video/mp4",
  "size": 104857600
}

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files/multipart
Content-Type: application/json

{
  "name": "large-video.mp4",
  "content_type": "video/mp4",
  "size": 104857600
}

Response: 201 Created
```

**关键变化**：
- ✅ 路径从 `/files/upload/multipart/init` 改为 `/files/multipart`
- ✅ 状态码从 200 改为 **201 Created**
- ✅ 简化路径层级（4 层 → 2 层）

---

### 7. 获取分片上传 URL

**旧端点**：
```http
POST /api/v1/files/upload/multipart/part-url
Content-Type: application/json

{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "part_number": 1
}

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files/{id}/multipart/parts
Content-Type: application/json

{
  "part_number": 1
}

Response: 200 OK
```

**关键变化**：
- ✅ 路径从 `/files/upload/multipart/part-url` 改为 `/files/{id}/multipart/parts`
- ✅ `file_id` 从请求体移到 URL 路径参数
- ✅ 使用复数 `parts` 表示资源集合

---

### 8. 完成分片上传

**旧端点**：
```http
POST /api/v1/files/upload/multipart/complete
Content-Type: application/json

{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "parts": [
    {"part_number": 1, "etag": "\"abc123\""},
    {"part_number": 2, "etag": "\"def456\""}
  ]
}

Response: 200 OK
```

**新端点**：
```http
POST /api/v1/files/{id}/multipart/completion
Content-Type: application/json

{
  "parts": [
    {"part_number": 1, "etag": "\"abc123\""},
    {"part_number": 2, "etag": "\"def456\""}
  ]
}

Response: 200 OK
```

**关键变化**：
- ✅ 路径从 `/files/upload/multipart/complete` 改为 `/files/{id}/multipart/completion`
- ✅ `file_id` 从请求体移到 URL 路径参数
- ✅ 使用子资源 `/completion` 表示操作

---

## 迁移指南

### 客户端需要修改的地方

#### 1. 更新端点 URL

```javascript
// 旧代码
const response = await fetch('/api/v1/files/upload', {
  method: 'POST',
  body: formData
});

// 新代码
const response = await fetch('/api/v1/files', {
  method: 'POST',
  body: formData
});
```

#### 2. 处理新的状态码

```javascript
// 旧代码
if (response.status === 200) {
  const data = await response.json();
  console.log('上传成功', data);
}

// 新代码
if (response.status === 201) {  // 注意：201 而不是 200
  const data = await response.json();
  console.log('上传成功', data);
}
```

#### 3. 删除操作不再返回响应体

```javascript
// 旧代码
const response = await fetch(`/api/v1/files/${id}`, {
  method: 'DELETE'
});
const data = await response.json();
console.log(data.message);  // "File deleted successfully"

// 新代码
const response = await fetch(`/api/v1/files/${id}`, {
  method: 'DELETE'
});
if (response.status === 204) {
  console.log('删除成功');  // 无响应体
}
```

#### 4. 确认上传完成的请求体变化

```javascript
// 旧代码
await fetch('/api/v1/files/upload/confirm', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ file_id: fileId })
});

// 新代码
await fetch(`/api/v1/files/${fileId}/completion`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({})  // 可以为空
});
```

---

## 技术实现细节

### 修改的文件

1. **路由配置**：`cmd/api/main.go`
   - 更新所有路由路径
   - 添加注释说明新的端点结构

2. **Handler 实现**：`internal/handlers/file_handler.go`
   - 更新 Swagger 注释（`@Router` 和 `@Success`）
   - 调整 HTTP 状态码（201、204）
   - 修改部分 Handler 逻辑（从路径参数获取 file_id）

3. **Swagger 文档**：`docs/`
   - 重新生成 `docs.go`、`swagger.json`、`swagger.yaml`

### 未修改的部分

- ✅ Service 层逻辑（`internal/services/file_service.go`）
- ✅ Repository 层（`internal/repositories/file_repository.go`）
- ✅ 数据模型（`internal/models/file.go`）
- ✅ 存储层（`pkg/storage/`）
- ✅ 响应结构体（`pkg/response/response.go`）

---

## RESTful 设计原则总结

本次重构遵循以下 RESTful 最佳实践：

### 1. 资源导向
- URL 表示资源，而不是操作
- 使用名词而非动词

### 2. HTTP 方法语义
- `POST` - 创建资源
- `GET` - 获取资源
- `DELETE` - 删除资源

### 3. 正确的状态码
- `201 Created` - 成功创建资源
- `200 OK` - 成功获取或更新资源
- `204 No Content` - 成功删除资源（无响应体）
- `400 Bad Request` - 客户端错误
- `404 Not Found` - 资源不存在
- `500 Internal Server Error` - 服务器错误

### 4. 子资源表示操作
- `/files/{id}/completion` - 表示"完成上传"操作
- `/files/{id}/multipart/parts` - 表示"分片资源"
- `/files/{id}/link` - 表示"下载链接"

### 5. 实用主义
- 保留 `/presigned` 后缀（明确表示预签名上传）
- 保留 `/multipart` 路径（明确表示分片上传）
- 不过度追求 REST 纯粹性，优先考虑 API 的清晰性和易用性

---

## 参考资料

- [AWS S3 Multipart Upload Guide](https://aws.amazon.com/blogs/compute/uploading-large-objects-to-amazon-s3-using-multipart-upload-and-transfer-acceleration/)
- [HTTP Status Codes for REST APIs](https://www.fullstackprep.dev/Articles/sda/apidesign/status-codes-in-rest)
- [API Endpoint Naming Guidelines](https://www.pranaypourkar.co.in/the-programmers-guide/api/naming-guidelines/api-endpoint-naming)
- [GitHub Release Assets API](https://docs.github.com/en/rest/releases/assets)
- [Stripe File Upload API](https://docs.stripe.com/api/files/create)

---

## 后续建议

### 1. 测试
- 更新集成测试以适配新的端点和状态码
- 测试所有上传流程（直接上传、预签名上传、分片上传）
- 验证错误处理和边界情况

### 2. 文档
- 更新 API 文档（README.md）
- 提供客户端迁移示例
- 更新 Postman/Insomnia 集合

### 3. 监控
- 监控新端点的错误率和响应时间
- 确保客户端已完成迁移

### 4. 未来优化
- 考虑使用 S3 事件通知自动确认上传完成（去掉 `/completion` 端点）
- 考虑支持批量操作（如批量删除）
- 考虑添加文件列表查询端点（`GET /files?page=1&limit=20`）

---

**重构完成日期**：2026-02-11
**重构负责人**：Claude Sonnet 4.5
**审核状态**：待审核
