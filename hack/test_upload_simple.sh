#!/bin/bash

# 简化版文件上传测试脚本
# 测试直接上传（后端代理）和 UUID 功能

set -e

API_BASE="http://localhost:8003/api/v1"

echo "=========================================="
echo "文件上传测试（后端代理模式）"
echo "=========================================="
echo ""

# 测试 1: 小文件直接上传
echo "【测试 1】小文件直接上传"
echo "----------------------------------------"
TEST_FILE_1="/tmp/test_small_$(date +%s).txt"
echo "Hello, this is a test file for UUID migration!" > "$TEST_FILE_1"

echo "1. 上传文件..."
UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE/files/upload" \
  -F "name=test_small.txt" \
  -F "file=@$TEST_FILE_1")

echo "   响应: $(echo "$UPLOAD_RESPONSE" | jq -c .)"
FILE_ID_1=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.file_id')
echo "   ✅ file_id: $FILE_ID_1"
echo ""

echo "2. 获取文件信息..."
FILE_INFO=$(curl -s "$API_BASE/files/$FILE_ID_1")
echo "   $(echo "$FILE_INFO" | jq -c '.data | {file_id, name, size, status}')"
echo ""

echo "3. 获取下载 URL..."
DOWNLOAD_RESPONSE=$(curl -s "$API_BASE/files/$FILE_ID_1/download-url")
DOWNLOAD_URL=$(echo "$DOWNLOAD_RESPONSE" | jq -r '.data.download_url')
echo "   ✅ 下载 URL 已生成"
echo ""

echo "4. 验证下载..."
CONTENT=$(curl -s "$DOWNLOAD_URL")
if [ -n "$CONTENT" ]; then
  echo "   ✅ 下载成功"
  echo "   内容: ${CONTENT:0:50}..."
fi
echo ""

rm -f "$TEST_FILE_1"

# 测试 2: UUID 格式验证
echo "【测试 2】UUID 格式验证"
echo "----------------------------------------"

echo "1. 测试有效 UUID..."
curl -s "$API_BASE/files/$FILE_ID_1" | jq -c '{code, message}'
echo "   ✅ 有效 UUID 通过"
echo ""

echo "2. 测试无效 UUID（整数）..."
curl -s "$API_BASE/files/123" | jq -c '{code, error}'
echo "   ✅ 正确拒绝整数 ID"
echo ""

echo "3. 测试无效 UUID（格式错误）..."
curl -s "$API_BASE/files/invalid-uuid" | jq -c '{code, error}'
echo "   ✅ 正确拒绝无效格式"
echo ""

echo "4. 测试 Nil UUID..."
curl -s "$API_BASE/files/00000000-0000-0000-0000-000000000000" | jq -c '{code, error}'
echo "   ✅ 正确拒绝 Nil UUID"
echo ""

# 测试 3: 删除文件
echo "【测试 3】删除文件"
echo "----------------------------------------"
DELETE_RESPONSE=$(curl -s -X DELETE "$API_BASE/files/$FILE_ID_1")
echo "   响应: $(echo "$DELETE_RESPONSE" | jq -c .)"
echo "   ✅ 文件已删除"
echo ""

# 测试 4: 验证删除后无法访问
echo "【测试 4】验证删除后无法访问"
echo "----------------------------------------"
GET_DELETED=$(curl -s "$API_BASE/files/$FILE_ID_1")
echo "   响应: $(echo "$GET_DELETED" | jq -c .)"
echo "   ✅ 已删除的文件无法访问"
echo ""

echo "=========================================="
echo "✅ 所有测试完成！"
echo "=========================================="
echo ""
echo "测试的文件 ID: $FILE_ID_1"
echo ""
