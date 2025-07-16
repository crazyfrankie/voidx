package resp

import (
	"github.com/google/uuid"
)

// AppResp 应用响应
type AppResp struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Icon                string     `json:"icon"`
	Description         string     `json:"description"`
	Status              string     `json:"status"`
	AppConfigID         *uuid.UUID `json:"app_config_id"`
	DraftAppConfigID    *uuid.UUID `json:"draft_app_config_id"`
	DebugConversationID *uuid.UUID `json:"debug_conversation_id"`
	Token               string     `json:"jwt"`
	Ctime               int64      `json:"ctime"`
	Utime               int64      `json:"utime"`
}

// AppConfigVersionResp 应用配置版本响应
type AppConfigVersionResp struct {
	ID         uuid.UUID `json:"id"`
	AppID      uuid.UUID `json:"app_id"`
	Version    int       `json:"version"`
	ConfigType string    `json:"config_type"`
	Ctime      int64     `json:"ctime"`
	Utime      int64     `json:"utime"`
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
