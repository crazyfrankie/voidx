package entity

import (
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
)

// Workflow 工作流模型
type Workflow struct {
	ID            uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID     uuid.UUID             `gorm:"type:uuid;not null;index:workflow_account_id_idx" json:"account_id"`
	Name          string                `gorm:"size:255;not null;default:''" json:"name"`
	ToolCallName  string                `gorm:"size:255;not null;default:'';index:workflow_tool_call_name_idx" json:"tool_call_name"`
	Icon          string                `gorm:"size:255;not null;default:''" json:"icon"`
	Description   string                `gorm:"type:text;not null;default:''" json:"description"`
	Graph         map[string]any        `gorm:"type:jsonb;serializer:json;not null;default:'{}'::jsonb" json:"graph"`
	DraftGraph    map[string]any        `gorm:"type:jsonb;serializer:json;not null;default:'{}'::jsonb" json:"draft_graph"`
	IsDebugPassed bool                  `gorm:"not null;default:false" json:"is_debug_passed"`
	Status        consts.WorkflowStatus `gorm:"size:255;not null;default:''" json:"status"`
	PublishedAt   int64                 `gorm:"" json:"published_at"`
	Utime         int64                 `gorm:"autoUpdateTime" json:"utime"`
	Ctime         int64                 `gorm:"autoCreateTime" json:"ctime"`
}

// WorkflowResult 工作流存储结果模型
type WorkflowResult struct {
	ID         uuid.UUID                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID      uuid.UUID                   `gorm:"type:uuid;index:workflow_result_app_id_idx" json:"app_id"`
	AccountID  uuid.UUID                   `gorm:"type:uuid;not null;index:workflow_result_account_id_idx" json:"account_id"`
	WorkflowID uuid.UUID                   `gorm:"type:uuid;not null;index:workflow_result_workflow_id_idx" json:"workflow_id"`
	Graph      map[string]any              `gorm:"type:jsonb;serializer:json;not null;default:'{}'::jsonb" json:"graph"`
	State      map[string]any              `gorm:"type:jsonb;serializer:json;not null;default:'{}'::jsonb" json:"state"`
	Latency    float64                     `gorm:"not null;default:0.0" json:"latency"`
	Status     consts.WorkflowResultStatus `gorm:"size:255;not null;default:''" json:"status"`
	Utime      int64                       `gorm:"autoUpdateTime" json:"utime"`
	Ctime      int64                       `gorm:"autoCreateTime" json:"ctime"`
}
