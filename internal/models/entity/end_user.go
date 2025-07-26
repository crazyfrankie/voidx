package entity

import "github.com/google/uuid"

// EndUser 终端用户表模型
type EndUser struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:end_user_tenant_id_idx" json:"tenant_id"`
	AppID    uuid.UUID `gorm:"type:uuid;not null;index:end_user_app_id_idx" json:"app_id"`
	Utime    int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime    int64     `gorm:"autoCreateTime" json:"ctime"`
}
