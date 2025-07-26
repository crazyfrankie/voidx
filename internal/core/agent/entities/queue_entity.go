package entities

import (
	"time"

	"github.com/google/uuid"
)

// QueueEvent represents different types of agent events
type QueueEvent string

const (
	// EventPing represents a ping event to check agent status
	EventPing QueueEvent = "ping"

	// EventAgentMessage represents an agent message event
	EventAgentMessage QueueEvent = "agent_message"

	// EventAgentEnd represents an agent end event
	EventAgentEnd QueueEvent = "agent_end"

	// EventLongTermMemoryRecall represents a long-term memory recall event
	EventLongTermMemoryRecall QueueEvent = "long_term_memory_recall"

	// EventStop represents a stop event
	EventStop QueueEvent = "stop"

	// EventTimeout represents a timeout event
	EventTimeout QueueEvent = "timeout"

	// EventError represents an error event
	EventError QueueEvent = "error"

	// EventAgentThought represents an agent thought event
	EventAgentThought QueueEvent = "agent_thought"

	// EventAgentAction represents an agent action event
	EventAgentAction QueueEvent = "agent_action"

	// EventDatasetRetrieval represents a dataset retrieval event
	EventDatasetRetrieval QueueEvent = "dataset_retrieval"
)

// AgentThought represents a single thought or action by the agent
type AgentThought struct {
	// ID represents the unique identifier for this thought
	ID uuid.UUID `json:"id"`

	// TaskID represents the associated task identifier
	TaskID uuid.UUID `json:"task_id"`

	// Event represents the type of event
	Event QueueEvent `json:"event"`

	// Thought represents the agent's reasoning process
	Thought string `json:"thought,omitempty"`

	// Message represents the chat messages involved
	Message []map[string]any `json:"message,omitempty"`

	// Answer represents the agent's response
	Answer string `json:"answer,omitempty"`

	// Observation represents additional information or tool output
	Observation string `json:"observation,omitempty"`

	// Tool represents the name of the tool being used
	Tool string `json:"tool,omitempty"`

	// ToolInput represents the input parameters for the tool
	ToolInput map[string]any `json:"tool_input,omitempty"`

	// MessageTokenCount represents the token count of the input message
	MessageTokenCount int `json:"message_token_count,omitempty"`

	// MessageUnitPrice represents the unit price for input tokens
	MessageUnitPrice float64 `json:"message_unit_price,omitempty"`

	// MessagePriceUnit represents the price unit for input tokens
	MessagePriceUnit string `json:"message_price_unit,omitempty"`

	// AnswerTokenCount represents the token count of the output answer
	AnswerTokenCount int `json:"answer_token_count,omitempty"`

	// AnswerUnitPrice represents the unit price for output tokens
	AnswerUnitPrice float64 `json:"answer_unit_price,omitempty"`

	// AnswerPriceUnit represents the price unit for output tokens
	AnswerPriceUnit string `json:"answer_price_unit,omitempty"`

	// TotalTokenCount represents the total token count
	TotalTokenCount int `json:"total_token_count,omitempty"`

	// TotalPrice represents the total price
	TotalPrice float64 `json:"total_price,omitempty"`

	// Latency represents the processing time in seconds
	Latency float64 `json:"latency"`
}

// AgentResult represents the final result of agent execution
type AgentResult struct {
	// Query represents the original query
	Query string `json:"query"`

	// ImageURLs represents any image URLs in the query
	ImageURLs []string `json:"image_urls,omitempty"`

	// Answer represents the final answer
	Answer string `json:"answer"`

	// Status represents the execution status
	Status QueueEvent `json:"status"`

	// Error represents any error message
	Error string `json:"error,omitempty"`

	// Message represents the chat messages
	Message []map[string]any `json:"message"`

	// AgentThoughts represents the sequence of thoughts
	AgentThoughts []AgentThought `json:"agent_thoughts"`

	// Latency represents the total processing time in seconds
	Latency float64 `json:"latency"`

	// CreatedAt represents the creation timestamp
	CreatedAt time.Time `json:"created_at"`
}
