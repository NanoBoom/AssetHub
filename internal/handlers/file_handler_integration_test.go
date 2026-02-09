package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/NanoBoom/asethub/internal/config"
	"github.com/NanoBoom/asethub/internal/handlers"
	"github.com/NanoBoom/asethub/internal/middleware"
	"github.com/NanoBoom/asethub/internal/models"
	"github.com/NanoBoom/asethub/internal/repositories"
	"github.com/NanoBoom/asethub/internal/services"
	"github.com/NanoBoom/asethub/pkg/response"
	"github.com/NanoBoom/asethub/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockStorage Mock 存储实现（用于测试）
type MockStorage struct {
	files map[string][]byte // 模拟存储的文件
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

func (m *MockStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.files[key] = data
	return nil
}

func (m *MockStorage) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("https://mock-s3.example.com/upload/%s", key), nil
}

func (m *MockStorage) InitMultipartUpload(ctx context.Context, key string, contentType string) (*storage.MultipartUpload, error) {
	return &storage.MultipartUpload{
		UploadID: "mock-upload-id-" + key,
		Key:      key,
		Parts:    []string{},
	}, nil
}

func (m *MockStorage) GeneratePresignedPartURL(ctx context.Context, key string, uploadID string, partNumber int, expiry time.Duration) (string, error) {
	return fmt.Sprintf("https://mock-s3.example.com/upload/%s/part/%d", key, partNumber), nil
}

func (m *MockStorage) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []storage.CompletedPart) error {
	// 模拟合并分片
	m.files[key] = []byte("multipart-upload-completed")
	return nil
}

func (m *MockStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, opts *storage.PresignOptions) (string, error) {
	return fmt.Sprintf("https://mock-s3.example.com/download/%s", key), nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.files, key)
	return nil
}

// setupTestServerWithMock 创建使用 Mock Storage 的测试服务器
func setupTestServerWithMock(t *testing.T) (*gin.Engine, *gorm.DB, storage.Storage, func()) {
	// 加载配置
	cfg, err := config.Load("../../configs")
	require.NoError(t, err, "Failed to load config")

	// 连接测试数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to database")

	// 自动迁移（忽略约束错误）
	_ = db.AutoMigrate(&models.File{})

	// 使用 Mock Storage
	mockStorage := NewMockStorage()

	// 初始化服务
	fileRepo := repositories.NewFileRepository(db)
	fileService := services.NewFileService(fileRepo, mockStorage, db)
	fileHandler := handlers.NewFileHandler(fileService)

	// 创建路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	logger, _ := zap.NewDevelopment()
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.ErrorHandler())

	api := router.Group("/api/v1")
	{
		files := api.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadDirect)
			files.POST("/upload/presigned", fileHandler.InitPresignedUpload)
			files.POST("/upload/confirm", fileHandler.ConfirmUpload)
			files.POST("/upload/multipart/init", fileHandler.InitMultipartUpload)
			files.POST("/upload/multipart/part-url", fileHandler.GeneratePartURL)
			files.POST("/upload/multipart/complete", fileHandler.CompleteMultipartUpload)
			files.GET("/:id/download-url", fileHandler.GetDownloadURL)
			files.GET("/:id", fileHandler.GetFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
		}
	}

	// 清理函数
	cleanup := func() {
		// 清理测试数据
		db.Exec("DELETE FROM files WHERE name LIKE 'test_%'")
	}

	return router, db, mockStorage, cleanup
}

