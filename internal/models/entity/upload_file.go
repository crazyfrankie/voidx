package entity

import "github.com/google/uuid"

// UploadFile 上传文件模型
type UploadFile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;index:upload_file_account_id_idx" json:"account_id"`
	Name      string    `gorm:"size:255;not null;default:''" json:"name"`
	Key       string    `gorm:"size:255;not null;default:''" json:"key"`
	Size      int64     `gorm:"not null;default:0" json:"size"`
	Extension string    `gorm:"size:255;not null;default:''" json:"extension"`
	MimeType  string    `gorm:"size:255;not null;default:''" json:"mime_type"`
	Hash      string    `gorm:"size:255;not null;default:''" json:"hash"`
	Utime     int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime" json:"ctime"`
}
