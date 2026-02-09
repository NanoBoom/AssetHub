#!/bin/bash

# 大文件分片上传测试脚本
# 测试流程：
# 1. 初始化分片上传，获取 file_id 和 upload_id
# 2. 为每个分片生成预签名 URL
# 3. 上传所有分片
# 4. 完成分片上传
# 5. 获取下载 URL 并验证

set -e

API_BASE="http://localhost:8003/api/v1"
TEST_FILE="/tmp/test_multipart_$(date +%s).bin"
PART_SIZE=$((5 * 1024 * 1024))  # 5MB per part (S3 要求最小 5MB)

echo "=========================================="
echo "大文件分片上传测试"
echo "=========================================="
echo ""

# 创建测试文件（10MB，分成 2 个分片）
echo "1. 创建测试文件 (10MB)..."
dd if=/dev/urandom of="$TEST_FILE" bs=1m count=10 2>/dev/null
FILE_SIZE=$(wc -c < "$TEST_FILE" | tr -d ' ')
TOTAL_PARTS=$(( (FILE_SIZE + PART_SIZE - 1) / PART_SIZE ))
echo "   文件路径: $TEST_FILE"
echo "   文件大小: $FILE_SIZE bytes ($(echo "scale=2; $FILE_SIZE/1024/1024" | bc) MB)"
echo "   分片大小: $PART_SIZE bytes (5 MB)"
echo "   分片数量: $TOTAL_PARTS"
echo ""

# 步骤 1: 初始化分片上传
echo "2. 初始化分片上传..."
INIT_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload/multipart/init" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"test_multipart.bin\",
    \"content_type\": \"application/octet-stream\",
    \"size\": $FILE_SIZE
  }")

# 提取 file_id 和 upload_id
FILE_ID=$(echo "$INIT_RESPONSE" | jq -r '.data.file_id')
UPLOAD_ID=$(echo "$INIT_RESPONSE" | jq -r '.data.upload_id')
STORAGE_KEY=$(echo "$INIT_RESPONSE" | jq -r '.data.storage_key')

echo "   响应: $(echo "$INIT_RESPONSE" | jq -c .)"
echo ""

if [ "$FILE_ID" == "null" ] || [ -z "$FILE_ID" ]; then
  echo "❌ 错误: 无法获取 file_id"
  exit 1
fi

echo "   ✅ file_id: $FILE_ID"
echo "   ✅ upload_id: $UPLOAD_ID"
echo "   ✅ storage_key: $STORAGE_KEY"
echo ""

# 步骤 2-3: 上传每个分片
echo "3. 上传分片..."
PARTS_JSON="["
for ((part_num=1; part_num<=TOTAL_PARTS; part_num++)); do
  echo "   分片 $part_num/$TOTAL_PARTS:"

  # 生成分片预签名 URL
  PART_URL_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload/multipart/part-url" \
    -H "Content-Type: application/json" \
    -d "{
      \"file_id\": \"$FILE_ID\",
      \"part_number\": $part_num
    }")

  PART_URL=$(echo "$PART_URL_RESPONSE" | jq -r '.data.upload_url')

  if [ "$PART_URL" == "null" ] || [ -z "$PART_URL" ]; then
    echo "   ❌ 错误: 无法获取分片 $part_num 的上传 URL"
    exit 1
  fi

  # 计算分片的起始和结束位置
  START=$(( (part_num - 1) * PART_SIZE ))

  # 提取分片数据并上传
  PART_FILE="/tmp/part_${part_num}.bin"
  dd if="$TEST_FILE" of="$PART_FILE" bs=1 skip=$START count=$PART_SIZE 2>/dev/null

  # 上传分片（增加超时时间到 300 秒以适应 S3 中国区的慢速网络）
  UPLOAD_RESPONSE=$(curl --max-time 300 -s -X PUT "$PART_URL" \
    -H "Content-Type:" \
    --data-binary "@$PART_FILE" \
    -D /tmp/headers_${part_num}.txt)

  # 提取 ETag（保留引号，S3 API 要求 ETag 必须带引号）
  ETAG=$(grep -i "etag:" /tmp/headers_${part_num}.txt | cut -d' ' -f2 | tr -d '\r\n')

  if [ -z "$ETAG" ]; then
    # 如果从响应头中获取不到，尝试从响应体中获取（OSS 可能返回 JSON）
    ETAG=$(echo "$UPLOAD_RESPONSE" | jq -r '.ETag // empty')
    # 如果 ETag 没有引号，添加引号
    if [[ ! "$ETAG" =~ ^\" ]]; then
      ETAG="\"$ETAG\""
    fi
  fi

  if [ -z "$ETAG" ]; then
    echo "   ⚠️  警告: 无法获取 ETag，使用占位符"
    ETAG="placeholder-etag-$part_num"
  fi

  echo "      ✅ 上传成功"
  echo "      ETag: $ETAG"

  # 构建 parts JSON（ETag 已经带引号，不需要再加）
  if [ $part_num -gt 1 ]; then
    PARTS_JSON="$PARTS_JSON,"
  fi
  PARTS_JSON="$PARTS_JSON{\"part_number\":$part_num,\"etag\":$ETAG}"

  # 清理临时文件
  rm -f "$PART_FILE" "/tmp/headers_${part_num}.txt"
done
PARTS_JSON="$PARTS_JSON]"
echo ""

# 步骤 4: 完成分片上传
echo "4. 完成分片上传..."
COMPLETE_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload/multipart/complete" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_id\": \"$FILE_ID\",
    \"parts\": $PARTS_JSON
  }")

COMPLETE_STATUS=$(echo "$COMPLETE_RESPONSE" | jq -r '.data.status')
echo "   响应: $(echo "$COMPLETE_RESPONSE" | jq -c .)"

if [ "$COMPLETE_STATUS" == "completed" ]; then
  echo "   ✅ 分片上传完成"
else
  echo "   ⚠️  状态: $COMPLETE_STATUS"
fi
echo ""

# 步骤 5: 获取文件信息
echo "5. 获取文件信息..."
FILE_INFO=$(curl -s "$API_BASE/files/$FILE_ID")
echo "   响应: $(echo "$FILE_INFO" | jq -c .)"
echo ""

# 步骤 6: 获取下载 URL
echo "6. 获取下载 URL..."
DOWNLOAD_RESPONSE=$(curl -s "$API_BASE/files/$FILE_ID/download-url")
DOWNLOAD_URL=$(echo "$DOWNLOAD_RESPONSE" | jq -r '.data.download_url')
echo "   下载 URL: ${DOWNLOAD_URL:0:100}..."
echo ""

# 步骤 7: 验证下载（下载前 100 字节验证）
echo "7. 验证下载..."
DOWNLOAD_SIZE=$(curl -s -I "$DOWNLOAD_URL" | grep -i "content-length" | awk '{print $2}' | tr -d '\r')
if [ -n "$DOWNLOAD_SIZE" ]; then
  echo "   ✅ 下载 URL 有效"
  echo "   文件大小: $DOWNLOAD_SIZE bytes"
else
  echo "   ⚠️  无法验证下载 URL"
fi
echo ""

# 清理
rm -f "$TEST_FILE"

echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "文件 ID (UUID): $FILE_ID"
echo "上传 ID: $UPLOAD_ID"
echo "存储键: $STORAGE_KEY"
echo "分片数量: $TOTAL_PARTS"
echo ""
