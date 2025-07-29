package resp

import "github.com/google/uuid"

type GetConversationMessagePageResp struct {
	ID              uuid.UUID                                `json:"id"`
	ConversationID  uuid.UUID                                `json:"conversation_id"`
	Query           string                                   `json:"query"`
	ImageUrls       []string                                 `json:"image_urls"`
	Answer          string                                   `json:"answer"`
	Latency         float64                                  `json:"latency"`
	TotalTokenCount int                                      `json:"total_token_count"`
	AgentThoughts   []GetConversationMessagePageAgentThought `json:"agent_thoughts"`
	Ctime           int64                                    `json:"ctime"`
}

type GetConversationMessagePageAgentThought struct {
	ID          uuid.UUID      `json:"id"`
	Position    int            `json:"position"`
	Event       string         `json:"event"`
	Thought     string         `json:"thought"`
	Observation string         `json:"observation"`
	Tool        string         `json:"tool"`
	ToolInput   map[string]any `json:"tool_input"`
	Latency     int            `json:"latency"`
	Ctime       int64          `json:"ctime"`
}
