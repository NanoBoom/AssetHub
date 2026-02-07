package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// MockOSSStorage 用于测试的 Mock 实现
type MockOSSStorage struct {
	storage map[string][]byte // 模拟存储
}

// NewMockOSSStorage 创建 Mock 存储实例
func NewMockOSSStorage() *MockOSSStorage {
	return &MockOSSStorage{
		storage: make(map[string][]byte),
	}
}

// Upload 实现 Storage 接口
func (m *MockOSSStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.storage[key] = data
	return nil
}

// GeneratePresignedUploadURL 实现 Storage 接口
func (m *MockOSSStorage) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://mock-oss.example.com/upload/" + key, nil
}

// InitMultipartUpload 实现 Storage 接口
func (m *MockOSSStorage) InitMultipartUpload(ctx context.Context, key string) (*MultipartUpload, error) {
	return &MultipartUpload{
		UploadID: "mock-upload-id-" + key,
		Key:      key,
		Parts:    []string{},
	}, nil
}

// GeneratePresignedPartURL 实现 Storage 接口
func (m *MockOSSStorage) GeneratePresignedPartURL(ctx context.Context, key string, uploadID string, partNumber int, expiry time.Duration) (string, error) {
	return "https://mock-oss.example.com/part/" + key + "?partNumber=" + string(rune(partNumber)), nil
}

// CompleteMultipartUpload 实现 Storage 接口
func (m *MockOSSStorage) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []CompletedPart) error {
	// 模拟合并分片
	m.storage[key] = []byte("multipart-upload-completed")
	return nil
}

// GeneratePresignedDownloadURL 实现 Storage 接口
func (m *MockOSSStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://mock-oss.example.com/download/" + key, nil
}

// Delete 实现 Storage 接口
func (m *MockOSSStorage) Delete(ctx context.Context, key string) error {
	delete(m.storage, key)
	return nil
}

// TestOSSUpload 测试小文件上传
func TestOSSUpload(t *testing.T) {
	ctx := context.Background()
	storage := NewMockOSSStorage()

	// 测试数据
	key := "test/file.txt"
	content := "Hello, OSS!"
	reader := strings.NewReader(content)

	// 执行上传
	err := storage.Upload(ctx, key, reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// 验证数据
	if data, ok := storage.storage[key]; !ok {
		t.Fatalf("File not found in storage")
	} else if string(data) != content {
		t.Fatalf("Content mismatch: got %s, want %s", string(data), content)
	}
}

// TestOSSGeneratePresignedUploadURL 测试生成上传预签名 URL
func TestOSSGeneratePresignedUploadURL(t *testing.T) {
	ctx := context.Background()
	storage := NewMockOSSStorage()

	key := "test/file.txt"
	expiry := 1 * time.Hour

	url, err := storage.GeneratePresignedUploadURL(ctx, key, expiry)
	if err != nil {
		t.Fatalf("GeneratePresignedUploadURL failed: %v", err)
	}

	if url == "" {
		t.Fatalf("URL is empty")
	}

	if !strings.Contains(url, key) {
		t.Fatalf("URL does not contain key: %s", url)
	}
}

// TestOSSMultipartUpload 测试分片上传流程
func TestOSSMultipartUpload(t *testing.T) {
	ctx := context.Background()
	storage := NewMockOSSStorage()

	key := "test/large-file.mp4"

	// 1. 初始化分片上传
	upload, err := storage.InitMultipartUpload(ctx, key)
	if err != nil {
		t.Fatalf("InitMultipartUpload failed: %v", err)
	}

	if upload.UploadID == "" {
		t.Fatalf("UploadID is empty")
	}

	if upload.Key != key {
		t.Fatalf("Key mismatch: got %s, want %s", upload.Key, key)
	}

	// 2. 生成分片预签名 URL
	partNumber := 1
	expiry := 1 * time.Hour
	partURL, err := storage.GeneratePresignedPartURL(ctx, key, upload.UploadID, partNumber, expiry)
	if err != nil {
		t.Fatalf("GeneratePresignedPartURL failed: %v", err)
	}

	if partURL == "" {
		t.Fatalf("Part URL is empty")
	}

	// 3. 完成分片上传
	parts := []CompletedPart{
		{PartNumber: 1, ETag: "etag1"},
		{PartNumber: 2, ETag: "etag2"},
	}

	err = storage.CompleteMultipartUpload(ctx, key, upload.UploadID, parts)
	if err != nil {
		t.Fatalf("CompleteMultipartUpload failed: %v", err)
	}

	// 验证文件已存储
	if _, ok := storage.storage[key]; !ok {
		t.Fatalf("File not found in storage after multipart upload")
	}
}

// TestOSSGeneratePresignedDownloadURL 测试生成下载预签名 URL
func TestOSSGeneratePresignedDownloadURL(t *testing.T) {
	ctx := context.Background()
	storage := NewMockOSSStorage()

	// 先上传一个文件
	key := "test/file.txt"
	content := "Hello, OSS!"
	storage.storage[key] = []byte(content)

	expiry := 15 * time.Minute
	url, err := storage.GeneratePresignedDownloadURL(ctx, key, expiry)
	if err != nil {
		t.Fatalf("GeneratePresignedDownloadURL failed: %v", err)
	}

	if url == "" {
		t.Fatalf("URL is empty")
	}

	if !strings.Contains(url, key) {
		t.Fatalf("URL does not contain key: %s", url)
	}
}

// TestOSSDelete 测试删除对象
func TestOSSDelete(t *testing.T) {
	ctx := context.Background()
	storage := NewMockOSSStorage()

	// 先上传一个文件
	key := "test/file.txt"
	content := "Hello, OSS!"
	reader := bytes.NewReader([]byte(content))
	err := storage.Upload(ctx, key, reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// 验证文件存在
	if _, ok := storage.storage[key]; !ok {
		t.Fatalf("File not found after upload")
	}

	// 删除文件
	err = storage.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证文件已删除
	if _, ok := storage.storage[key]; ok {
		t.Fatalf("File still exists after delete")
	}
}
