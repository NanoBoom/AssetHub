package services

import (
	"context"
	"testing"

	"github.com/NanoBoom/asethub/internal/models"
)

// MockFileRepository 用于测试的 Mock 实现
type MockFileRepository struct {
	files  map[uint]*models.File
	nextID uint
}

func NewMockFileRepository() *MockFileRepository {
	return &MockFileRepository{
		files:  make(map[uint]*models.File),
		nextID: 1,
	}
}

func (m *MockFileRepository) Create(ctx context.Context, file *models.File) error {
	file.ID = m.nextID
	m.nextID++
	m.files[file.ID] = file
	return nil
}

func (m *MockFileRepository) GetByID(ctx context.Context, id uint) (*models.File, error) {
	if file, ok := m.files[id]; ok {
		return file, nil
	}
	return nil, nil
}

func (m *MockFileRepository) GetByStorageKey(ctx context.Context, storageKey string) (*models.File, error) {
	for _, file := range m.files {
		if file.StorageKey == storageKey {
			return file, nil
		}
	}
	return nil, nil
}

func (m *MockFileRepository) Update(ctx context.Context, file *models.File) error {
	m.files[file.ID] = file
	return nil
}

func (m *MockFileRepository) UpdateStatus(ctx context.Context, id uint, status models.FileStatus) error {
	if file, ok := m.files[id]; ok {
		file.Status = status
		return nil
	}
	return nil
}

func (m *MockFileRepository) Delete(ctx context.Context, id uint) error {
	delete(m.files, id)
	return nil
}

func (m *MockFileRepository) List(ctx context.Context, offset, limit int) ([]*models.File, int64, error) {
	var files []*models.File
	for _, file := range m.files {
		files = append(files, file)
	}
	return files, int64(len(files)), nil
}

// TestMockFileRepository 测试 Mock Repository 基本功能
func TestMockFileRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewMockFileRepository()

	// 测试创建
	file := &models.File{
		Name:        "test.txt",
		Size:        100,
		ContentType: "text/plain",
		StorageKey:  "files/test.txt",
		Status:      models.FileStatusPending,
	}

	err := repo.Create(ctx, file)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if file.ID == 0 {
		t.Fatalf("File ID not set")
	}

	// 测试查询
	found, err := repo.GetByID(ctx, file.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != file.Name {
		t.Fatalf("Name mismatch: got %s, want %s", found.Name, file.Name)
	}

	// 测试更新
	file.Status = models.FileStatusCompleted
	err = repo.Update(ctx, file)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ = repo.GetByID(ctx, file.ID)
	if found.Status != models.FileStatusCompleted {
		t.Fatalf("Status not updated")
	}

	// 测试删除
	err = repo.Delete(ctx, file.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, _ = repo.GetByID(ctx, file.ID)
	if found != nil {
		t.Fatalf("File not deleted")
	}
}

// 注意：完整的 FileService 测试需要 mock gorm.DB
// 这需要使用 sqlmock 或类似工具，暂时跳过
// 在阶段 5（测试与文档）中会添加完整的集成测试
