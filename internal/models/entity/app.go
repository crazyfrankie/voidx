package entity

import (
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
)

// App AI应用基础模型类
type App struct {
	ID                  uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID           uuid.UUID        `gorm:"type:uuid;not null;index:app_account_id_idx" json:"account_id"`
	AppConfigID         uuid.UUID        `gorm:"type:uuid" json:"app_config_id"`
	DraftAppConfigID    uuid.UUID        `gorm:"type:uuid" json:"draft_app_config_id"`
	DebugConversationID uuid.UUID        `gorm:"type:uuid" json:"debug_conversation_id"`
	Name                string           `gorm:"size:255;not null;default:''" json:"name"`
	Icon                string           `gorm:"size:255;not null;default:''" json:"icon"`
	Description         string           `gorm:"type:text;not null;default:''" json:"description"`
	Token               string           `gorm:"size:255;default:'';index:app_token_idx" json:"token"`
	Status              consts.AppStatus `gorm:"size:255;not null;default:''" json:"status"`
	Utime               int64            `gorm:"autoUpdateTime" json:"utime"`
	Ctime               int64            `gorm:"autoCreateTime" json:"ctime"`
}

// AppConfig 应用配置模型
type AppConfig struct {
	ID                   uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID                uuid.UUID        `gorm:"type:uuid;not null;index:app_config_app_id_idx" json:"app_id"`
	ModelConfig          map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"model_config"`
	DialogRound          int              `gorm:"not null;default:0" json:"dialog_round"`
	PresetPrompt         string           `gorm:"type:text;not null;default:''" json:"preset_prompt"`
	Tools                []map[string]any `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"tools"`
	Workflows            []string         `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"workflows"`
	RetrievalConfig      map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"retrieval_config"`
	LongTermMemory       map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"long_term_memory"`
	OpeningStatement     string           `gorm:"type:text;not null;default:''" json:"opening_statement"`
	OpeningQuestions     []string         `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"opening_questions"`
	SpeechToText         map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"speech_to_text"`
	TextToSpeech         map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"text_to_speech"`
	SuggestedAfterAnswer map[string]any   `gorm:"type:jsonb;not null;default:'{\"enable\": true}'::jsonb" json:"suggested_after_answer"`
	ReviewConfig         map[string]any   `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"review_config"`
	Utime                int64            `gorm:"autoUpdateTime" json:"utime"`
	Ctime                int64            `gorm:"autoCreateTime" json:"ctime"`
}

// AppConfigVersion 应用配置版本历史表，用于存储草稿配置+历史发布配置
type AppConfigVersion struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID                uuid.UUID            `gorm:"type:uuid;not null;index:app_config_version_app_id_idx" json:"app_id"`
	ModelConfig          map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"model_config"`
	DialogRound          int                  `gorm:"not null;default:0" json:"dialog_round"`
	PresetPrompt         string               `gorm:"type:text;not null;default:''" json:"preset_prompt"`
	Tools                []map[string]any     `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"tools"`
	Workflows            []string             `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"workflows"`
	Datasets             []string             `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"datasets"`
	RetrievalConfig      map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"retrieval_config"`
	LongTermMemory       map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"long_term_memory"`
	OpeningStatement     string               `gorm:"type:text;not null;default:''" json:"opening_statement"`
	OpeningQuestions     []string             `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"opening_questions"`
	SpeechToText         map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"speech_to_text"`
	TextToSpeech         map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"text_to_speech"`
	SuggestedAfterAnswer map[string]any       `gorm:"type:jsonb;not null;default:'{\"enable\": true}'::jsonb" json:"suggested_after_answer"`
	ReviewConfig         map[string]any       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"review_config"`
	Version              int                  `gorm:"not null;default:0" json:"version"`
	ConfigType           consts.AppConfigType `gorm:"size:255;not null;default:''" json:"config_type"`
	Utime                int64                `gorm:"autoUpdateTime" json:"utime"`
	Ctime                int64                `gorm:"autoCreateTime" json:"ctime"`
}

// AppDatasetJoin 应用知识库关联表模型
type AppDatasetJoin struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID     uuid.UUID `gorm:"type:uuid;not null;index:app_dataset_join_app_id_dataset_id_idx,composite:app_dataset" json:"app_id"`
	DatasetID uuid.UUID `gorm:"type:uuid;not null;index:app_dataset_join_app_id_dataset_id_idx,composite:app_dataset" json:"dataset_id"`
	Utime     int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime" json:"ctime"`
}
