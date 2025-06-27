package dao

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type App struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;index:idx_app_account_id" json:"account_id"`
	Name      string    `gorm:"size:255;not null;default:''" json:"name"`
	Icon      string    `gorm:"size:255;not null;default:''" json:"icon"`
	Config    []byte    `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	UpdatedAt int64     `gorm:"autoUpdateTime:milli" json:"updated_at"`
	CreatedAt int64     `gorm:"autoCreateTime:milli" json:"created_at"`
}

type ChatDao struct {
	db *gorm.DB
}

func NewChatDao(db *gorm.DB) *ChatDao {
	return &ChatDao{db: db}
}
