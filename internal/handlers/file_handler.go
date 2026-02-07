package handlers

import (
	"strconv"
	"time"

	"github.com/NanoBoom/asethub/internal/errors"
	"github.com/NanoBoom/asethub/internal/services"
	"github.com/NanoBoom/asethub/pkg/response"
	"github.com/NanoBoom/asethub/pkg/storage"
	"github.com/gin-gonic/gin"
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
	FileID      uint   `json:"file_id" example:"1"`
	Name        string `json:"name" example:"example.txt"`
	Size        int64  `json:"size" example:"1024"`
	StorageKey  string `json:"storage_key" example:"files/1234567890/example.txt"`
	Status      string `json:"status" example:"completed"`
	DownloadURL string `json:"download_url" example:"https://s3.amazonaws.com/..."`
}

// InitPresignedUploadRequest 初始化预签名上传请求
type InitPresignedUploadRequest struct {
	Name        string `json:"name" binding:"required" example:"example.txt"`
	ContentType string `json:"content_type" example:"text/plain"`
	Size        int64  `json:"size" binding:"required" example:"1024"`
}

// InitPresignedUploadResponse 初始化预签名上传响应
type InitPresignedUploadResponse struct {
	FileID     uint   `json:"file_id" example:"1"`
	UploadURL  string `json:"upload_url" example:"https://s3.amazonaws.com/..."`
	StorageKey string `json:"storage_key" example:"files/1234567890/example.txt"`
	ExpiresIn  int64  `json:"expires_in" example:"3600"`
}

// ConfirmUploadRequest 确认上传请求
type ConfirmUploadRequest struct {
	FileID uint `json:"file_id" binding:"required" example:"1"`
}

// ConfirmUploadResponse 确认上传响应
type ConfirmUploadResponse struct {
	FileID uint   `json:"file_id" example:"1"`
	Status string `json:"status" example:"completed"`
}

// InitMultipartUploadRequest 初始化分片上传请求
type InitMultipartUploadRequest struct {
	Name        string `json:"name" binding:"required" example:"large-video.mp4"`
	ContentType string `json:"content_type" example:"video/mp4"`
	Size        int64  `json:"size" binding:"required" example:"104857600"`
}

// InitMultipartUploadResponse 初始化分片上传响应
type InitMultipartUploadResponse struct {
	FileID     uint   `json:"file_id" example:"1"`
	UploadID   string `json:"upload_id" example:"upload-id-123"`
	StorageKey string `json:"storage_key" example:"files/1234567890/large-video.mp4"`
}

// GeneratePartURLRequest 生成分片 URL 请求
type GeneratePartURLRequest struct {
	FileID     uint `json:"file_id" binding:"required" example:"1"`
	PartNumber int  `json:"part_number" binding:"required,min=1" example:"1"`
}

// GeneratePartURLResponse 生成分片 URL 响应
type GeneratePartURLResponse struct {
	PartNumber int    `json:"part_number" example:"1"`
	UploadURL  string `json:"upload_url" example:"https://s3.amazonaws.com/..."`
	ExpiresIn  int64  `json:"expires_in" example:"3600"`
}

// CompleteMultipartUploadRequest 完成分片上传请求
type CompleteMultipartUploadRequest struct {
	FileID uint                    `json:"file_id" binding:"required" example:"1"`
	Parts  []CompletedPartRequest  `json:"parts" binding:"required"`
}

// CompletedPartRequest 已完成的分片
type CompletedPartRequest struct {
	PartNumber int    `json:"part_number" binding:"required" example:"1"`
	ETag       string `json:"etag" binding:"required" example:"\"abc123\""`
}

// CompleteMultipartUploadResponse 完成分片上传响应
type CompleteMultipartUploadResponse struct {
	FileID uint   `json:"file_id" example:"1"`
	Status string `json:"status" example:"completed"`
}

// GetDownloadURLResponse 获取下载 URL 响应
type GetDownloadURLResponse struct {
	FileID      uint   `json:"file_id" example:"1"`
	DownloadURL string `json:"download_url" example:"https://s3.amazonaws.com/..."`
	ExpiresIn   int64  `json:"expires_in" example:"900"`
}

