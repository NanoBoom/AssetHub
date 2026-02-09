package utils

import (
	"path/filepath"
	"strings"
)

// extensionToMIME 文件扩展名到 MIME 类型的映射
var extensionToMIME = map[string]string{
	// 图片
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".bmp":  "image/bmp",

	// 视频
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".ogv":  "video/ogg",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".flv":  "video/x-flv",

	// 音频
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".oga":  "audio/ogg",
	".m4a":  "audio/mp4",
	".flac": "audio/flac",

	// 文档
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

	// 文本
	".txt":  "text/plain",
	".html": "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".csv":  "text/csv",

	// 压缩
	".zip": "application/zip",
	".rar": "application/x-rar-compressed",
	".7z":  "application/x-7z-compressed",
	".tar": "application/x-tar",
	".gz":  "application/gzip",
}

// DetectContentTypeFromFilename 根据文件名推断 Content-Type
// 如果无法推断，返回 "application/octet-stream"
func DetectContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if mimeType, ok := extensionToMIME[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// IsPreviewable 判断 MIME 类型是否可在浏览器中预览
func IsPreviewable(contentType string) bool {
	previewable := []string{
		"image/",
		"video/",
		"audio/",
		"application/pdf",
		"text/",
	}

	for _, prefix := range previewable {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

// ValidateContentType 验证 Content-Type 是否与文件名匹配
// 返回：是否匹配，推荐的 Content-Type
func ValidateContentType(filename string, providedType string) (bool, string) {
	expectedType := DetectContentTypeFromFilename(filename)

	// 如果提供的类型是默认值，使用推断的类型
	if providedType == "" || providedType == "application/octet-stream" {
		return false, expectedType
	}

	// 检查是否匹配（忽略参数，如 charset）
	providedBase := strings.Split(providedType, ";")[0]
	expectedBase := strings.Split(expectedType, ";")[0]

	return providedBase == expectedBase, expectedType
}

// GetExtensionFromMIME 从 MIME 类型推断文件扩展名
// 如果无法推断，返回 ".bin"
func GetExtensionFromMIME(mimeType string) string {
	// 移除参数（如 charset）
	baseType := strings.Split(mimeType, ";")[0]
	baseType = strings.TrimSpace(baseType)

	// 优先返回常用扩展名
	preferredExtensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
	}

	if ext, ok := preferredExtensions[baseType]; ok {
		return ext
	}

	// 反向查找扩展名
	for ext, mime := range extensionToMIME {
		if mime == baseType {
			return ext
		}
	}

	// 默认扩展名
	return ".bin"
}
