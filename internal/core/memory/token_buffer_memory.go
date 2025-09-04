package memory

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	model "github.com/crazyfrankie/voidx/internal/models/entity"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusNormal  MessageStatus = "normal"
	MessageStatusStop    MessageStatus = "stop"
	MessageStatusTimeout MessageStatus = "timeout"
)

func (m *TokenBufferMemory) WithConversationID(conversionID uuid.UUID) *TokenBufferMemory {
	m.conversationID = conversionID

	return m
}

func (m *TokenBufferMemory) WithLLM(modelInstance entity.BaseLanguageModel) *TokenBufferMemory {
	m.modelInstance = modelInstance

	return m
}

// TokenBufferMemory implements token-based buffer memory
type TokenBufferMemory struct {
	db             *gorm.DB
	conversationID uuid.UUID
	modelInstance  entity.BaseLanguageModel
}

// NewTokenBufferMemory creates a new TokenBufferMemory instance
func NewTokenBufferMemory(db *gorm.DB) *TokenBufferMemory {
	return &TokenBufferMemory{
		db: db,
	}
}

// GetHistoryPromptMessages returns a list of messages based on token and message count limits
func (m *TokenBufferMemory) GetHistoryPromptMessages(maxTokenLimit int, messageLimit int) ([]llms.MessageContent, error) {
	// Check if conversation exists
	if m.conversationID == uuid.Nil {
		return nil, nil
	}

	// Query messages for the conversation
	var messages []model.Message
	result := m.db.Where(
		"conversation_id = ? AND answer != '' AND is_deleted = ? AND status IN ?",
		m.conversationID,
		false,
		[]MessageStatus{MessageStatusNormal, MessageStatusStop, MessageStatusTimeout},
	).Order("ctime desc").Limit(messageLimit).Find(&messages)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query messages: %w", result.Error)
	}

	// Reverse the messages to maintain chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Convert to LangChain messages
	var promptMessages []llms.MessageContent
	for _, msg := range messages {
		// Add human message
		humanMsg := m.modelInstance.ConvertToHumanMessage(msg.Query, msg.ImageUrls)
		promptMessages = append(promptMessages, humanMsg)

		// Add AI message
		aiMsg := llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(msg.Answer)},
		}
		promptMessages = append(promptMessages, aiMsg)
	}

	// Trim messages based on token limit
	trimmedMessages, err := m.trimMessages(promptMessages, maxTokenLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to trim messages: %w", err)
	}

	return trimmedMessages, nil
}

// GetHistoryPromptText returns the conversation history as a formatted string
func (m *TokenBufferMemory) GetHistoryPromptText(
	humanPrefix string,
	aiPrefix string,
	maxTokenLimit int,
	messageLimit int,
) (string, error) {
	// Get history messages
	messages, err := m.GetHistoryPromptMessages(maxTokenLimit, messageLimit)
	if err != nil {
		return "", err
	}

	// Convert messages to string format
	return m.getBufferString(messages, humanPrefix, aiPrefix), nil
}

// trimMessages trims the message list based on token limit
func (m *TokenBufferMemory) trimMessages(messages []llms.MessageContent, maxTokens int) ([]llms.MessageContent, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// Calculate total tokens
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += m.estimateTokenCount(msg)
	}

	// If within limit, return all messages
	if totalTokens <= maxTokens {
		return messages, nil
	}

	// Trim from the beginning, keeping the most recent messages
	var trimmedMessages []llms.MessageContent
	currentTokens := 0

	// Start from the end and work backwards
	for i := len(messages) - 1; i >= 0; i-- {
		msgTokens := m.estimateTokenCount(messages[i])
		if currentTokens+msgTokens <= maxTokens {
			trimmedMessages = append([]llms.MessageContent{messages[i]}, trimmedMessages...)
			currentTokens += msgTokens
		} else {
			break
		}
	}

	// Ensure we start with human message and end with AI message if possible
	trimmedMessages = m.ensureProperMessageOrder(trimmedMessages)

	return trimmedMessages, nil
}

// estimateTokenCount provides a simple token count estimation for a message
func (m *TokenBufferMemory) estimateTokenCount(msg llms.MessageContent) int {
	totalChars := 0
	for _, part := range msg.Parts {
		if textPart, ok := part.(llms.TextContent); ok {
			totalChars += len(textPart.Text)
		}
	}
	// Rough estimation: 4 characters per token (this can be improved with actual tokenizer)
	return totalChars / 4
}

// ensureProperMessageOrder ensures messages start with human and end with AI
func (m *TokenBufferMemory) ensureProperMessageOrder(messages []llms.MessageContent) []llms.MessageContent {
	if len(messages) == 0 {
		return messages
	}

	// Find first human message
	startIdx := 0
	for i, msg := range messages {
		if msg.Role == llms.ChatMessageTypeHuman {
			startIdx = i
			break
		}
	}

	// Find last AI message
	endIdx := len(messages) - 1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llms.ChatMessageTypeAI {
			endIdx = i
			break
		}
	}

	// Ensure we have a valid range
	if startIdx <= endIdx {
		return messages[startIdx : endIdx+1]
	}

	return messages
}

// getBufferString converts chat messages to a formatted string (improved version)
func (m *TokenBufferMemory) getBufferString(messages []llms.MessageContent, humanPrefix, aiPrefix string) string {
	var buffer strings.Builder

	for i, msg := range messages {
		if i > 0 {
			buffer.WriteString("\n\n")
		}

		// Extract text content from message parts
		var content string
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				content += textPart.Text
			}
		}

		switch msg.Role {
		case llms.ChatMessageTypeHuman:
			buffer.WriteString(fmt.Sprintf("%s: %s", humanPrefix, content))
		case llms.ChatMessageTypeAI:
			buffer.WriteString(fmt.Sprintf("%s: %s", aiPrefix, content))
		case llms.ChatMessageTypeSystem:
			buffer.WriteString(fmt.Sprintf("System: %s", content))
		}
	}

	return buffer.String()
}
