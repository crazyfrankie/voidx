package entity

import (
	"github.com/google/uuid"
)

// ApiToolProvider API工具提供者模型
type ApiToolProvider struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID     uuid.UUID `gorm:"type:uuid;not null;index:api_tool_provider_account_id_idx" json:"account_id"`
	Name          string    `gorm:"size:255;not null;default:'';index:api_tool_name_idx" json:"name"`
	Icon          string    `gorm:"size:255;not null;default:''" json:"icon"`
	Description   string    `gorm:"type:text;not null;default:''" json:"description"`
	OpenapiSchema string    `gorm:"type:text;not null;default:''" json:"openapi_schema"`
	Headers       []Header  `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"headers"`
	Utime         int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime         int64     `gorm:"autoCreateTime" json:"ctime"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ApiTool API工具表
type ApiTool struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID   uuid.UUID        `gorm:"type:uuid;not null;index:api_tool_account_id_idx" json:"account_id"`
	ProviderID  uuid.UUID        `gorm:"type:uuid;not null;index:api_tool_provider_id_name_idx,composite:provider_name" json:"provider_id"`
	Name        string           `gorm:"size:255;not null;default:'';index:api_tool_provider_id_name_idx,composite:provider_name" json:"name"`
	Description string           `gorm:"type:text;not null;default:''" json:"description"`
	URL         string           `gorm:"size:255;not null;default:''" json:"url"`
	Method      string           `gorm:"size:255;not null;default:''" json:"method"`
	Parameters  []map[string]any `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"parameters"`
	Utime       int64            `gorm:"autoUpdateTime" json:"utime"`
	Ctime       int64            `gorm:"autoCreateTime" json:"ctime"`
}
