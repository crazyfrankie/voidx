package resp

import "github.com/google/uuid"

// AssistantAgentMessageResp 辅助智能体消息响应
type AssistantAgentMessageResp struct {
	ID              uuid.UUID                   `json:"id"`
	ConversationID  uuid.UUID                   `json:"conversation_id"`
	Query           string                      `json:"query"`
	ImageUrls       []string                    `json:"image_urls"`
	Answer          string                      `json:"answer"`
	TotalTokenCount int                         `json:"total_token_count"`
	Latency         float64                     `json:"latency"`
	AgentThoughts   []AssistantAgentThoughtResp `json:"agent_thoughts"`
	Ctime           int64                       `json:"ctime"`
}

// AssistantAgentThoughtResp 辅助智能体思考过程响应
type AssistantAgentThoughtResp struct {
	ID          uuid.UUID      `json:"id"`
	Position    int            `json:"position"`
	Event       string         `json:"event"`
	Thought     string         `json:"thought"`
	Observation string         `json:"observation"`
	Tool        string         `json:"tool"`
	ToolInput   map[string]any `json:"tool_input"`
	Latency     float64        `json:"latency"`
	Ctime       int64          `json:"ctime"`
}

// AssistantAgentChatEvent 辅助Agent会话事件响应（流式）
type AssistantAgentChatEvent struct {
	ID              string  `json:"id"`
	ConversationID  string  `json:"conversation_id"`
	MessageID       string  `json:"message_id"`
	TaskID          string  `json:"task_id"`
	Event           string  `json:"event"`
	Thought         string  `json:"thought"`
	Observation     string  `json:"observation"`
	Tool            string  `json:"tool"`
	ToolInput       string  `json:"tool_input"`
	Answer          string  `json:"answer"`
	Latency         float64 `json:"latency"`
	TotalTokenCount int     `json:"total_token_count"`
}
