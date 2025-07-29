package entity

import (
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
)

// Conversation 交流会话模型
type Conversation struct {
	ID         uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID      uuid.UUID         `gorm:"type:uuid;not null;index:conversation_app_id_idx" json:"app_id"`
	Name       string            `gorm:"size:255;not null;default:''" json:"name"`
	Summary    string            `gorm:"type:text;not null;default:''" json:"summary"`
	IsPinned   bool              `gorm:"not null;default:false" json:"is_pinned"`
	IsDeleted  bool              `gorm:"not null;default:false" json:"is_deleted"`
	InvokeFrom consts.InvokeFrom `gorm:"size:255;not null;default:''" json:"invoke_from"`
	CreatedBy  uuid.UUID         `gorm:"type:uuid;index:conversation_app_created_by_idx" json:"created_by"`
	Utime      int64             `gorm:"autoUpdateTime" json:"utime"`
	Ctime      int64             `gorm:"autoCreateTime" json:"ctime"`
}

// Message 交流消息模型
type Message struct {
	ID                uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID             uuid.UUID            `gorm:"type:uuid;not null" json:"app_id"`
	ConversationID    uuid.UUID            `gorm:"type:uuid;not null;index:message_conversation_id_idx" json:"conversation_id"`
	InvokeFrom        consts.InvokeFrom    `gorm:"size:255;not null;default:''" json:"invoke_from"`
	CreatedBy         uuid.UUID            `gorm:"type:uuid;not null;index:message_created_by_idx" json:"created_by"`
	Query             string               `gorm:"type:text;not null;default:''" json:"query"`
	ImageUrls         []string             `gorm:"type:text[];not null;default:'{}'" json:"image_urls"`
	MessageTokenCount int                  `gorm:"not null;default:0" json:"message_token_count"`
	MessageUnitPrice  float64              `gorm:"type:decimal(10,7);not null;default:0.0" json:"message_unit_price"`
	MessagePriceUnit  float64              `gorm:"type:decimal(10,4);not null;default:0.0" json:"message_price_unit"`
	Answer            string               `gorm:"type:text;not null;default:''" json:"answer"`
	AnswerTokenCount  int                  `gorm:"not null;default:0" json:"answer_token_count"`
	AnswerUnitPrice   float64              `gorm:"type:decimal(10,7);not null;default:0.0" json:"answer_unit_price"`
	AnswerPriceUnit   float64              `gorm:"type:decimal(10,4);not null;default:0.0" json:"answer_price_unit"`
	Latency           float64              `gorm:"not null;default:0.0" json:"latency"`
	IsDeleted         bool                 `gorm:"not null;default:false" json:"is_deleted"`
	Status            consts.MessageStatus `gorm:"size:255;not null;default:''" json:"status"`
	Error             string               `gorm:"type:text;not null;default:''" json:"error"`
	TotalTokenCount   int                  `gorm:"not null;default:0" json:"total_token_count"`
	TotalPrice        float64              `gorm:"type:decimal(10,7);not null;default:0.0" json:"total_price"`
	Utime             int64                `gorm:"autoUpdateTime" json:"utime"`
	Ctime             int64                `gorm:"autoCreateTime" json:"ctime"`
}

// AgentThought 智能体消息推理模型，用于记录Agent生成最终消息答案时
type AgentThought struct {
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID             uuid.UUID        `gorm:"type:uuid;not null;index:message_agent_thought_app_id_idx" json:"app_id"`
	ConversationID    uuid.UUID        `gorm:"type:uuid;not null;index:message_agent_thought_conversation_id_idx" json:"conversation_id"`
	MessageID         uuid.UUID        `gorm:"type:uuid;not null;index:message_agent_thought_message_id_idx" json:"message_id"`
	InvokeFrom        string           `gorm:"size:255;not null;default:''" json:"invoke_from"`
	CreatedBy         uuid.UUID        `gorm:"type:uuid;not null" json:"created_by"`
	Position          int              `gorm:"not null;default:0" json:"position"`
	Event             string           `gorm:"size:255;not null;default:''" json:"event"`
	Thought           string           `gorm:"type:text;not null;default:''" json:"thought"`
	Observation       string           `gorm:"type:text;not null;default:''" json:"observation"`
	Tool              string           `gorm:"type:text;not null;default:''" json:"tool"`
	ToolInput         map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"tool_input"`
	Message           []map[string]any `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"message"`
	MessageTokenCount int              `gorm:"not null;default:0" json:"message_token_count"`
	MessageUnitPrice  float64          `gorm:"type:decimal(10,7);not null;default:0.0" json:"message_unit_price"`
	MessagePriceUnit  float64          `gorm:"type:decimal(10,4);not null;default:0" json:"message_price_unit"`
	Answer            string           `gorm:"type:text;not null;default:''" json:"answer"`
	AnswerTokenCount  int              `gorm:"not null;default:0" json:"answer_token_count"`
	AnswerUnitPrice   float64          `gorm:"type:decimal(10,7);not null;default:0.0" json:"answer_unit_price"`
	AnswerPriceUnit   float64          `gorm:"type:decimal(10,4);not null;default:0.0" json:"answer_price_unit"`
	TotalTokenCount   int              `gorm:"not null;default:0" json:"total_token_count"`
	TotalPrice        float64          `gorm:"type:decimal(10,7);not null;default:0.0" json:"total_price"`
	Latency           float64          `gorm:"not null;default:0.0" json:"latency"`
	Utime             int64            `gorm:"autoUpdateTime" json:"utime"`
	Ctime             int64            `gorm:"autoCreateTime" json:"ctime"`
}
