#!/bin/bash

# E2E 测试脚本 - 统一文件存储服务

set -e

BASE_URL="http://localhost:8003"
TEST_FILE="test-upload.txt"
LARGE_FILE="test-large.txt"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

echo_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# 创建测试文件
create_test_files() {
    echo_info "创建测试文件..."
    echo "This is a test file for E2E testing" > $TEST_FILE

    # 创建 10MB 的大文件用于测试
    dd if=/dev/zero of=$LARGE_FILE bs=1m count=10 2>/dev/null
    echo_success "测试文件创建完成"
}

# 清理测试文件
cleanup() {
    echo_info "清理测试文件..."
    rm -f $TEST_FILE $LARGE_FILE
    echo_success "清理完成"
}

# 测试 1: 健康检查
test_health_check() {
    echo_info "测试 1: 健康检查"
    response=$(curl -s $BASE_URL/health)

    if echo "$response" | grep -q "ok"; then
        echo_success "健康检查通过"
        echo "响应: $response"
    else
        echo_error "健康检查失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 2: 小文件直接上传
test_direct_upload() {
    echo_info "测试 2: 小文件直接上传"

    response=$(curl -s -X POST $BASE_URL/api/v1/files/upload \
        -F "name=test.txt" \
        -F "content_type=text/plain" \
        -F "file=@$TEST_FILE")

    if echo "$response" | grep -q "file_id"; then
        FILE_ID=$(echo "$response" | grep -o '"file_id":[0-9]*' | grep -o '[0-9]*')
        echo_success "文件上传成功，文件 ID: $FILE_ID"
        echo "响应: $response"
        echo "$FILE_ID" > /tmp/file_id.txt
    else
        echo_error "文件上传失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 3: 获取文件信息
test_get_file() {
    echo_info "测试 3: 获取文件信息"

    FILE_ID=$(cat /tmp/file_id.txt)
    response=$(curl -s $BASE_URL/api/v1/files/$FILE_ID)

    if echo "$response" | grep -q "file_id"; then
        echo_success "获取文件信息成功"
        echo "响应: $response"
    else
        echo_error "获取文件信息失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 4: 获取下载 URL
test_get_download_url() {
    echo_info "测试 4: 获取下载 URL"

    FILE_ID=$(cat /tmp/file_id.txt)
    response=$(curl -s $BASE_URL/api/v1/files/$FILE_ID/download-url)

    if echo "$response" | grep -q "download_url"; then
        DOWNLOAD_URL=$(echo "$response" | grep -o '"download_url":"[^"]*"' | cut -d'"' -f4)
        echo_success "获取下载 URL 成功"
        echo "下载 URL: $DOWNLOAD_URL"

        # 测试下载
        echo_info "测试下载文件..."
        curl -s -o /tmp/downloaded.txt "$DOWNLOAD_URL"

        if [ -f /tmp/downloaded.txt ]; then
            echo_success "文件下载成功"
            echo "下载内容: $(cat /tmp/downloaded.txt)"
            rm -f /tmp/downloaded.txt
        else
            echo_error "文件下载失败"
            exit 1
        fi
    else
        echo_error "获取下载 URL 失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 5: 预签名上传
test_presigned_upload() {
    echo_info "测试 5: 预签名上传"

    # 初始化预签名上传
    response=$(curl -s -X POST $BASE_URL/api/v1/files/upload/presigned \
        -H "Content-Type: application/json" \
        -d '{
            "name": "presigned-test.txt",
            "content_type": "text/plain",
            "size": 100
        }')

    if echo "$response" | grep -q "upload_url"; then
        FILE_ID=$(echo "$response" | grep -o '"file_id":[0-9]*' | grep -o '[0-9]*')
        UPLOAD_URL=$(echo "$response" | grep -o '"upload_url":"[^"]*"' | cut -d'"' -f4)
        echo_success "预签名 URL 生成成功，文件 ID: $FILE_ID"

        # 上传到 S3
        echo_info "上传文件到 S3..."
        curl -s -X PUT "$UPLOAD_URL" \
            -H "Content-Type: text/plain" \
            --data-binary "@$TEST_FILE" > /dev/null

        echo_success "文件上传到 S3 成功"

        # 确认上传
        echo_info "确认上传..."
        response=$(curl -s -X POST $BASE_URL/api/v1/files/upload/confirm \
            -H "Content-Type: application/json" \
            -d "{\"file_id\": $FILE_ID}")

        if echo "$response" | grep -q "completed"; then
            echo_success "上传确认成功"
            echo "$FILE_ID" > /tmp/presigned_file_id.txt
        else
            echo_error "上传确认失败"
            echo "响应: $response"
        fi
    else
        echo_error "预签名 URL 生成失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 6: 分片上传
test_multipart_upload() {
    echo_info "测试 6: 分片上传"

    # 初始化分片上传
    response=$(curl -s -X POST $BASE_URL/api/v1/files/upload/multipart/init \
        -H "Content-Type: application/json" \
        -d '{
            "name": "large-file.txt",
            "content_type": "text/plain",
            "size": 10485760
        }')

    if echo "$response" | grep -q "upload_id"; then
        FILE_ID=$(echo "$response" | grep -o '"file_id":[0-9]*' | grep -o '[0-9]*')
        UPLOAD_ID=$(echo "$response" | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)
        echo_success "分片上传初始化成功，文件 ID: $FILE_ID, Upload ID: $UPLOAD_ID"

        # 生成分片 URL
        echo_info "生成分片 1 的上传 URL..."
        response=$(curl -s -X POST $BASE_URL/api/v1/files/upload/multipart/part-url \
            -H "Content-Type: application/json" \
            -d "{\"file_id\": $FILE_ID, \"part_number\": 1}")

        if echo "$response" | grep -q "upload_url"; then
            PART_URL=$(echo "$response" | grep -o '"upload_url":"[^"]*"' | cut -d'"' -f4)
            echo_success "分片 URL 生成成功"

            # 上传分片
            echo_info "上传分片 1..."
            ETAG=$(curl -s -X PUT "$PART_URL" \
                --data-binary "@$LARGE_FILE" \
                -D - | grep -i "etag:" | cut -d' ' -f2 | tr -d '\r')

            echo_success "分片上传成功，ETag: $ETAG"

            # 完成分片上传
            echo_info "完成分片上传..."
            response=$(curl -s -X POST $BASE_URL/api/v1/files/upload/multipart/complete \
                -H "Content-Type: application/json" \
                -d "{
                    \"file_id\": $FILE_ID,
                    \"parts\": [
                        {\"part_number\": 1, \"etag\": $ETAG}
                    ]
                }")

            if echo "$response" | grep -q "completed"; then
                echo_success "分片上传完成"
                echo "$FILE_ID" > /tmp/multipart_file_id.txt
            else
                echo_error "分片上传完成失败"
                echo "响应: $response"
            fi
        else
            echo_error "分片 URL 生成失败"
            echo "响应: $response"
        fi
    else
        echo_error "分片上传初始化失败"
        echo "响应: $response"
        exit 1
    fi
}

# 测试 7: 删除文件
test_delete_file() {
    echo_info "测试 7: 删除文件"

    FILE_ID=$(cat /tmp/file_id.txt)
    response=$(curl -s -X DELETE $BASE_URL/api/v1/files/$FILE_ID)

    if echo "$response" | grep -q "deleted successfully"; then
        echo_success "文件删除成功"
        echo "响应: $response"
    else
        echo_error "文件删除失败"
        echo "响应: $response"
        exit 1
    fi
}

# 主测试流程
main() {
    echo "=========================================="
    echo "统一文件存储服务 E2E 测试"
    echo "=========================================="
    echo ""

    # 创建测试文件
    create_test_files
    echo ""

    # 运行测试
    test_health_check
    echo ""

    test_direct_upload
    echo ""

    test_get_file
    echo ""

    test_get_download_url
    echo ""

    test_presigned_upload
    echo ""

    test_multipart_upload
    echo ""

    test_delete_file
    echo ""

    # 清理
    cleanup
    echo ""

    echo "=========================================="
    echo_success "所有测试通过！"
    echo "=========================================="
}

# 捕获错误并清理
trap cleanup EXIT

# 运行测试
main
