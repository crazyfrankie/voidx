package langchainx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

// MessageType 表示消息的类型
type MessageType string

const (
	// MessageTypeHuman 表示人类消息
	MessageTypeHuman MessageType = "human"
	// MessageTypeAI 表示AI消息
	MessageTypeAI MessageType = "ai"
	// MessageTypeSystem 表示系统消息
	MessageTypeSystem MessageType = "system"
	// MessageTypeGeneric 表示通用消息
	MessageTypeGeneric MessageType = "generic"
)

// StoredMessage 是存储在文件中的消息格式
type StoredMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
	Name    string      `json:"name,omitempty"`
}

// FileChatHistory 实现了基于文件的聊天历史记录
type FileChatHistory struct {
	filePath string
	mutex    sync.Mutex
}

// NewFileChatHistory 创建一个新的基于文件的聊天历史记录
func NewFileChatHistory(filePath string) (*FileChatHistory, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 如果文件不存在，创建一个空的JSON数组文件
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			return nil, fmt.Errorf("创建历史记录文件失败: %w", err)
		}
	}

	return &FileChatHistory{
		filePath: filePath,
	}, nil
}

// loadMessages 从文件加载消息
func (f *FileChatHistory) loadMessages() ([]StoredMessage, error) {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("读取历史记录文件失败: %w", err)
	}

	var messages []StoredMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("解析历史记录文件失败: %w", err)
	}

	return messages, nil
}

// saveMessages 保存消息到文件
func (f *FileChatHistory) saveMessages(messages []StoredMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := os.WriteFile(f.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入历史记录文件失败: %w", err)
	}

	return nil
}

// convertToStoredMessage 将 ChatMessage 转换为 StoredMessage
func convertToStoredMessage(message llms.ChatMessage) StoredMessage {
	var msgType MessageType
	var name string

	switch msg := message.(type) {
	case llms.HumanChatMessage:
		msgType = MessageTypeHuman
	case llms.AIChatMessage:
		msgType = MessageTypeAI
	case llms.SystemChatMessage:
		msgType = MessageTypeSystem
	case llms.ChatMessage:
		msgType = MessageTypeGeneric
		name = string(msg.GetType())
	}

	return StoredMessage{
		Type:    msgType,
		Content: message.GetContent(),
		Name:    name,
	}
}

// convertToChatMessage 将 StoredMessage 转换为 ChatMessage
func convertToChatMessage(message StoredMessage) llms.ChatMessage {
	switch message.Type {
	case MessageTypeHuman:
		return llms.HumanChatMessage{Content: message.Content}
	case MessageTypeAI:
		return llms.AIChatMessage{Content: message.Content}
	case MessageTypeSystem:
		return llms.SystemChatMessage{Content: message.Content}
	default:
		return llms.GenericChatMessage{
			Role:    message.Name,
			Content: message.Content,
		}
	}
}

// AddMessage 添加一条消息到历史记录
func (f *FileChatHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	messages, err := f.loadMessages()
	if err != nil {
		return err
	}

	storedMessage := convertToStoredMessage(message)
	messages = append(messages, storedMessage)

	return f.saveMessages(messages)
}

// AddUserMessage 添加一条用户消息到历史记录
func (f *FileChatHistory) AddUserMessage(ctx context.Context, message string) error {
	return f.AddMessage(ctx, llms.HumanChatMessage{Content: message})
}

// AddAIMessage 添加一条AI消息到历史记录
func (f *FileChatHistory) AddAIMessage(ctx context.Context, message string) error {
	return f.AddMessage(ctx, llms.AIChatMessage{Content: message})
}

// Clear 清空历史记录
func (f *FileChatHistory) Clear(ctx context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.saveMessages([]StoredMessage{})
}

// Messages 获取所有历史消息
func (f *FileChatHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	storedMessages, err := f.loadMessages()
	if err != nil {
		return nil, err
	}

	messages := make([]llms.ChatMessage, len(storedMessages))
	for i, msg := range storedMessages {
		messages[i] = convertToChatMessage(msg)
	}

	return messages, nil
}

// SetMessages 设置历史消息
func (f *FileChatHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	storedMessages := make([]StoredMessage, len(messages))
	for i, msg := range messages {
		storedMessages[i] = convertToStoredMessage(msg)
	}

	return f.saveMessages(storedMessages)
}