// TestUploadDirectWithMock 测试直接上传小文件（使用 Mock）
func TestUploadDirectWithMock(t *testing.T) {
	router, db, mockStorage, cleanup := setupTestServerWithMock(t)
	_ = cleanup // 临时禁用清理，保留测试数据
	// defer cleanup()

	t.Run("成功上传小文件", func(t *testing.T) {
		// 准备测试文件
		fileContent := []byte("This is a test file content")
		fileName := fmt.Sprintf("test_upload_%d.txt", time.Now().Unix())

		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加表单字段
		_ = writer.WriteField("name", fileName)
		_ = writer.WriteField("content_type", "text/plain")

		// 添加文件
		part, err := writer.CreateFormFile("file", fileName)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
		writer.Close()

		// 发送请求
		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 0, resp.Code)

		// 验证响应数据
		data := resp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))
		assert.Greater(t, fileID, uint(0))
		assert.Equal(t, fileName, data["name"])
		assert.Equal(t, "completed", data["status"])
		assert.NotEmpty(t, data["storage_key"])
		assert.NotEmpty(t, data["download_url"])

		// 验证数据库记录
		var file models.File
		err = db.First(&file, fileID).Error
		require.NoError(t, err)
		assert.Equal(t, fileName, file.Name)
		assert.Equal(t, int64(len(fileContent)), file.Size)
		assert.Equal(t, models.FileStatusCompleted, file.Status)

		// 验证 Mock Storage 中的文件
		storedData := mockStorage.(*MockStorage).files[file.StorageKey]
		assert.Equal(t, fileContent, storedData)
	})

	t.Run("缺少文件参数", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", "test.txt")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestPresignedUploadWithMock 测试预签名上传流程（使用 Mock）
func TestPresignedUploadWithMock(t *testing.T) {
	router, db, _, cleanup := setupTestServerWithMock(t)
	defer cleanup()

	t.Run("完整预签名上传流程", func(t *testing.T) {
		fileName := fmt.Sprintf("test_presigned_%d.txt", time.Now().Unix())
		fileContent := []byte("Presigned upload test content")

		// Step 1: 初始化预签名上传
		initReq := map[string]interface{}{
			"name":         fileName,
			"content_type": "text/plain",
			"size":         len(fileContent),
		}
		initBody, _ := json.Marshal(initReq)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload/presigned", bytes.NewReader(initBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var initResp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &initResp)
		require.NoError(t, err)

		data := initResp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))
		uploadURL := data["upload_url"].(string)
		storageKey := data["storage_key"].(string)

		assert.Greater(t, fileID, uint(0))
		assert.NotEmpty(t, uploadURL)
		assert.NotEmpty(t, storageKey)
		assert.Contains(t, uploadURL, "mock-s3.example.com")

		// Step 2: 确认上传完成（跳过实际上传到 Mock S3）
		confirmReq := map[string]interface{}{
			"file_id": fileID,
		}
		confirmBody, _ := json.Marshal(confirmReq)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/files/upload/confirm", bytes.NewReader(confirmBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var confirmResp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &confirmResp)
		require.NoError(t, err)

		confirmData := confirmResp.Data.(map[string]interface{})
		assert.Equal(t, "completed", confirmData["status"])

		// 验证数据库状态
		var file models.File
		err = db.First(&file, fileID).Error
		require.NoError(t, err)
		assert.Equal(t, models.FileStatusCompleted, file.Status)
	})
}

// TestMultipartUploadWithMock 测试分片上传流程（使用 Mock）
func TestMultipartUploadWithMock(t *testing.T) {
	router, db, _, cleanup := setupTestServerWithMock(t)
	defer cleanup()

	t.Run("完整分片上传流程", func(t *testing.T) {
		fileName := fmt.Sprintf("test_multipart_%d.bin", time.Now().Unix())
		totalSize := 10 * 1024 * 1024 // 10MB

		// Step 1: 初始化分片上传
		initReq := map[string]interface{}{
			"name":         fileName,
			"content_type": "application/octet-stream",
			"size":         totalSize,
		}
		initBody, _ := json.Marshal(initReq)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload/multipart/init", bytes.NewReader(initBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var initResp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &initResp)
		require.NoError(t, err)

		data := initResp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))
		uploadID := data["upload_id"].(string)
		storageKey := data["storage_key"].(string)

		assert.Greater(t, fileID, uint(0))
		assert.NotEmpty(t, uploadID)
		assert.NotEmpty(t, storageKey)

		// Step 2: 生成分片 URL
		for partNumber := 1; partNumber <= 2; partNumber++ {
			partURLReq := map[string]interface{}{
				"file_id":     fileID,
				"part_number": partNumber,
			}
			partURLBody, _ := json.Marshal(partURLReq)

			req = httptest.NewRequest(http.MethodPost, "/api/v1/files/upload/multipart/part-url", bytes.NewReader(partURLBody))
			req.Header.Set("Content-Type", "application/json")
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var partURLResp response.Response
			err = json.Unmarshal(w.Body.Bytes(), &partURLResp)
			require.NoError(t, err)

			partData := partURLResp.Data.(map[string]interface{})
			partURL := partData["upload_url"].(string)
			assert.Contains(t, partURL, "mock-s3.example.com")
		}

		// Step 3: 完成分片上传
		completedParts := []map[string]interface{}{
			{"part_number": 1, "etag": "\"mock-etag-1\""},
			{"part_number": 2, "etag": "\"mock-etag-2\""},
		}

		completeReq := map[string]interface{}{
			"file_id": fileID,
			"parts":   completedParts,
		}
		completeBody, _ := json.Marshal(completeReq)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/files/upload/multipart/complete", bytes.NewReader(completeBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var completeResp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &completeResp)
		require.NoError(t, err)

		completeData := completeResp.Data.(map[string]interface{})
		assert.Equal(t, "completed", completeData["status"])

		// 验证数据库状态
		var file models.File
		err = db.First(&file, fileID).Error
		require.NoError(t, err)
		assert.Equal(t, models.FileStatusCompleted, file.Status)
	})
}

