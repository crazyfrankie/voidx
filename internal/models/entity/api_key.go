package entity

import "github.com/google/uuid"

// ApiKey API秘钥模型
type ApiKey struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;index:api_key_account_id_idx" json:"account_id"`
	ApiKey    string    `gorm:"size:255;not null;default:'';index:api_key_api_key_idx" json:"api_key"`
	IsActive  bool      `gorm:"not null;default:false" json:"is_active"`
	Remark    string    `gorm:"size:255;not null;default:''" json:"remark"`
	Utime     int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime" json:"ctime"`
}
