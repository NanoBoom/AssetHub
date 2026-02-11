package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string // OSS endpoint (如 oss-cn-hangzhou.aliyuncs.com)
	Bucket          string // OSS bucket 名称
	AccessKeyID     string // 阿里云 Access Key ID
	AccessKeySecret string // 阿里云 Access Key Secret
}

// OSSStorage OSS 存储实现
type OSSStorage struct {
	client *oss.Client
	bucket string
}

// NewOSSStorage 创建 OSS 存储实例
func NewOSSStorage(ctx context.Context, cfg OSSConfig) (*OSSStorage, error) {
	// 创建凭证提供者
	credentialsProvider := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret)

	// 从 endpoint 提取区域
	// endpoint 格式: oss-{region}.aliyuncs.com
	region := "cn-hangzhou" // 默认区域
	if len(cfg.Endpoint) > 4 && cfg.Endpoint[:4] == "oss-" {
		// 提取 oss- 和 .aliyuncs.com 之间的部分
		endIdx := len(cfg.Endpoint)
		if idx := strings.Index(cfg.Endpoint, ".aliyuncs.com"); idx > 0 {
			endIdx = idx
		}
		region = cfg.Endpoint[4:endIdx]
	}

	// 创建 OSS 客户端配置
	clientCfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentialsProvider).
		WithRegion(region).
		WithEndpoint(cfg.Endpoint)

	// 创建 OSS 客户端
	client := oss.NewClient(clientCfg)

	return &OSSStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload 直接上传文件（后端代理）
func (o *OSSStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
		Body:   reader,
	}

	// 设置 Content-Type（必须在上传时设置，不能在下载时覆盖）
	if contentType != "" {
		req.ContentType = oss.Ptr(contentType)
	}

	_, err := o.client.PutObject(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// GeneratePresignedUploadURL 生成小文件上传预签名 URL（前端直传）
func (o *OSSStorage) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration, contentType string) (string, error) {
	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	}

	// 设置 Content-Type（必须与实际上传时一致，否则 OSS V4 签名验证失败）
	if contentType != "" {
		req.ContentType = oss.Ptr(contentType)
	}

	result, err := o.client.Presign(ctx, req, oss.PresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return result.URL, nil
}

// InitMultipartUpload 初始化分片上传
func (o *OSSStorage) InitMultipartUpload(ctx context.Context, key string, contentType string) (*MultipartUpload, error) {
	req := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	}

	// 设置 Content-Type（必须在初始化时设置）
	if contentType != "" {
		req.ContentType = oss.Ptr(contentType)
	}

	output, err := o.client.InitiateMultipartUpload(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to init multipart upload: %w", err)
	}

	return &MultipartUpload{
		UploadID: oss.ToString(output.UploadId),
		Key:      key,
		Parts:    []string{}, // 预签名 URL 需要按需生成
	}, nil
}

// GeneratePresignedPartURL 生成分片上传预签名 URL
func (o *OSSStorage) GeneratePresignedPartURL(ctx context.Context, key string, uploadID string, partNumber int, expiry time.Duration) (string, error) {
	req, err := o.client.Presign(ctx, &oss.UploadPartRequest{
		Bucket:     oss.Ptr(o.bucket),
		Key:        oss.Ptr(key),
		UploadId:   oss.Ptr(uploadID),
		PartNumber: int32(partNumber),
	}, oss.PresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned part URL: %w", err)
	}

	return req.URL, nil
}

// CompleteMultipartUpload 完成分片上传
func (o *OSSStorage) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []CompletedPart) error {
	// 转换为 OSS 类型
	completedParts := make([]oss.UploadPart, len(parts))
	for i, part := range parts {
		completedParts[i] = oss.UploadPart{
			PartNumber: int32(part.PartNumber),
			ETag:       oss.Ptr(part.ETag),
		}
	}

	_, err := o.client.CompleteMultipartUpload(ctx, &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(o.bucket),
		Key:      oss.Ptr(key),
		UploadId: oss.Ptr(uploadID),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// GetObject 获取对象内容（流式读取）
func (o *OSSStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	req := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	}

	result, err := o.client.GetObject(ctx, req)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get object: %w", err)
	}

	// 提取 Content-Type
	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	// 提取 Content-Length
	contentLength := result.ContentLength

	return result.Body, contentType, contentLength, nil
}

// GeneratePresignedDownloadURL 生成下载预签名 URL
func (o *OSSStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, opts *PresignOptions) (string, error) {
	req := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	}

	// 只设置 Content-Disposition（OSS 不允许覆盖 Content-Type）
	if opts != nil && opts.ContentDisposition != "" {
		req.ResponseContentDisposition = oss.Ptr(opts.ContentDisposition)
	}

	result, err := o.client.Presign(ctx, req, oss.PresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return result.URL, nil
}

// Delete 删除对象
func (o *OSSStorage) Delete(ctx context.Context, key string) error {
	_, err := o.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(o.bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
