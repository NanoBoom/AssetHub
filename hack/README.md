# 测试脚本说明

本目录包含用于测试 AssetHub 文件上传功能的脚本。

## 脚本列表

### 1. test_upload_simple.sh ✅ 推荐
**简化版文件上传测试**

测试内容：
- 小文件直接上传（后端代理模式）
- UUID 格式验证（有效/无效/Nil UUID）
- 预签名 URL 下载验证
- 直接下载端点验证（流式传输）
- 文件删除功能

使用方法：
```bash
./hack/test_upload_simple.sh
```

**状态**: ✅ 完全可用

---

### 2. test_presigned_upload.sh ⚠️ 需要配置
**小文件预签名 URL 上传测试**

测试流程：
1. 初始化预签名上传，获取 file_id 和 upload_url
2. 使用预签名 URL 直接上传到 OSS/S3
3. 确认上传完成
4. 获取预签名下载 URL 并验证
5. 测试直接下载端点（流式传输）

使用方法：
```bash
./hack/test_presigned_upload.sh
```

**注意事项**：
- 需要正确配置 OSS/S3 凭证
- 预签名 URL 的签名必须与上传参数匹配
- 如果遇到 403 错误，检查 Content-Type 和签名配置

**状态**: ⚠️ 需要 OSS 配置正确

---

### 3. test_multipart_upload.sh ⚠️ 需要配置
**大文件分片上传测试**

测试流程：
1. 初始化分片上传，获取 file_id 和 upload_id
2. 为每个分片生成预签名 URL
3. 上传所有分片（默认 5MB/分片）
4. 完成分片上传
5. 获取预签名下载 URL 并验证
6. 测试直接下载端点（流式传输）

使用方法：
```bash
./hack/test_multipart_upload.sh
```

**注意事项**：
- 会创建 10MB 的测试文件
- 需要正确配置 OSS/S3 凭证
- 分片上传需要 ETag 验证

**状态**: ⚠️ 需要 OSS 配置正确

---

## UUID 迁移验证

所有脚本都验证了 UUID 功能：

### ✅ 已验证的功能
- UUID v4 自动生成（应用层 BeforeCreate hook）
- UUID 格式验证（路径参数）
- UUID 格式验证（JSON body）
- Nil UUID 拒绝
- 无效格式拒绝（整数、错误格式）
- 错误处理（400/404/500）

### 测试结果示例
```bash
# 有效 UUID
GET /api/v1/files/7c91d7ec-bdb2-42e7-a63a-9fad6ca4c2d2
→ 200 OK

# 无效 UUID（整数）
GET /api/v1/files/123
→ 400 "invalid or nil UUID"

# 无效 UUID（格式错误）
GET /api/v1/files/invalid-uuid
→ 400 "invalid or nil UUID"

# Nil UUID
GET /api/v1/files/00000000-0000-0000-0000-000000000000
→ 400 "invalid or nil UUID"
```

---

## 环境要求

- curl
- jq
- bash
- 运行中的 AssetHub 服务（默认端口 8003）

---

## 故障排除

### 问题 1: 应用未运行
```bash
# 检查应用状态
curl http://localhost:8003/health

# 启动应用
go run cmd/api/main.go
```

### 问题 2: 预签名 URL 403 错误
- 检查 OSS/S3 凭证配置（.env 文件）
- 确认 Content-Type 匹配
- 验证时间同步

### 问题 3: jq 命令未找到
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq
```

---

## 贡献

如果发现问题或有改进建议，请提交 Issue 或 PR。
