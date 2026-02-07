package repositories

import (
	"context"

	"github.com/NanoBoom/asethub/internal/models"
	"gorm.io/gorm"
)

// FileRepository 文件元数据仓储接口
type FileRepository interface {
	// Create 创建文件记录
	Create(ctx context.Context, file *models.File) error

	// GetByID 根据 ID 查询文件
	GetByID(ctx context.Context, id uint) (*models.File, error)

	// GetByStorageKey 根据存储键查询文件
	GetByStorageKey(ctx context.Context, storageKey string) (*models.File, error)

	// Update 更新文件记录
	Update(ctx context.Context, file *models.File) error

	// UpdateStatus 更新文件状态
	UpdateStatus(ctx context.Context, id uint, status models.FileStatus) error

	// Delete 删除文件记录（软删除）
	Delete(ctx context.Context, id uint) error

	// List 分页查询文件列表
	List(ctx context.Context, offset, limit int) ([]*models.File, int64, error)
}

// fileRepository 文件仓储实现
type fileRepository struct {
	*BaseRepository
}

// NewFileRepository 创建文件仓储实例
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建文件记录
func (r *fileRepository) Create(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// GetByID 根据 ID 查询文件
func (r *fileRepository) GetByID(ctx context.Context, id uint) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByStorageKey 根据存储键查询文件
func (r *fileRepository) GetByStorageKey(ctx context.Context, storageKey string) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).Where("storage_key = ?", storageKey).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Update 更新文件记录
func (r *fileRepository) Update(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// UpdateStatus 更新文件状态
func (r *fileRepository) UpdateStatus(ctx context.Context, id uint, status models.FileStatus) error {
	return r.db.WithContext(ctx).Model(&models.File{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除文件记录（软删除）
func (r *fileRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.File{}, id).Error
}

// List 分页查询文件列表
func (r *fileRepository) List(ctx context.Context, offset, limit int) ([]*models.File, int64, error) {
	var files []*models.File
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&models.File{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
