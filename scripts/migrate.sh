#!/bin/bash

# 数据库迁移脚本
# 用法: ./migrate.sh up|down [步数]

set -e

# 配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-assethub}"

MIGRATIONS_DIR="$(cd "$(dirname "$0")/migrations" && pwd)"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查 psql 是否安装
if ! command -v psql &> /dev/null; then
    error "psql 未安装，请先安装 PostgreSQL 客户端"
    exit 1
fi

# 执行 SQL 文件
execute_sql() {
    local file=$1
    info "执行迁移: $(basename "$file")"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
}

# 向上迁移
migrate_up() {
    local steps=${1:-all}
    info "开始向上迁移..."

    local count=0
    for file in "$MIGRATIONS_DIR"/*.up.sql; do
        if [ -f "$file" ]; then
            execute_sql "$file"
            count=$((count + 1))

            if [ "$steps" != "all" ] && [ "$count" -ge "$steps" ]; then
                break
            fi
        fi
    done

    info "迁移完成，共执行 $count 个文件"
}

# 向下迁移
migrate_down() {
    local steps=${1:-1}
    warn "开始向下迁移（回滚）..."

    local count=0
    local files=($(ls -r "$MIGRATIONS_DIR"/*.down.sql 2>/dev/null))

    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            execute_sql "$file"
            count=$((count + 1))

            if [ "$count" -ge "$steps" ]; then
                break
            fi
        fi
    done

    info "回滚完成，共执行 $count 个文件"
}

# 主逻辑
case "${1:-}" in
    up)
        migrate_up "${2:-all}"
        ;;
    down)
        migrate_down "${2:-1}"
        ;;
    *)
        echo "用法: $0 {up|down} [步数]"
        echo ""
        echo "示例:"
        echo "  $0 up          # 执行所有向上迁移"
        echo "  $0 up 1        # 执行 1 个向上迁移"
        echo "  $0 down        # 回滚 1 个迁移"
        echo "  $0 down 2      # 回滚 2 个迁移"
        exit 1
        ;;
esac
