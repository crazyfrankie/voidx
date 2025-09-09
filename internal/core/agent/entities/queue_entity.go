package entities

import (
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// QueueEvent represents different types of events in the agent queue
type QueueEvent string

const (
	EventPing                 QueueEvent = "ping"
	EventAgentMessage         QueueEvent = "agent_message"
	EventAgentThought         QueueEvent = "agent_thought"
	EventAgentAction          QueueEvent = "agent_action"
	EventDatasetRetrieval     QueueEvent = "dataset_retrieval"
	EventLongTermMemoryRecall QueueEvent = "long_term_memory_recall"
	EventAgentEnd             QueueEvent = "agent_end"
	EventStop                 QueueEvent = "stop"
	EventTimeout              QueueEvent = "timeout"
	EventError                QueueEvent = "error"
)

// AgentThought represents a single thought or action in the agent's reasoning process
type AgentThought struct {
	ID     uuid.UUID  `json:"id"`
	TaskID uuid.UUID  `json:"task_id"`
	Event  QueueEvent `json:"event"`

	// Thought content
	Thought string `json:"thought,omitempty"`
	Answer  string `json:"answer,omitempty"`

	// Message related fields
	Message           []*schema.Message `json:"message,omitempty"`
	MessageTokenCount int               `json:"message_token_count,omitempty"`
	MessageUnitPrice  float64           `json:"message_unit_price,omitempty"`
	MessagePriceUnit  float64           `json:"message_price_unit,omitempty"`

	// Answer related fields
	AnswerTokenCount int     `json:"answer_token_count,omitempty"`
	AnswerUnitPrice  float64 `json:"answer_unit_price,omitempty"`
	AnswerPriceUnit  float64 `json:"answer_price_unit,omitempty"`

	// Tool related fields
	Tool      string                 `json:"tool,omitempty"`
	ToolInput map[string]interface{} `json:"tool_input,omitempty"`

	// Observation and error
	Observation string `json:"observation,omitempty"`

	// Statistics
	TotalTokenCount int     `json:"total_token_count,omitempty"`
	TotalPrice      float64 `json:"total_price,omitempty"`
	Latency         float64 `json:"latency,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

// AgentResult represents the final result of an agent execution
type AgentResult struct {
	Query     string   `json:"query"`
	Answer    string   `json:"answer"`
	ImageURLs []string `json:"image_urls,omitempty"`

	// Message and thoughts
	Message       []*schema.Message `json:"message,omitempty"`
	AgentThoughts []AgentThought    `json:"agent_thoughts"`

	// Status and error
	Status QueueEvent `json:"status"`
	Error  string     `json:"error,omitempty"`

	// Statistics
	Latency float64 `json:"latency"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}
