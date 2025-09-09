package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

// BaseAgent represents the interface for all agent implementations
type BaseAgent interface {
	// Invoke executes the agent with the given input and returns a result
	Invoke(ctx context.Context, input entities.AgentState) (entities.AgentResult, error)

	// Stream executes the agent and returns a channel of thoughts
	Stream(ctx context.Context, input entities.AgentState) (<-chan *entities.AgentThought, error)
}

// baseAgentImpl provides a base implementation of the BaseAgent interface
type baseAgentImpl struct {
	llm          model.BaseChatModel
	agentConfig  *entities.AgentConfig
	queueFactory *AgentQueueManagerFactory
}

// NewBaseAgent creates a new base agent implementation
func NewBaseAgent(llm model.BaseChatModel, config *entities.AgentConfig, queueFactory *AgentQueueManagerFactory) BaseAgent {
	return &baseAgentImpl{
		llm:          llm,
		agentConfig:  config,
		queueFactory: queueFactory,
	}
}

// Invoke implements the BaseAgent interface
func (b *baseAgentImpl) Invoke(ctx context.Context, input entities.AgentState) (entities.AgentResult, error) {
	// Extract query and image URLs from input
	content := ""
	query := ""
	var imageURLs []string

	if len(input.Messages) > 0 {
		lastMsg := input.Messages[len(input.Messages)-1]
		content = lastMsg.Content
		query = content
		imageURLs = extractImageURLsFromMessage(lastMsg)
	}

	agentResult := entities.AgentResult{
		Query:     query,
		ImageURLs: imageURLs,
		Status:    "normal",
	}

	agentThoughts := make(map[string]*entities.AgentThought)

	// Get streaming channel
	thoughtChan, err := b.Stream(ctx, input)
	if err != nil {
		return agentResult, fmt.Errorf("failed to start streaming: %w", err)
	}

	// Process streaming thoughts
	for agentThought := range thoughtChan {
		eventID := agentThought.ID.String()

		// Skip ping events
		if agentThought.Event != entities.EventPing {
			// Handle agent_message events (accumulative)
			if agentThought.Event == entities.EventAgentMessage {
				if existing, exists := agentThoughts[eventID]; exists {
					// Accumulate message content
					existing.Thought = existing.Thought + agentThought.Thought
					existing.Answer = existing.Answer + agentThought.Answer
					existing.Latency = agentThought.Latency
					agentThoughts[eventID] = existing
				} else {
					// Initialize message event
					agentThoughts[eventID] = agentThought
				}
				// Update agent result answer
				agentResult.Answer += agentThought.Answer
			} else {
				// Handle other event types (overwrite)
				agentThoughts[eventID] = agentThought

				// Handle error/stop/timeout events
				if agentThought.Event == entities.EventStop ||
					agentThought.Event == entities.EventTimeout ||
					agentThought.Event == entities.EventError {
					agentResult.Status = agentThought.Event
					if agentThought.Event == entities.EventError {
						agentResult.Error = agentThought.Observation
					} else {
						agentResult.Error = ""
					}
				}
			}
		}
	}

	// Convert thoughts map to slice
	agentResult.AgentThoughts = make([]entities.AgentThought, 0, len(agentThoughts))
	for _, agentThought := range agentThoughts {
		agentResult.AgentThoughts = append(agentResult.AgentThoughts, *agentThought)
	}

	// Set message from agent thoughts
	for _, agentThought := range agentThoughts {
		if agentThought.Event == entities.EventAgentMessage && len(agentThought.Message) > 0 {
			agentResult.Message = agentThought.Message
			break
		}
	}

	// Calculate total latency
	for _, agentThought := range agentResult.AgentThoughts {
		agentResult.Latency += agentThought.Latency
	}

	return agentResult, nil
}

// Stream implements the BaseAgent interface
func (b *baseAgentImpl) Stream(ctx context.Context, input entities.AgentState) (<-chan *entities.AgentThought, error) {
	// Initialize task ID and other fields
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}
	if input.History == nil {
		input.History = make([]*schema.Message, 0)
	}
	if input.IterationCount == 0 {
		input.IterationCount = 0
	}

	// Create queue manager for this task
	queueManager := b.queueFactory.CreateManager(b.agentConfig.UserID, b.agentConfig.InvokeFrom)
	defer queueManager.Close()

	// Create queue for this task
	thoughtChan, err := queueManager.Listen(ctx, input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// Start processing in background
	go func() {
		defer func() {
			// Send end event
			queueManager.Publish(input.TaskID, &entities.AgentThought{
				ID:     uuid.New(),
				TaskID: input.TaskID,
				Event:  entities.EventAgentEnd,
			})
		}()

		// Base implementation - simple echo
		if len(input.Messages) > 0 {
			lastMsg := input.Messages[len(input.Messages)-1]
			content := extractQueryFromMessage(lastMsg)

			queueManager.Publish(input.TaskID, &entities.AgentThought{
				ID:      uuid.New(),
				TaskID:  input.TaskID,
				Event:   entities.EventAgentMessage,
				Thought: content,
				Answer:  content,
				Latency: 0.1,
			})
		}
	}()

	return thoughtChan, nil
}

// extractQueryFromMessage extracts the text content from a message
func extractQueryFromMessage(msg *schema.Message) string {
	return msg.Content
}

// extractImageURLsFromMessage extracts image URLs from a message
func extractImageURLsFromMessage(msg *schema.Message) []string {
	var imageURLs []string

	// Check MultiContent for image URLs
	for _, part := range msg.MultiContent {
		if part.Type == schema.ChatMessagePartTypeImageURL && part.ImageURL != nil {
			if part.ImageURL.URL != "" {
				imageURLs = append(imageURLs, part.ImageURL.URL)
			}
			if part.ImageURL.URI != "" {
				imageURLs = append(imageURLs, part.ImageURL.URI)
			}
		}
	}

	// Simple check for image URLs in content
	content := msg.Content
	if strings.Contains(content, "http") && (strings.Contains(content, ".jpg") ||
		strings.Contains(content, ".png") || strings.Contains(content, ".gif") ||
		strings.Contains(content, ".jpeg")) {
		words := strings.Fields(content)
		for _, word := range words {
			if strings.HasPrefix(word, "http") && (strings.Contains(word, ".jpg") ||
				strings.Contains(word, ".png") || strings.Contains(word, ".gif") ||
				strings.Contains(word, ".jpeg")) {
				imageURLs = append(imageURLs, word)
			}
		}
	}

	return imageURLs
}
