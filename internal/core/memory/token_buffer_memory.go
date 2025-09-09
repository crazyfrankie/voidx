package memory

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/cloudwego/eino/schema"

	model "github.com/crazyfrankie/voidx/internal/models/entity"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusNormal  MessageStatus = "normal"
	MessageStatusStop    MessageStatus = "stop"
	MessageStatusTimeout MessageStatus = "timeout"
)

// TokenBufferMemory implements token-based buffer memory using eino schema
type TokenBufferMemory struct {
	db             *gorm.DB
	conversationID uuid.UUID
}

// NewTokenBufferMemory creates a new TokenBufferMemory instance
func NewTokenBufferMemory(db *gorm.DB) *TokenBufferMemory {
	return &TokenBufferMemory{
		db: db,
	}
}

// WithConversationID sets the conversation ID for this memory instance
func (m *TokenBufferMemory) WithConversationID(conversationID uuid.UUID) *TokenBufferMemory {
	m.conversationID = conversationID
	return m
}

// GetHistoryPromptMessages returns a list of eino messages based on token and message count limits
func (m *TokenBufferMemory) GetHistoryPromptMessages(maxTokenLimit int, messageLimit int) ([]*schema.Message, error) {
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
	).Order("created_at desc").Limit(messageLimit).Find(&messages)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query messages: %w", result.Error)
	}

	// Reverse the messages to maintain chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// Convert to eino schema messages
	var promptMessages []*schema.Message
	for _, msg := range messages {
		// Add human message
		humanMsg := m.convertToHumanMessage(msg.Query, msg.ImageUrls)
		promptMessages = append(promptMessages, humanMsg)

		// Add AI message
		aiMsg := schema.AssistantMessage(msg.Answer, nil)
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

// convertToHumanMessage converts query and image URLs to eino human message
func (m *TokenBufferMemory) convertToHumanMessage(query string, imageUrls []string) *schema.Message {
	if len(imageUrls) == 0 {
		// Simple text message
		return schema.UserMessage(query)
	}

	// Multi-content message with text and images
	var parts []schema.ChatMessagePart

	// Add text part
	if query != "" {
		parts = append(parts, schema.ChatMessagePart{
			Type: schema.ChatMessagePartTypeText,
			Text: query,
		})
	}

	// Add image parts
	for _, imageUrl := range imageUrls {
		parts = append(parts, schema.ChatMessagePart{
			Type: schema.ChatMessagePartTypeImageURL,
			ImageURL: &schema.ChatMessageImageURL{
				URL: imageUrl,
			},
		})
	}

	return &schema.Message{
		Role:         schema.User,
		Content:      query,
		MultiContent: parts,
	}
}

// trimMessages trims the message list based on token limit using eino messages
func (m *TokenBufferMemory) trimMessages(messages []*schema.Message, maxTokens int) ([]*schema.Message, error) {
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
	var trimmedMessages []*schema.Message
	currentTokens := 0

	// Start from the end and work backwards
	for i := len(messages) - 1; i >= 0; i-- {
		msgTokens := m.estimateTokenCount(messages[i])
		if currentTokens+msgTokens <= maxTokens {
			trimmedMessages = append([]*schema.Message{messages[i]}, trimmedMessages...)
			currentTokens += msgTokens
		} else {
			break
		}
	}

	// Ensure we start with human message and end with AI message if possible
	trimmedMessages = m.ensureProperMessageOrder(trimmedMessages)

	return trimmedMessages, nil
}

// estimateTokenCount provides simple token count estimation for eino messages
func (m *TokenBufferMemory) estimateTokenCount(msg *schema.Message) int {
	// Simple character-based estimation (roughly 4 characters per token)
	totalTokens := len(msg.Content) / 4

	// Add tokens for multi-content parts
	for _, part := range msg.MultiContent {
		if part.Type == schema.ChatMessagePartTypeText {
			totalTokens += len(part.Text) / 4
		}
	}

	// Add tokens for tool calls
	for _, toolCall := range msg.ToolCalls {
		totalTokens += len(toolCall.Function.Name) / 4
		totalTokens += len(toolCall.Function.Arguments) / 4
	}

	// Ensure minimum token count
	if totalTokens < 1 {
		totalTokens = 1
	}

	return totalTokens
}

// ensureProperMessageOrder ensures messages start with human and end with AI
func (m *TokenBufferMemory) ensureProperMessageOrder(messages []*schema.Message) []*schema.Message {
	if len(messages) == 0 {
		return messages
	}

	// Find first human message
	startIdx := 0
	for i, msg := range messages {
		if msg.Role == schema.User {
			startIdx = i
			break
		}
	}

	// Find last AI message
	endIdx := len(messages) - 1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.Assistant {
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

// getBufferString converts eino messages to a formatted string
func (m *TokenBufferMemory) getBufferString(messages []*schema.Message, humanPrefix, aiPrefix string) string {
	var buffer strings.Builder

	for i, msg := range messages {
		if i > 0 {
			buffer.WriteString("\n\n")
		}

		switch msg.Role {
		case schema.User:
			buffer.WriteString(fmt.Sprintf("%s: %s", humanPrefix, msg.Content))
		case schema.Assistant:
			buffer.WriteString(fmt.Sprintf("%s: %s", aiPrefix, msg.Content))
		case schema.System:
			buffer.WriteString(fmt.Sprintf("System: %s", msg.Content))
		case schema.Tool:
			buffer.WriteString(fmt.Sprintf("Tool: %s", msg.Content))
		}
	}

	return buffer.String()
}

// ClearHistory clears the conversation history
func (m *TokenBufferMemory) ClearHistory() error {
	if m.conversationID == uuid.Nil {
		return nil
	}

	// Soft delete messages by setting is_deleted = true
	result := m.db.Model(&model.Message{}).
		Where("conversation_id = ?", m.conversationID).
		Update("is_deleted", true)

	if result.Error != nil {
		return fmt.Errorf("failed to clear history: %w", result.Error)
	}

	return nil
}

// GetMessageCount returns the number of messages in the conversation
func (m *TokenBufferMemory) GetMessageCount() (int64, error) {
	if m.conversationID == uuid.Nil {
		return 0, nil
	}

	var count int64
	result := m.db.Model(&model.Message{}).
		Where("conversation_id = ? AND is_deleted = ?", m.conversationID, false).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to get message count: %w", result.Error)
	}

	return count, nil
}
