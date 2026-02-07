# S3 配置指南

## 概述

统一文件存储服务使用 AWS S3（或兼容 S3 的对象存储）作为后端存储。本文档说明如何配置 S3 连接。

## 配置方式

### 方式 1：环境变量（推荐）

在 `.env` 文件中配置：

```bash
# S3 Storage
S3_ACCESS_KEY_ID=your_access_key_id
S3_SECRET_ACCESS_KEY=your_secret_access_key
S3_BUCKET=your_bucket_name
S3_REGION=us-east-1
```

### 方式 2：YAML 配置文件

在 `configs/config.yaml` 中配置：

```yaml
storage:
  type: "s3"
  s3:
    region: "us-east-1"
    bucket: "your-bucket-name"
    access_key_id: "your-access-key-id"
    secret_access_key: "your-secret-access-key"
    endpoint: ""
    use_path_style: false
```

**注意**：环境变量优先级高于配置文件。

## AWS S3 配置

### 1. 创建 S3 存储桶

```bash
aws s3 mb s3://your-bucket-name --region us-east-1
```

### 2. 配置存储桶 CORS

为了支持前端直传，需要配置 CORS：

```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
    "AllowedOrigins": ["*"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3000
  }
]
```

应用 CORS 配置：

```bash
aws s3api put-bucket-cors \
  --bucket your-bucket-name \
  --cors-configuration file://cors.json
```

### 3. 创建 IAM 用户

创建专用的 IAM 用户用于文件存储服务：

```bash
aws iam create-user --user-name storage-service
```

### 4. 配置 IAM 策略

创建策略文件 `storage-policy.json`：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:ListBucket",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket-name",
        "arn:aws:s3:::your-bucket-name/*"
      ]
    }
  ]
}
```

应用策略：

```bash
aws iam put-user-policy \
  --user-name storage-service \
  --policy-name StorageServicePolicy \
  --policy-document file://storage-policy.json
```

### 5. 生成访问密钥

```bash
aws iam create-access-key --user-name storage-service
```

输出示例：

```json
{
  "AccessKey": {
    "UserName": "storage-service",
    "AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
    "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "Status": "Active"
  }
}
```

将 `AccessKeyId` 和 `SecretAccessKey` 配置到 `.env` 文件中。

## 阿里云 OSS 配置

阿里云 OSS 兼容 S3 API，可以使用相同的配置方式。

### 1. 创建 OSS 存储桶

在阿里云控制台创建 OSS Bucket。

### 2. 获取访问凭证

在阿里云控制台创建 AccessKey。

### 3. 配置端点

在 `configs/config.yaml` 中配置：

```yaml
storage:
  type: "s3"
  s3:
    region: "oss-cn-hangzhou"
    bucket: "your-bucket-name"
    access_key_id: "your-access-key-id"
    secret_access_key: "your-secret-access-key"
    endpoint: "https://oss-cn-hangzhou.aliyuncs.com"
    use_path_style: false
```

## MinIO 配置

MinIO 是开源的 S3 兼容对象存储，适合本地开发和私有部署。

### 1. 启动 MinIO

使用 Docker 启动：

```bash
docker run -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"
```

### 2. 创建存储桶

访问 MinIO 控制台：`http://localhost:9001`

使用默认凭证登录：
- 用户名：`minioadmin`
- 密码：`minioadmin`

创建一个新的 Bucket。

### 3. 配置服务

在 `configs/config.yaml` 中配置：

```yaml
storage:
  type: "s3"
  s3:
    region: "us-east-1"
    bucket: "your-bucket-name"
    access_key_id: "minioadmin"
    secret_access_key: "minioadmin"
    endpoint: "http://localhost:9000"
    use_path_style: true  # MinIO 需要设置为 true
```

## 本地存储配置

用于开发环境，不依赖 S3。

在 `configs/config.yaml` 中配置：

```yaml
storage:
  type: "local"
  local:
    base_path: "./storage"
```

**注意**：本地存储不支持预签名 URL，仅用于开发测试。

## 配置验证

### 1. 检查配置

启动服务后，检查日志确认 S3 连接成功：

```bash
go run cmd/api/main.go
```

### 2. 测试上传

使用 Health Check 端点验证服务运行：

```bash
curl http://localhost:8003/health
```

使用文件上传 API 测试 S3 连接：

```bash
curl -X POST http://localhost:8003/api/v1/files/upload \
  -F "name=test.txt" \
  -F "file=@test.txt"
```

## 安全最佳实践

### 1. 访问密钥管理

- ✅ 使用环境变量存储密钥
- ✅ 不要将密钥提交到版本控制
- ✅ 定期轮换访问密钥
- ❌ 不要在代码中硬编码密钥

### 2. 存储桶权限

- ✅ 使用最小权限原则
- ✅ 禁用公共访问
- ✅ 使用预签名 URL 控制访问
- ❌ 不要开放存储桶公共读写

### 3. 网络安全

- ✅ 使用 HTTPS 传输
- ✅ 配置 VPC 端点（生产环境）
- ✅ 启用服务器端加密
- ❌ 不要在公网暴露 MinIO

### 4. 监控和审计

- ✅ 启用 S3 访问日志
- ✅ 配置 CloudWatch 告警
- ✅ 定期审查 IAM 权限
- ✅ 监控存储使用量和成本

## 故障排查

### 问题 1：连接失败

**错误信息**：`failed to load AWS config`

**解决方法**：
1. 检查 `.env` 文件是否存在
2. 验证环境变量是否正确加载
3. 检查 S3 区域配置是否正确

### 问题 2：权限拒绝

**错误信息**：`Access Denied`

**解决方法**：
1. 验证 IAM 用户权限
2. 检查存储桶策略
3. 确认访问密钥有效

### 问题 3：存储桶不存在

**错误信息**：`NoSuchBucket`

**解决方法**：
1. 确认存储桶名称正确
2. 验证存储桶所在区域
3. 检查存储桶是否已创建

### 问题 4：CORS 错误

**错误信息**：`CORS policy blocked`

**解决方法**：
1. 配置存储桶 CORS 策略
2. 检查 AllowedOrigins 配置
3. 确保 ExposeHeaders 包含 ETag

## 参考资料

- [AWS S3 文档](https://docs.aws.amazon.com/s3/)
- [阿里云 OSS 文档](https://help.aliyun.com/product/31815.html)
- [MinIO 文档](https://min.io/docs/minio/linux/index.html)
- [S3 API 兼容性](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html)
