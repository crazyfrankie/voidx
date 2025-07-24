package entity

import "github.com/google/uuid"

// UploadFile 上传文件实体
type UploadFile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;index;not null" json:"account_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Key       string    `gorm:"size:500;not null" json:"key"`
	Size      int64     `gorm:"not null" json:"size"`
	Extension string    `gorm:"size:10;not null" json:"extension"`
	Hash      string    `gorm:"size:64;not null" json:"hash"`
	Utime     int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime" json:"ctime"`
}

// AllowedDocumentExtensions 允许的文档扩展名
var AllowedDocumentExtensions = []string{
	"txt", "md", "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
	"csv", "json", "xml", "html", "htm",
}

// AllowedImageExtensions 允许的图片扩展名
var AllowedImageExtensions = []string{
	"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg",
}
