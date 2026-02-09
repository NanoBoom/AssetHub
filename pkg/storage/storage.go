package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/NanoBoom/asethub/internal/config"
)

// MultipartUpload 分片上传信息
type MultipartUpload struct {
	UploadID string   // S3 返回的 upload ID
	Key      string   // 对象键
	Parts    []string // 预签名 URL 列表（按 part number 顺序）
}

// CompletedPart 已完成的分片信息
type CompletedPart struct {
	PartNumber int    // 分片编号（从 1 开始）
	ETag       string // S3 返回的 ETag
}

// PresignOptions 预签名 URL 选项
// 用于设置下载 URL 的响应头，控制浏览器行为（预览 vs 下载）
type PresignOptions struct {
	ContentType        string // 响应 Content-Type（如 "image/png"）
	ContentDisposition string // 响应 Content-Disposition（"inline" 预览 / "attachment" 下载）
}

// Storage 统一存储接口
// 抽象云存储后端（S3/OSS/本地存储），提供文件上传、下载、预签名 URL 生成等核心能力
type Storage interface {
	// === 小文件上传（< 100MB）===

	// Upload 直接上传文件（后端代理）
	// 适用场景：小文件（< 100MB），如缩略图、字幕、海报
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键（存储路径）
	//   - reader: 文件内容流
	//   - size: 文件大小（字节）
	//   - contentType: 文件 MIME 类型（如 "image/png"）
	// 返回：错误信息
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error

	// GeneratePresignedUploadURL 生成小文件上传预签名 URL（前端直传）
	// 适用场景：前端直接上传到 S3，避免后端代理
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键（存储路径）
	//   - expiry: URL 过期时间（建议 1 小时）
	//   - contentType: 文件 MIME 类型（必须与实际上传时一致，否则签名验证失败）
	// 返回：预签名 URL、错误信息
	GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration, contentType string) (string, error)

	// === 大文件分片上传（>= 100MB）===

	// InitMultipartUpload 初始化分片上传
	// 适用场景：大文件（>= 100MB），如视频素材
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键（存储路径）
	//   - contentType: 文件 MIME 类型（如 "video/mp4"）
	// 返回：分片上传信息（包含 upload ID 和预签名 URL 列表）、错误信息
	InitMultipartUpload(ctx context.Context, key string, contentType string) (*MultipartUpload, error)

	// GeneratePresignedPartURL 生成分片上传预签名 URL
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键
	//   - uploadID: 分片上传 ID
	//   - partNumber: 分片编号（从 1 开始）
	//   - expiry: URL 过期时间
	// 返回：预签名 URL、错误信息
	GeneratePresignedPartURL(ctx context.Context, key string, uploadID string, partNumber int, expiry time.Duration) (string, error)

	// CompleteMultipartUpload 完成分片上传
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键
	//   - uploadID: 分片上传 ID
	//   - parts: 已完成的分片列表（必须按 part number 排序）
	// 返回：错误信息
	CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []CompletedPart) error

	// === 通用操作 ===

	// GeneratePresignedDownloadURL 生成下载预签名 URL
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键
	//   - expiry: URL 过期时间（建议 15 分钟）
	//   - opts: 响应头选项（可选，传 nil 使用默认行为）
	// 返回：预签名 URL、错误信息
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, opts *PresignOptions) (string, error)

	// Delete 删除对象
	// 参数：
	//   - ctx: 上下文
	//   - key: 对象键
	// 返回：错误信息
	Delete(ctx context.Context, key string) error
}

// NewStorage 根据配置创建存储实例（工厂函数）
// 这是唯一需要修改的地方，添加新存储类型时只需在此处添加 case
func NewStorage(ctx context.Context, cfg *config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "s3":
		s3Config := S3Config{
			Region:          cfg.S3.Region,
			Bucket:          cfg.S3.Bucket,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			Endpoint:        cfg.S3.Endpoint,
			UsePathStyle:    cfg.S3.UsePathStyle,
		}
		return NewS3Storage(ctx, s3Config)

	case "oss":
		ossConfig := OSSConfig{
			Endpoint:        cfg.OSS.Endpoint,
			Bucket:          cfg.OSS.Bucket,
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
		}
		return NewOSSStorage(ctx, ossConfig)

	case "local":
		// TODO: 实现本地存储
		return nil, fmt.Errorf("local storage not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported storage type: %s (supported: s3, oss, local)", cfg.Type)
	}
}