// GetFileResponse 获取文件信息响应
type GetFileResponse struct {
	FileID      uint   `json:"file_id" example:"1"`
	Name        string `json:"name" example:"example.txt"`
	Size        int64  `json:"size" example:"1024"`
	ContentType string `json:"content_type" example:"text/plain"`
	StorageKey  string `json:"storage_key" example:"files/1234567890/example.txt"`
	Status      string `json:"status" example:"completed"`
	CreatedAt   string `json:"created_at" example:"2026-02-06T00:00:00Z"`
}

// DeleteFileResponse 删除文件响应
type DeleteFileResponse struct {
	FileID  uint   `json:"file_id" example:"1"`
	Message string `json:"message" example:"File deleted successfully"`
}

// ===== API 端点实现 =====

// UploadDirect godoc
// @Summary      直接上传小文件
// @Description  后端代理上传小文件到 S3
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        name formData string true "文件名"
// @Param        content_type formData string false "MIME 类型"
// @Param        file formData file true "文件内容"
// @Success      200 {object} response.Response{data=UploadDirectResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload [post]
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body body InitPresignedUploadRequest true "上传信息"
// @Success      200 {object} response.Response{data=InitPresignedUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload/presigned [post]
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body body ConfirmUploadRequest true "文件 ID"
// @Success      200 {object} response.Response{data=ConfirmUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload/confirm [post]
func (h *FileHandler) ConfirmUpload(c *gin.Context) {
	var req ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 调用 Service 层确认上传
	file, err := h.fileService.ConfirmUpload(c.Request.Context(), req.FileID)
	if err != nil {
		c.Error(errors.NewInternalError(err))
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body body InitMultipartUploadRequest true "文件信息"
// @Success      200 {object} response.Response{data=InitMultipartUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload/multipart/init [post]
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
	response.Success(c, InitMultipartUploadResponse{
		FileID:     result.FileID,
		UploadID:   result.UploadID,
		StorageKey: result.StorageKey,
	})
}

// GeneratePartURL godoc
// @Summary      生成分片上传预签名 URL
// @Description  为指定分片生成预签名 URL
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body body GeneratePartURLRequest true "分片信息"
// @Success      200 {object} response.Response{data=GeneratePartURLResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload/multipart/part-url [post]
func (h *FileHandler) GeneratePartURL(c *gin.Context) {
	var req GeneratePartURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError("invalid request", err))
		return
	}

	// 调用 Service 层生成分片 URL
	partURL, err := h.fileService.GeneratePartUploadURL(
		c.Request.Context(),
		req.FileID,
		req.PartNumber,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        body body CompleteMultipartUploadRequest true "分片列表"
// @Success      200 {object} response.Response{data=CompleteMultipartUploadResponse}
// @Failure      400 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/upload/multipart/complete [post]
func (h *FileHandler) CompleteMultipartUpload(c *gin.Context) {
	var req CompleteMultipartUploadRequest
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
		req.FileID,
		parts,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        id path int true "文件 ID"
// @Success      200 {object} response.Response{data=GetDownloadURLResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id}/download-url [get]
func (h *FileHandler) GetDownloadURL(c *gin.Context) {
	// 解析文件 ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid file ID", err))
		return
	}

	// 调用 Service 层生成下载 URL（15 分钟有效期）
	downloadURL, err := h.fileService.GetDownloadURL(
		c.Request.Context(),
		uint(fileID),
		15*time.Minute,
	)
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 返回响应
	response.Success(c, GetDownloadURLResponse{
		FileID:      uint(fileID),
		DownloadURL: downloadURL,
		ExpiresIn:   900, // 15 分钟
	})
}

// GetFile godoc
// @Summary      获取文件元数据
// @Description  根据文件 ID 获取文件信息
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        id path int true "文件 ID"
// @Success      200 {object} response.Response{data=GetFileResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	// 解析文件 ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid file ID", err))
		return
	}

	// 调用 Service 层获取文件信息
	file, err := h.fileService.GetFile(c.Request.Context(), uint(fileID))
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
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        id path int true "文件 ID"
// @Success      200 {object} response.Response{data=DeleteFileResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /api/v1/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// 解析文件 ID
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid file ID", err))
		return
	}

	// 调用 Service 层删除文件
	err = h.fileService.DeleteFile(c.Request.Context(), uint(fileID))
	if err != nil {
		c.Error(errors.NewInternalError(err))
		return
	}

	// 返回响应
	response.Success(c, DeleteFileResponse{
		FileID:  uint(fileID),
		Message: "File deleted successfully",
	})
}

