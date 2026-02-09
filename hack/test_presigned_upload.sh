#!/bin/bash

# 小文件预签名 URL 上传测试脚本
# 测试流程：
# 1. 初始化预签名上传，获取 file_id 和 upload_url
# 2. 使用预签名 URL 直接上传到 OSS/S3
# 3. 确认上传完成
# 4. 获取下载 URL 并验证

set -e

API_BASE="http://localhost:8003/api/v1"
TEST_FILE="/tmp/test_presigned_$(date +%s).txt"

echo "=========================================="
echo "小文件预签名 URL 上传测试"
echo "=========================================="
echo ""

# 创建测试文件
echo "1. 创建测试文件..."
echo "This is a test file for presigned URL upload. Created at $(date)" > "$TEST_FILE"
FILE_SIZE=$(wc -c < "$TEST_FILE" | tr -d ' ')
echo "   文件路径: $TEST_FILE"
echo "   文件大小: $FILE_SIZE bytes"
echo ""

# 步骤 1: 初始化预签名上传
echo "2. 初始化预签名上传..."
INIT_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload/presigned" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"test_presigned.txt\",
    \"content_type\": \"text/plain\",
    \"size\": $FILE_SIZE
  }")

# 检查响应是否为空
if [ -z "$INIT_RESPONSE" ]; then
  echo "   ❌ 错误: API 无响应"
  exit 1
fi

# 检查响应是否为有效 JSON
if ! echo "$INIT_RESPONSE" | jq . > /dev/null 2>&1; then
  echo "   ❌ 错误: 响应不是有效的 JSON"
  echo "   原始响应: $INIT_RESPONSE"
  exit 1
fi

echo "   响应: $(echo "$INIT_RESPONSE" | jq -c .)"
echo ""

# 提取 file_id 和 upload_url
FILE_ID=$(echo "$INIT_RESPONSE" | jq -r '.data.file_id')
UPLOAD_URL=$(echo "$INIT_RESPONSE" | jq -r '.data.upload_url')
STORAGE_KEY=$(echo "$INIT_RESPONSE" | jq -r '.data.storage_key')

if [ "$FILE_ID" == "null" ] || [ -z "$FILE_ID" ]; then
  echo "❌ 错误: 无法获取 file_id"
  exit 1
fi

echo "   ✅ file_id: $FILE_ID"
echo "   ✅ storage_key: $STORAGE_KEY"
echo ""

# 步骤 2: 使用预签名 URL 上传文件
echo "3. 使用预签名 URL 上传文件到 OSS/S3..."
echo "   (注意: 必须指定 Content-Type: text/plain，与初始化时一致)"
UPLOAD_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$UPLOAD_URL" \
  -H "Content-Type: text/plain" \
  --data-binary "@$TEST_FILE")

if [ "$UPLOAD_STATUS" == "200" ]; then
  echo "   ✅ 上传成功 (HTTP $UPLOAD_STATUS)"
else
  echo "   ❌ 上传失败 (HTTP $UPLOAD_STATUS)"
  echo "   提示: 如果是 403 错误，可能是签名问题或 Content-Type 不匹配"
  exit 1
fi
echo ""

# 步骤 3: 确认上传完成
echo "4. 确认上传完成..."
CONFIRM_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload/confirm" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_id\": \"$FILE_ID\"
  }")

echo "   响应: $(echo "$CONFIRM_RESPONSE" | jq -c .)"
CONFIRM_STATUS=$(echo "$CONFIRM_RESPONSE" | jq -r '.data.status')

if [ "$CONFIRM_STATUS" == "completed" ]; then
  echo "   ✅ 状态已更新为 completed"
else
  echo "   ⚠️  状态: $CONFIRM_STATUS"
fi
echo ""

# 步骤 4: 获取文件信息
echo "5. 获取文件信息..."
FILE_INFO=$(curl -s "$API_BASE/files/$FILE_ID")
echo "   响应: $(echo "$FILE_INFO" | jq -c .)"
echo ""

# 步骤 5: 获取下载 URL
echo "6. 获取下载 URL..."
DOWNLOAD_RESPONSE=$(curl -s "$API_BASE/files/$FILE_ID/download-url")
DOWNLOAD_URL=$(echo "$DOWNLOAD_RESPONSE" | jq -r '.data.download_url')
echo "   下载 URL: $DOWNLOAD_URL"
echo ""

# 步骤 6: 验证下载
echo "7. 验证下载..."
DOWNLOAD_CONTENT=$(curl -s "$DOWNLOAD_URL")
if [ -n "$DOWNLOAD_CONTENT" ]; then
  echo "   ✅ 下载成功"
  echo "   内容预览: ${DOWNLOAD_CONTENT:0:50}..."
else
  echo "   ❌ 下载失败"
fi
echo ""

# 清理
rm -f "$TEST_FILE"

echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "文件 ID (UUID): $FILE_ID"
echo "存储键: $STORAGE_KEY"
echo ""