// TestGetDownloadURLWithMock 测试获取下载 URL（使用 Mock）
func TestGetDownloadURLWithMock(t *testing.T) {
	router, _, _, cleanup := setupTestServerWithMock(t)
	defer cleanup()

	t.Run("成功获取下载 URL", func(t *testing.T) {
		// 先上传一个文件
		fileName := fmt.Sprintf("test_download_%d.txt", time.Now().Unix())
		fileContent := []byte("Download test content")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", fileName)
		_ = writer.WriteField("content_type", "text/plain")
		part, _ := writer.CreateFormFile("file", fileName)
		_, _ = part.Write(fileContent)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var uploadResp response.Response
		_ = json.Unmarshal(w.Body.Bytes(), &uploadResp)
		data := uploadResp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))

		// 获取下载 URL
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/files/%d/download-url", fileID), nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var downloadResp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &downloadResp)
		require.NoError(t, err)

		downloadData := downloadResp.Data.(map[string]interface{})
		downloadURL := downloadData["download_url"].(string)
		assert.NotEmpty(t, downloadURL)
		assert.Contains(t, downloadURL, "mock-s3.example.com")
		assert.Equal(t, float64(900), downloadData["expires_in"])
	})

	t.Run("文件不存在", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/99999/download-url", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestGetFileWithMock 测试获取文件信息（使用 Mock）
func TestGetFileWithMock(t *testing.T) {
	router, _, _, cleanup := setupTestServerWithMock(t)
	defer cleanup()

	t.Run("成功获取文件信息", func(t *testing.T) {
		// 先上传一个文件
		fileName := fmt.Sprintf("test_getfile_%d.txt", time.Now().Unix())
		fileContent := []byte("Get file test content")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", fileName)
		_ = writer.WriteField("content_type", "text/plain")
		part, _ := writer.CreateFormFile("file", fileName)
		_, _ = part.Write(fileContent)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var uploadResp response.Response
		_ = json.Unmarshal(w.Body.Bytes(), &uploadResp)
		data := uploadResp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))

		// 获取文件信息
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/files/%d", fileID), nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var getResp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &getResp)
		require.NoError(t, err)

		fileData := getResp.Data.(map[string]interface{})
		assert.Equal(t, float64(fileID), fileData["file_id"])
		assert.Equal(t, fileName, fileData["name"])
		assert.Equal(t, float64(len(fileContent)), fileData["size"])
		assert.Equal(t, "text/plain", fileData["content_type"])
		assert.Equal(t, "completed", fileData["status"])
		assert.NotEmpty(t, fileData["created_at"])
	})

	t.Run("文件不存在", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/99999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestDeleteFileWithMock 测试删除文件（使用 Mock）
func TestDeleteFileWithMock(t *testing.T) {
	router, db, _, cleanup := setupTestServerWithMock(t)
	defer cleanup()

	t.Run("成功删除文件", func(t *testing.T) {
		// 先上传一个文件
		fileName := fmt.Sprintf("test_delete_%d.txt", time.Now().Unix())
		fileContent := []byte("Delete test content")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", fileName)
		_ = writer.WriteField("content_type", "text/plain")
		part, _ := writer.CreateFormFile("file", fileName)
		_, _ = part.Write(fileContent)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var uploadResp response.Response
		_ = json.Unmarshal(w.Body.Bytes(), &uploadResp)
		data := uploadResp.Data.(map[string]interface{})
		fileID := uint(data["file_id"].(float64))

		// 删除文件
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/files/%d", fileID), nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var deleteResp response.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		require.NoError(t, err)

		deleteData := deleteResp.Data.(map[string]interface{})
		assert.Equal(t, "File deleted successfully", deleteData["message"])

		// 验证数据库记录已删除
		var file models.File
		err = db.First(&file, fileID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("删除不存在的文件", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/99999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	// 设置测试环境变量
	os.Setenv("APP_ENV", "test")

	// 运行测试
	code := m.Run()

	// 退出
	os.Exit(code)
}
