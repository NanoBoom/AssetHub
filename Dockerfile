# ========================================
# Stage 1: Builder
# ========================================
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 安装 swag（用于生成 Swagger 文档）
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum（利用 Docker 缓存层）
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 生成 Swagger 文档
RUN swag init -g cmd/api/main.go -o docs

# 构建应用（静态链接，减小体积）
# -s: 去除符号表
# -w: 去除 DWARF 调试信息
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o api \
    cmd/api/main.go

# ========================================
# Stage 2: Runtime
# ========================================
FROM alpine:3.19

# 安装运行时依赖
# ca-certificates: 用于 HTTPS 请求（S3/OSS SDK 需要）
# curl: 用于健康检查
RUN apk add --no-cache ca-certificates curl tzdata

# 设置时区（默认 UTC，可通过环境变量 TZ 修改）
ENV TZ=UTC

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/api .

# 复制必要的配置文件（可选，Docker 环境主要使用环境变量）
COPY configs ./configs

# 创建存储目录（用于本地存储开发）
RUN mkdir -p /app/storage && \
    chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=2s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./api"]
