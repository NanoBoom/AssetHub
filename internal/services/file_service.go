package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/NanoBoom/asethub/internal/models"
	"github.com/NanoBoom/asethub/internal/repositories"
	"github.com/NanoBoom/asethub/pkg/storage"
	"gorm.io/gorm"
)

// FileService 文件服务接口
type FileService interface {
	// UploadDirect 直接上传小文件（后端代理）
	UploadDirect(ctx context.Context, name string, contentType string, size int64, reader io.Reader) (*models.File, error)

	// InitPresignedUpload 生成小文件上传预签名 URL
	InitPresignedUpload(ctx context.Context, name string, contentType string, size int64) (*PresignedUploadResult, error)

	// ConfirmUpload 确认前端直传完成
	ConfirmUpload(ctx context.Context, fileID uint) (*models.File, error)

	// InitMultipartUpload 初始化大文件分片上传
	InitMultipartUpload(ctx context.Context, name string, contentType string, size int64) (*MultipartUploadResult, error)

	// GeneratePartUploadURL 生成分片上传预签名 URL
	GeneratePartUploadURL(ctx context.Context, fileID uint, partNumber int) (string, error)

	// CompleteMultipartUpload 完成大文件分片上传
	CompleteMultipartUpload(ctx context.Context, fileID uint, parts []storage.CompletedPart) (*models.File, error)

	// GetDownloadURL 生成下载预签名 URL
	GetDownloadURL(ctx context.Context, fileID uint, expiry time.Duration) (string, error)

	// DeleteFile 删除文件（S3 + 数据库）
	DeleteFile(ctx context.Context, fileID uint) error

	// GetFile 获取文件信息
	GetFile(ctx context.Context, fileID uint) (*models.File, error)

	// ListFiles 分页查询文件列表
	ListFiles(ctx context.Context, offset, limit int) ([]*models.File, int64, error)
}

// PresignedUploadResult 预签名上传结果
type PresignedUploadResult struct {
	FileID     uint   `json:"file_id"`
	UploadURL  string `json:"upload_url"`
	StorageKey string `json:"storage_key"`
	ExpiresIn  int64  `json:"expires_in"` // 秒
}

// MultipartUploadResult 分片上传初始化结果
type MultipartUploadResult struct {
	FileID     uint   `json:"file_id"`
	UploadID   string `json:"upload_id"`
	StorageKey string `json:"storage_key"`
}

