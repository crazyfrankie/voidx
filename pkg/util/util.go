package util

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"strings"

	"github.com/crazyfrankie/voidx/pkg/errno"
)

func GetCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errno.ErrUnauthorized.AppendBizMessage("未登录")
	}

	return userID, nil
}

func GenerateHash(text string) string {
	text = text + "None"
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash[:])
}

func ConvertViaJSON(dest interface{}, src interface{}) error {
	jsonData, err := sonic.Marshal(src)
	if err != nil {
		return err
	}

	return sonic.Unmarshal(jsonData, dest)
}

func ChatMessageToMessageContent(msg llms.ChatMessage) llms.MessageContent {
	return llms.MessageContent{
		Role: msg.GetType(),
		Parts: []llms.ContentPart{
			llms.TextContent{Text: msg.GetContent()},
		},
	}
}

func MessageContentToChatMessage(mc llms.MessageContent) llms.ChatMessage {
	var contentParts []string

	for _, part := range mc.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			contentParts = append(contentParts, p.Text)
		case llms.ImageURLContent:
			contentParts = append(contentParts, p.URL)
		case llms.BinaryContent:
			contentParts = append(contentParts, p.String())
		case llms.ToolCall:
			jsonBytes, _ := sonic.Marshal(p)
			contentParts = append(contentParts, string(jsonBytes))
		}
	}

	content := strings.Join(contentParts, "\n")

	switch mc.Role {
	case llms.ChatMessageTypeHuman:
		return llms.HumanChatMessage{Content: content}
	case llms.ChatMessageTypeAI:
		return llms.AIChatMessage{Content: content}
	case llms.ChatMessageTypeSystem:
		return llms.SystemChatMessage{Content: content}
	default:
		return llms.GenericChatMessage{Content: content, Role: string(mc.Role)}
	}
}
