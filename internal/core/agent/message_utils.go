package agent

import (
	"github.com/tmc/langchaingo/llms"
)

// MessagesToDict converts chat messages to a dictionary format
func MessagesToDict(messages []llms.ChatMessage) []map[string]any {
	result := make([]map[string]any, len(messages))
	for i, msg := range messages {
		msgMap := map[string]any{
			"type":    msg.GetType(),
			"content": msg.GetContent(),
		}
		result[i] = msgMap
	}
	return result
}

// MessageContentToDict converts message content to a dictionary format
func MessageContentToDict(messages []llms.MessageContent) []map[string]any {
	result := make([]map[string]any, len(messages))
	for i, msg := range messages {
		data, err := msg.MarshalJSON()
		if err != nil {
			continue
		}
		msgMap := map[string]any{
			"type":    msg.Role,
			"content": string(data),
		}
		result[i] = msgMap
	}
	return result
}