// fileService 文件服务实现
type fileService struct {
	fileRepo repositories.FileRepository
	storage  storage.Storage
	db       *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService(fileRepo repositories.FileRepository, storage storage.Storage, db *gorm.DB) FileService {
	return &fileService{
		fileRepo: fileRepo,
		storage:  storage,
		db:       db,
	}
}

// generateStorageKey 生成存储键
func (s *fileService) generateStorageKey(name string) string {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(name)
	return fmt.Sprintf("files/%d/%s%s", timestamp, fmt.Sprintf("%d", timestamp), ext)
}

// UploadDirect 直接上传小文件（后端代理）
func (s *fileService) UploadDirect(ctx context.Context, name string, contentType string, size int64, reader io.Reader) (*models.File, error) {
	// 生成存储键
	storageKey := s.generateStorageKey(name)

	// 创建文件记录
	file := &models.File{
		Name:        name,
		Size:        size,
		ContentType: contentType,
		StorageKey:  storageKey,
		Status:      models.FileStatusPending,
	}

	// 开启事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建数据库记录
	if err := tx.Create(file).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// 上传到 S3
	if err := s.storage.Upload(ctx, storageKey, reader, size); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// 更新状态为已完成
	file.Status = models.FileStatusCompleted
	if err := tx.Save(file).Error; err != nil {
		tx.Rollback()
		// 尝试删除 S3 文件
		_ = s.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return file, nil
}

// InitPresignedUpload 生成小文件上传预签名 URL
func (s *fileService) InitPresignedUpload(ctx context.Context, name string, contentType string, size int64) (*PresignedUploadResult, error) {
	// 生成存储键
	storageKey := s.generateStorageKey(name)

	// 创建文件记录
	file := &models.File{
		Name:        name,
		Size:        size,
		ContentType: contentType,
		StorageKey:  storageKey,
		Status:      models.FileStatusPending,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// 生成预签名 URL（1 小时有效期）
	expiry := 1 * time.Hour
	uploadURL, err := s.storage.GeneratePresignedUploadURL(ctx, storageKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &PresignedUploadResult{
		FileID:     file.ID,
		UploadURL:  uploadURL,
		StorageKey: storageKey,
		ExpiresIn:  int64(expiry.Seconds()),
	}, nil
}

// ConfirmUpload 确认前端直传完成
func (s *fileService) ConfirmUpload(ctx context.Context, fileID uint) (*models.File, error) {
	// 查询文件记录
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 更新状态为已完成
	file.Status = models.FileStatusCompleted
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	return file, nil
}

// InitMultipartUpload 初始化大文件分片上传
func (s *fileService) InitMultipartUpload(ctx context.Context, name string, contentType string, size int64) (*MultipartUploadResult, error) {
	// 生成存储键
	storageKey := s.generateStorageKey(name)

	// 初始化 S3 分片上传
	multipartUpload, err := s.storage.InitMultipartUpload(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init multipart upload: %w", err)
	}

	// 创建文件记录
	file := &models.File{
		Name:        name,
		Size:        size,
		ContentType: contentType,
		StorageKey:  storageKey,
		Status:      models.FileStatusUploading,
		UploadID:    multipartUpload.UploadID,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return &MultipartUploadResult{
		FileID:     file.ID,
		UploadID:   multipartUpload.UploadID,
		StorageKey: storageKey,
	}, nil
}

// GeneratePartUploadURL 生成分片上传预签名 URL
func (s *fileService) GeneratePartUploadURL(ctx context.Context, fileID uint, partNumber int) (string, error) {
	// 查询文件记录
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	if file.UploadID == "" {
		return "", fmt.Errorf("file is not in multipart upload mode")
	}

	// 生成分片预签名 URL（1 小时有效期）
	expiry := 1 * time.Hour
	partURL, err := s.storage.GeneratePresignedPartURL(ctx, file.StorageKey, file.UploadID, partNumber, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate part URL: %w", err)
	}

	return partURL, nil
}

// CompleteMultipartUpload 完成大文件分片上传
func (s *fileService) CompleteMultipartUpload(ctx context.Context, fileID uint, parts []storage.CompletedPart) (*models.File, error) {
	// 查询文件记录
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if file.UploadID == "" {
		return nil, fmt.Errorf("file is not in multipart upload mode")
	}

	// 完成 S3 分片上传
	if err := s.storage.CompleteMultipartUpload(ctx, file.StorageKey, file.UploadID, parts); err != nil {
		return nil, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	// 更新状态为已完成
	file.Status = models.FileStatusCompleted
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	return file, nil
}

// GetDownloadURL 生成下载预签名 URL
func (s *fileService) GetDownloadURL(ctx context.Context, fileID uint, expiry time.Duration) (string, error) {
	// 查询文件记录
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	if file.Status != models.FileStatusCompleted {
		return "", fmt.Errorf("file is not ready for download")
	}

	// 生成下载预签名 URL
	downloadURL, err := s.storage.GeneratePresignedDownloadURL(ctx, file.StorageKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return downloadURL, nil
}

// DeleteFile 删除文件（S3 + 数据库）
func (s *fileService) DeleteFile(ctx context.Context, fileID uint) error {
	// 查询文件记录
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 开启事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除数据库记录
	if err := tx.Delete(file).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// 删除 S3 文件
	if err := s.storage.Delete(ctx, file.StorageKey); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetFile 获取文件信息
func (s *fileService) GetFile(ctx context.Context, fileID uint) (*models.File, error) {
	return s.fileRepo.GetByID(ctx, fileID)
}

// ListFiles 分页查询文件列表
func (s *fileService) ListFiles(ctx context.Context, offset, limit int) ([]*models.File, int64, error) {
	return s.fileRepo.List(ctx, offset, limit)
}
