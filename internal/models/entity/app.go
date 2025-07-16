package entity

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type App struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID           uuid.UUID  `gorm:"type:uuid;index:idx_app_account_id;not null" json:"account_id"`
	AppConfigID         *uuid.UUID `gorm:"type:uuid" json:"app_config_id"`         // 发布配置ID，可为空
	DraftAppConfigID    *uuid.UUID `gorm:"type:uuid" json:"draft_app_config_id"`   // 草稿配置ID，可为空
	DebugConversationID *uuid.UUID `gorm:"type:uuid" json:"debug_conversation_id"` // 调试会话ID，可为空
	Name                string     `gorm:"size:255;not null" json:"name"`
	Icon                string     `gorm:"size:255;not null" json:"icon"`
	Description         string     `gorm:"type:text;not null" json:"description"`
	Status              string     `gorm:"type:varchar(100);not null" json:"status"` // draft, published
	Token               string     `gorm:"size:255" json:"jwt"`                      // WebApp凭证标识
	Utime               int64      `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime               int64      `gorm:"autoCreateTime:milli" json:"ctime"`
}

// AppConfigVersion 应用配置版本模型
type AppConfigVersion struct {
	ID         uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID      uuid.UUID              `gorm:"type:uuid;index;not null" json:"app_id"`
	Version    int                    `gorm:"not null" json:"version"`
	ConfigType string                 `gorm:"type:varchar(100);not null" json:"config_type"` // draft, published
	Config     map[string]interface{} `gorm:"type:jsonb" json:"config"`
	Utime      int64                  `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime      int64                  `gorm:"autoCreateTime:milli" json:"ctime"`
}

type AppConfig struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID                uuid.UUID      `gorm:"type:uuid;index;not null" json:"app_id"`
	ModelConfig          datatypes.JSON `gorm:"type:jsonb" json:"model_config"`
	DialogRound          int            `gorm:"not null;default:10" json:"dialog_round"`
	PresetPrompt         string         `gorm:"type:text" json:"preset_prompt"`
	Tools                datatypes.JSON `gorm:"type:jsonb" json:"tools"`
	Workflows            datatypes.JSON `gorm:"type:jsonb" json:"workflows"`
	RetrievalConfig      datatypes.JSON `gorm:"type:jsonb" json:"retrieval_config"`
	LongTermMemory       datatypes.JSON `gorm:"type:jsonb" json:"long_term_memory"`
	OpeningStatement     string         `gorm:"type:text" json:"opening_statement"`
	OpeningQuestions     datatypes.JSON `gorm:"type:jsonb" json:"opening_questions"`
	SpeechToText         datatypes.JSON `gorm:"type:jsonb" json:"speech_to_text"`
	TextToSpeech         datatypes.JSON `gorm:"type:jsonb" json:"text_to_speech"`
	SuggestedAfterAnswer datatypes.JSON `gorm:"type:jsonb" json:"suggested_after_answer"`
	ReviewConfig         datatypes.JSON `gorm:"type:jsonb" json:"review_config"`
	Ctime                int64          `gorm:"autoCreateTime:milli" json:"ctime"`
}

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;index;not null" json:"account_id"`
	AppID     uuid.UUID `gorm:"type:uuid;index;not null" json:"app_id"`
	Summary   string    `gorm:"type:text" json:"summary"` // 长期记忆
	Utime     int64     `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime:milli" json:"ctime"`
}

type Message struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID      `gorm:"type:uuid;index;not null" json:"conversation_id"`
	AppID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"app_id"`
	InvokeFrom     string         `gorm:"type:varchar(100);not null" json:"invoke_from"` // debugger, web_app
	CreatedBy      uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Query          string         `gorm:"type:text;not null" json:"query"`
	ImageUrls      []string       `gorm:"type:text[]" json:"image_urls"`
	Answer         string         `gorm:"type:text" json:"answer"`
	Status         string         `gorm:"type:varchar(100);not null" json:"status"` // normal, stop
	AgentThoughts  []AgentThought `gorm:"foreignKey:MessageID" json:"agent_thoughts"`
	Utime          int64          `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime          int64          `gorm:"autoCreateTime:milli" json:"ctime"`
}

// AgentThought 智能体思考过程模型
type AgentThought struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID       uuid.UUID `gorm:"type:uuid;index;not null" json:"message_id"`
	Event           string    `gorm:"type:varchar(100);not null" json:"event"`
	Thought         string    `gorm:"type:text" json:"thought"`
	Observation     string    `gorm:"type:text" json:"observation"`
	Tool            string    `gorm:"type:varchar(255)" json:"tool"`
	ToolInput       string    `gorm:"type:text" json:"tool_input"`
	Answer          string    `gorm:"type:text" json:"answer"`
	TotalTokenCount int       `gorm:"not null;default:0" json:"total_token_count"`
	TotalPrice      float64   `gorm:"not null;default:0" json:"total_price"`
	Latency         float64   `gorm:"not null;default:0" json:"latency"`
	Ctime           int64     `gorm:"autoCreateTime" json:"ctime"`
}

// AppDatasetJoin 应用关联知识库模型
type AppDatasetJoin struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID     uuid.UUID `gorm:"type:uuid;index:idx_app_dataset;not null" json:"app_id"`
	DatasetID uuid.UUID `gorm:"type:uuid;index:idx_app_dataset;not null" json:"dataset_id"`
	Utime     int64     `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime:milli" json:"ctime"`
}
