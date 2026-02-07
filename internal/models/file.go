package models

// FileStatus 文件上传状态
type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"     // 待上传（已创建记录，等待上传）
	FileStatusUploading  FileStatus = "uploading"   // 上传中（分片上传进行中）
	FileStatusCompleted  FileStatus = "completed"   // 上传完成
	FileStatusFailed     FileStatus = "failed"      // 上传失败
)

// File 文件元数据模型
type File struct {
	BaseModel
	Name        string     `gorm:"type:varchar(255);not null;index" json:"name"`                    // 文件名
	Size        int64      `gorm:"not null" json:"size"`                                            // 文件大小（字节）
	ContentType string     `gorm:"type:varchar(100)" json:"content_type"`                           // MIME 类型
	StorageKey  string     `gorm:"type:varchar(500);not null;uniqueIndex" json:"storage_key"`       // 存储键（S3 对象键）
	Status      FileStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"` // 上传状态
	Hash        string     `gorm:"type:varchar(64);index" json:"hash"`                              // 文件哈希值（SHA256，可选）
	UploadID    string     `gorm:"type:varchar(255)" json:"upload_id"`                              // 分片上传 ID（仅分片上传时使用）
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}
