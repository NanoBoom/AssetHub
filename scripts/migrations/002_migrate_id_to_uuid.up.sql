-- 迁移文件 ID 从 SERIAL 到 UUID
-- 警告：这是破坏性变更，会删除所有现有数据

BEGIN;

-- 1. 启用 UUID 扩展（PostgreSQL 13+ 默认启用）
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2. 删除旧表（如果需要保留数据，请先手动备份）
DROP TABLE IF EXISTS files CASCADE;

-- 3. 重新创建表，使用 UUID 作为主键（应用层生成 UUID，不使用数据库默认值）
CREATE TABLE files (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    content_type VARCHAR(100),
    storage_key VARCHAR(500) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    hash VARCHAR(64),
    upload_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 4. 重新创建索引
CREATE INDEX idx_files_name ON files(name);
CREATE INDEX idx_files_status ON files(status);
CREATE INDEX idx_files_hash ON files(hash);
CREATE INDEX idx_files_deleted_at ON files(deleted_at);

-- 5. 添加注释
COMMENT ON TABLE files IS '文件元数据表';
COMMENT ON COLUMN files.id IS '主键ID（UUID v4，应用层生成）';
COMMENT ON COLUMN files.name IS '文件名';
COMMENT ON COLUMN files.size IS '文件大小（字节）';
COMMENT ON COLUMN files.content_type IS 'MIME类型';
COMMENT ON COLUMN files.storage_key IS '存储键（S3对象键）';
COMMENT ON COLUMN files.status IS '上传状态: pending, uploading, completed, failed';
COMMENT ON COLUMN files.hash IS '文件哈希值（SHA256）';
COMMENT ON COLUMN files.upload_id IS '分片上传ID（仅分片上传时使用）';
COMMENT ON COLUMN files.created_at IS '创建时间';
COMMENT ON COLUMN files.updated_at IS '更新时间';
COMMENT ON COLUMN files.deleted_at IS '软删除时间';

COMMIT;
