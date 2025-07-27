package resp

import "github.com/google/uuid"

// OpenAPIChatResp 开放API聊天响应
type OpenAPIChatResp struct {
	ID              uuid.UUID             `json:"id"`
	EndUserID       uuid.UUID             `json:"end_user_id"`
	ConversationID  uuid.UUID             `json:"conversation_id"`
	Query           string                `json:"query"`
	ImageUrls       []string              `json:"image_urls"`
	Answer          string                `json:"answer"`
	TotalTokenCount int                   `json:"total_token_count"`
	Latency         float64               `json:"latency"`
	AgentThoughts   []OpenAPIAgentThought `json:"agent_thoughts"`
}

// OpenAPIAgentThought 开放API智能体思考过程
type OpenAPIAgentThought struct {
	ID          string  `json:"id"`
	Event       string  `json:"event"`
	Thought     string  `json:"thought"`
	Observation string  `json:"observation"`
	Tool        string  `json:"tool"`
	ToolInput   string  `json:"tool_input"`
	Latency     float64 `json:"latency"`
	CreatedAt   int64   `json:"created_at"`
}

// OpenAPIChatEvent 开放API聊天事件（流式）
type OpenAPIChatEvent struct {
	ID             string  `json:"id"`
	EndUserID      string  `json:"end_user_id"`
	ConversationID string  `json:"conversation_id"`
	MessageID      string  `json:"message_id"`
	TaskID         string  `json:"task_id"`
	Event          string  `json:"event"`
	Thought        string  `json:"thought"`
	Observation    string  `json:"observation"`
	Tool           string  `json:"tool"`
	ToolInput      string  `json:"tool_input"`
	Answer         string  `json:"answer"`
	Latency        float64 `json:"latency"`
}
