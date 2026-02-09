package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Config S3 配置
type S3Config struct {
	Region          string // AWS 区域
	Bucket          string // S3 存储桶名称
	AccessKeyID     string // AWS Access Key ID（可选，留空使用 IAM 角色）
	SecretAccessKey string // AWS Secret Access Key（可选，留空使用 IAM 角色）
	Endpoint        string // 自定义端点（用于 MinIO 等 S3 兼容服务）
	UsePathStyle    bool   // 是否使用路径风格（MinIO 需要设为 true）
}

// S3Storage S3 存储实现
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage 创建 S3 存储实例
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	var awsCfg aws.Config
	var err error

	// 配置选项
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	// 如果提供了访问密钥，使用静态凭证
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// 加载配置
	awsCfg, err = config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建 S3 客户端
	clientOpts := []func(*s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = cfg.UsePathStyle
		},
	}

	// 如果提供了自定义端点
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOpts...)

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload 直接上传文件（后端代理）
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	// 设置 Content-Type
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// GeneratePresignedUploadURL 生成小文件上传预签名 URL（前端直传）
func (s *S3Storage) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration, contentType string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// 设置 Content-Type（必须与实际上传时一致，否则 S3 V4 签名验证失败）
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	req, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return req.URL, nil
}

// InitMultipartUpload 初始化分片上传
func (s *S3Storage) InitMultipartUpload(ctx context.Context, key string, contentType string) (*MultipartUpload, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// 设置 Content-Type（必须在初始化时设置）
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	output, err := s.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to init multipart upload: %w", err)
	}

	return &MultipartUpload{
		UploadID: aws.ToString(output.UploadId),
		Key:      key,
		Parts:    []string{}, // 预签名 URL 需要按需生成
	}, nil
}

// GeneratePresignedPartURL 生成分片上传预签名 URL
func (s *S3Storage) GeneratePresignedPartURL(ctx context.Context, key string, uploadID string, partNumber int, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(partNumber)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned part URL: %w", err)
	}

	return req.URL, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *S3Storage) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []CompletedPart) error {
	// 转换为 S3 类型
	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(part.PartNumber)),
			ETag:       aws.String(part.ETag),
		}
	}

	_, err := s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// GeneratePresignedDownloadURL 生成下载预签名 URL
func (s *S3Storage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, opts *PresignOptions) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// 设置响应头选项
	if opts != nil {
		if opts.ContentType != "" {
			input.ResponseContentType = aws.String(opts.ContentType)
		}
		if opts.ContentDisposition != "" {
			input.ResponseContentDisposition = aws.String(opts.ContentDisposition)
		}
	}

	req, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return req.URL, nil
}

// Delete 删除对象
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
