package resp

import (
	"github.com/google/uuid"
)

// AppResp 应用响应
type AppResp struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Icon                string    `json:"icon"`
	Description         string    `json:"description"`
	Status              string    `json:"status"`
	AppConfigID         uuid.UUID `json:"app_config_id"`
	DraftAppConfigID    uuid.UUID `json:"draft_app_config_id"`
	DebugConversationID uuid.UUID `json:"debug_conversation_id"`
	Token               string    `json:"jwt"`
	Ctime               int64     `json:"ctime"`
	Utime               int64     `json:"utime"`
}

// MessageResp 消息响应
type MessageResp struct {
	ID             uuid.UUID      `json:"id"`
	ConversationID uuid.UUID      `json:"conversation_id"`
	AppID          uuid.UUID      `json:"app_id"`
	InvokeFrom     string         `json:"invoke_from"`
	CreatedBy      uuid.UUID      `json:"created_by"`
	Query          string         `json:"query"`
	ImageUrls      []string       `json:"image_urls"`
	Answer         string         `json:"answer"`
	Status         string         `json:"status"`
	AgentThoughts  []AgentThought `json:"agent_thoughts"`
	Ctime          int64          `json:"ctime"`
	Utime          int64          `json:"utime"`
}

// AgentThought 智能体思考过程
type AgentThought struct {
	ID              uuid.UUID `json:"id"`
	MessageID       uuid.UUID `json:"message_id"`
	Event           string    `json:"event"`
	Thought         string    `json:"thought"`
	Observation     string    `json:"observation"`
	Tool            string    `json:"tool"`
	ToolInput       string    `json:"tool_input"`
	Answer          string    `json:"answer"`
	TotalTokenCount int       `json:"total_token_count"`
	TotalPrice      float64   `json:"total_price"`
	Latency         float64   `json:"latency"`
	Ctime           int64     `json:"ctime"`
}

// Paginator 分页器
type Paginator struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalPage   int `json:"total_page"`
	TotalRecord int `json:"total_record"`
}

// PublishedConfigResp 发布配置响应
type PublishedConfigResp struct {
	WebApp struct {
		Token  string `json:"jwt"`
		Status string `json:"status"`
	} `json:"web_app"`
}

type DebugConversationMessageResp struct {
	ID              uuid.UUID      `json:"id"`
	ConversationID  uuid.UUID      `json:"conversation_id"`
	AgentThoughts   []AgentThought `json:"agent_thoughts"`
	Query           string         `json:"query"`
	Answer          string         `json:"answer"`
	TotalTokenCount int            `json:"total_token_count"`
	TotalPrice      float64        `json:"total_price"`
	Latency         float64        `json:"latency"`
	Ctime           int64          `json:"ctime"`
}

// GetAppsWithPageResp 获取应用分页列表响应
type GetAppsWithPageResp struct {
	List      []AppWithPage `json:"list"`
	Paginator Paginator     `json:"paginator"`
}

type AppWithPage struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Icon         string         `json:"icon"`
	Description  string         `json:"description"`
	PresetPrompt string         `json:"preset_prompt"`
	ModelConfig  map[string]any `json:"model_config"`
	Status       string         `json:"status"`
	Ctime        int64          `json:"ctime"`
	Utime        int64          `json:"utime"`
}

// AppDebugChatResp 应用调试对话响应
type AppDebugChatResp struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	MessageID      uuid.UUID `json:"message_id"`
}

type AppDraftConfigResp struct {
	Id                   uuid.UUID        `json:"id"`
	ModelConfig          map[string]any   `json:"model_config"`
	DialogRound          int              `json:"dialog_round"`
	PresetPrompt         string           `json:"preset_prompt"`
	Tools                []map[string]any `json:"tools"`
	Workflows            []map[string]any `json:"workflows"`
	Datasets             []map[string]any `json:"datasets"`
	RetrievalConfig      map[string]any   `json:"retrieval_config"`
	LongTermMemory       map[string]any   `json:"long_term_memory"`
	OpeningStatement     string           `json:"opening_statement"`
	OpeningQuestions     []string         `json:"opening_questions"`
	SpeechToText         map[string]any   `json:"speech_to_text"`
	TextToSpeech         map[string]any   `json:"text_to_speech"`
	SuggestedAfterAnswer map[string]any   `json:"suggested_after_answer"`
	ReviewConfig         map[string]any   `json:"review_config"`
}

type GetPublishHistoriesWithPageResp struct {
	ID      uuid.UUID `json:"id"`
	Version int       `json:"version"`
	Ctime   int64     `json:"ctime"`
}
