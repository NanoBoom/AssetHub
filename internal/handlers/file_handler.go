package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NanoBoom/asethub/internal/errors"
	"github.com/NanoBoom/asethub/internal/services"
	"github.com/NanoBoom/asethub/pkg/response"
	"github.com/NanoBoom/asethub/pkg/storage"
	"github.com/NanoBoom/asethub/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService services.FileService
}

// NewFileHandler 创建文件处理器实例
func NewFileHandler(fileService services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// ===== 请求/响应结构体定义 =====

// UploadDirectRequest 直接上传请求
type UploadDirectRequest struct {
	Name        string `form:"name" binding:"required" example:"example.txt"`
	ContentType string `form:"content_type" example:"text/plain"`
}

// UploadDirectResponse 直接上传响应
type UploadDirectResponse struct {
	FileID      uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"example.txt"`
	Size        int64     `json:"size" example:"1024"`
	StorageKey  string    `json:"storage_key" example:"files/1234567890/example.txt"`
	Status      string    `json:"status" example:"completed"`
	DownloadURL string    `json:"download_url" example:"https://s3.amazonaws.com/..."`
}

// InitPresignedUploadRequest 初始化预签名上传请求
type InitPresignedUploadRequest struct {
	Name        string `json:"name" binding:"required" example:"example.txt"`
	ContentType string `json:"content_type" example:"text/plain"`
	Size        int64  `json:"size" binding:"required" example:"1024"`
}

// InitPresignedUploadResponse 初始化预签名上传响应
type InitPresignedUploadResponse struct {
	FileID     uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UploadURL  string    `json:"upload_url" example:"https://s3.amazonaws.com/..."`
	StorageKey string    `json:"storage_key" example:"files/1234567890/example.txt"`
	ExpiresIn  int64     `json:"expires_in" example:"3600"`
}

// ConfirmUploadRequest 确认上传请求
type ConfirmUploadRequest struct {
	FileID uuid.UUID `json:"file_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ConfirmUploadResponse 确认上传响应
type ConfirmUploadResponse struct {
	FileID uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status string    `json:"status" example:"completed"`
}

// InitMultipartUploadRequest 初始化分片上传请求
type InitMultipartUploadRequest struct {
	Name        string `json:"name" binding:"required" example:"large-video.mp4"`
	ContentType string `json:"content_type" example:"video/mp4"`
	Size        int64  `json:"size" binding:"required" example:"104857600"`
}

// InitMultipartUploadResponse 初始化分片上传响应
type InitMultipartUploadResponse struct {
	FileID     uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UploadID   string    `json:"upload_id" example:"upload-id-123"`
	StorageKey string    `json:"storage_key" example:"files/1234567890/large-video.mp4"`
}

// GeneratePartURLRequest 生成分片 URL 请求
type GeneratePartURLRequest struct {
	FileID     uuid.UUID `json:"file_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	PartNumber int       `json:"part_number" binding:"required,min=1" example:"1"`
}

// GeneratePartURLResponse 生成分片 URL 响应
type GeneratePartURLResponse struct {
	PartNumber int    `json:"part_number" example:"1"`
	UploadURL  string `json:"upload_url" example:"https://s3.amazonaws.com/..."`
	ExpiresIn  int64  `json:"expires_in" example:"3600"`
}

// CompleteMultipartUploadRequest 完成分片上传请求
type CompleteMultipartUploadRequest struct {
	FileID uuid.UUID              `json:"file_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Parts  []CompletedPartRequest `json:"parts" binding:"required"`
}

// CompletedPartRequest 已完成的分片
type CompletedPartRequest struct {
	PartNumber int    `json:"part_number" binding:"required" example:"1"`
	ETag       string `json:"etag" binding:"required" example:"\"abc123\""`
}

// CompleteMultipartUploadResponse 完成分片上传响应
type CompleteMultipartUploadResponse struct {
	FileID uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status string    `json:"status" example:"completed"`
}

// GetDownloadURLResponse 获取下载 URL 响应
type GetDownloadURLResponse struct {
	FileID      uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DownloadURL string    `json:"download_url" example:"https://s3.amazonaws.com/..."`
	ExpiresIn   int64     `json:"expires_in" example:"900"`
}

// GetFileResponse 获取文件信息响应
type GetFileResponse struct {
	FileID      uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"example.txt"`
	Size        int64     `json:"size" example:"1024"`
	ContentType string    `json:"content_type" example:"text/plain"`
	StorageKey  string    `json:"storage_key" example:"files/1234567890/example.txt"`
	Status      string    `json:"status" example:"completed"`
	CreatedAt   string    `json:"created_at" example:"2026-02-06T00:00:00Z"`
}

// DeleteFileResponse 删除文件响应
type DeleteFileResponse struct {
	FileID  uuid.UUID `json:"file_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message string    `json:"message" example:"File deleted successfully"`
}

// ===== API 端点实现 =====

// UploadDirect godoc
// @Summary      直接上传小文件
// @Description  后端代理上传小文件到 S3
// @Tags         Direct Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        name formData string true "文件名"
// @Param        content_type formData string false "MIME 类型"
// @Param        file formData file true "文件内容"
// @Success      201 {object} response.Response{data=UploadDirectResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files [post]
func (h *FileHandler) UploadDirect(c *gin.Context) {
	// 解析表单
	var req UploadDirectRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(errors.NewBadRequestError("file is required", err))
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}
	defer file.Close()

	// 调用 Service 层上传
	uploadedFile, err := h.fileService.UploadDirect(
		c.Request.Context(),
		req.Name,
		req.ContentType,
		fileHeader.Size,
		file,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 生成下载 URL
	downloadURL, err := h.fileService.GetDownloadURL(c.Request.Context(), uploadedFile.ID, 15*time.Minute)
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 返回响应
	c.Status(http.StatusCreated)
	response.Success(c, UploadDirectResponse{
		FileID:      uploadedFile.ID,
		Name:        uploadedFile.Name,
		Size:        uploadedFile.Size,
		StorageKey:  uploadedFile.StorageKey,
		Status:      string(uploadedFile.Status),
		DownloadURL: downloadURL,
	})
}

// InitPresignedUpload godoc
// @Summary      获取小文件上传预签名 URL
// @Description  生成预签名 URL 供前端直接上传到 S3
// @Tags         Presigned Upload
// @Accept       json
// @Produce      json
// @Param        body body InitPresignedUploadRequest true "上传信息"
// @Success      201 {object} response.Response{data=InitPresignedUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/presigned [post]
func (h *FileHandler) InitPresignedUpload(c *gin.Context) {
	var req InitPresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 调用 Service 层生成预签名 URL
	result, err := h.fileService.InitPresignedUpload(
		c.Request.Context(),
		req.Name,
		req.ContentType,
		req.Size,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 返回响应
	c.Status(http.StatusCreated)
	response.Success(c, InitPresignedUploadResponse{
		FileID:     result.FileID,
		UploadURL:  result.UploadURL,
		StorageKey: result.StorageKey,
		ExpiresIn:  result.ExpiresIn,
	})
}

// ConfirmUpload godoc
// @Summary      确认前端直传完成
// @Description  前端上传完成后调用此接口确认
// @Tags         Presigned Upload
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Success      200 {object} response.Response{data=ConfirmUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/completion [post]
func (h *FileHandler) ConfirmUpload(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 调用 Service 层确认上传
	file, err := h.fileService.ConfirmUpload(c.Request.Context(), fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}

	// 返回响应
	response.Success(c, ConfirmUploadResponse{
		FileID: file.ID,
		Status: string(file.Status),
	})
}

// InitMultipartUpload godoc
// @Summary      初始化大文件分片上传
// @Description  初始化分片上传，返回 UploadID
// @Tags         Multipart Upload
// @Accept       json
// @Produce      json
// @Param        body body InitMultipartUploadRequest true "文件信息"
// @Success      201 {object} response.Response{data=InitMultipartUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/multipart [post]
func (h *FileHandler) InitMultipartUpload(c *gin.Context) {
	var req InitMultipartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 调用 Service 层初始化分片上传
	result, err := h.fileService.InitMultipartUpload(
		c.Request.Context(),
		req.Name,
		req.ContentType,
		req.Size,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 返回响应
	c.Status(http.StatusCreated)
	response.Success(c, InitMultipartUploadResponse{
		FileID:     result.FileID,
		UploadID:   result.UploadID,
		StorageKey: result.StorageKey,
	})
}

// GeneratePartURL godoc
// @Summary      生成分片上传预签名 URL
// @Description  为指定分片生成预签名 URL
// @Tags         Multipart Upload
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Param        body body GeneratePartURLRequest true "分片信息"
// @Success      200 {object} response.Response{data=GeneratePartURLResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/multipart/parts [post]
func (h *FileHandler) GeneratePartURL(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 解析请求体（只需要 part_number）
	var req struct {
		PartNumber int `json:"part_number" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 调用 Service 层生成分片 URL
	partURL, err := h.fileService.GeneratePartUploadURL(
		c.Request.Context(),
		fileID,
		req.PartNumber,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}

	// 返回响应
	response.Success(c, GeneratePartURLResponse{
		PartNumber: req.PartNumber,
		UploadURL:  partURL,
		ExpiresIn:  3600, // 1 小时
	})
}

// CompleteMultipartUpload godoc
// @Summary      完成大文件分片上传
// @Description  提交所有分片的 ETag，完成上传
// @Tags         Multipart Upload
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Param        body body CompleteMultipartUploadRequest true "分片列表"
// @Success      200 {object} response.Response{data=CompleteMultipartUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/multipart/completion [post]
func (h *FileHandler) CompleteMultipartUpload(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 解析请求体（只需要 parts）
	var req struct {
		Parts []CompletedPartRequest `json:"parts" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 转换为 Service 层的类型
	parts := make([]storage.CompletedPart, len(req.Parts))
	for i, part := range req.Parts {
		parts[i] = storage.CompletedPart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	// 调用 Service 层完成分片上传
	file, err := h.fileService.CompleteMultipartUpload(
		c.Request.Context(),
		fileID,
		parts,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}

	// 返回响应
	response.Success(c, CompleteMultipartUploadResponse{
		FileID: file.ID,
		Status: string(file.Status),
	})
}

// GetDownloadURL godoc
// @Summary      获取文件下载 URL
// @Description  生成文件下载预签名 URL
// @Tags         File Management
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Success      200 {object} response.Response{data=GetDownloadURLResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/link [get]
func (h *FileHandler) GetDownloadURL(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 调用 Service 层生成下载 URL（15 分钟有效期）
	downloadURL, err := h.fileService.GetDownloadURL(
		c.Request.Context(),
		fileID,
		15*time.Minute,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}

	// 返回响应
	response.Success(c, GetDownloadURLResponse{
		FileID:      fileID,
		DownloadURL: downloadURL,
		ExpiresIn:   900, // 15 分钟
	})
}

// GetFile godoc
// @Summary      获取文件元数据
// @Description  根据文件 UUID 获取文件信息
// @Tags         File Management
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Success      200 {object} response.Response{data=GetFileResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 调用 Service 层获取文件信息
	file, err := h.fileService.GetFile(c.Request.Context(), fileID)
	if err != nil {
		c.Error(errors.NewNotFoundError("file not found"))
		return
	}

	// 返回响应
	response.Success(c, GetFileResponse{
		FileID:      file.ID,
		Name:        file.Name,
		Size:        file.Size,
		ContentType: file.ContentType,
		StorageKey:  file.StorageKey,
		Status:      string(file.Status),
		CreatedAt:   file.CreatedAt.Format(time.RFC3339),
	})
}

// DeleteFile godoc
// @Summary      删除文件
// @Description  删除文件（S3 + 数据库）
// @Tags         File Management
// @Accept       json
// @Produce      json
// @Param        id path string true "文件 UUID" format(uuid)
// @Success      204 "No Content"
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 调用 Service 层删除文件
	err = h.fileService.DeleteFile(c.Request.Context(), fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}

	// 返回 204 No Content（无响应体）
	c.Status(http.StatusNoContent)
}

// DownloadFile godoc
// @Summary      直接下载文件
// @Description  获取文件内容（流式传输）
// @Tags         File Management
// @Produce      octet-stream
// @Param        id path string true "文件 UUID" format(uuid)
// @Success      200 {file} binary "文件内容"
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/download [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	// 解析 UUID
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil || fileID == uuid.Nil {
		c.Error(errors.NewBadRequestError("invalid or nil UUID", err))
		return
	}

	// 调用 Service 层下载文件
	reader, file, err := h.fileService.DownloadFile(c.Request.Context(), fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Error(errors.NewNotFoundError("file not found"))
		} else if strings.Contains(err.Error(), "not ready") {
			c.Error(errors.NewBadRequestError("file is not ready for download", err))
		} else {
			c.Error(errors.NewInternalError(err))
		}
		return
	}
	defer reader.Close()

	// 设置 Content-Type
	c.Header("Content-Type", file.ContentType)

	// 设置 Content-Length
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	// 根据文件类型决定 Content-Disposition
	disposition := "attachment"
	if utils.IsPreviewable(file.ContentType) {
		disposition = "inline"
	}
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, file.Name))

	// 流式传输文件内容
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		// 注意：此时已经开始写入响应，无法返回错误响应
		// 只能记录日志
		c.Error(errors.NewInternalError(err))
		return
	}
}

