package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/crazyfrankie/voidx/internal/repository/dao"
)

type ChatService struct {
	llm *openai.LLM
	dao *dao.ChatDao
}

func NewChatService(llm *openai.LLM, dao *dao.ChatDao) *ChatService {
	return &ChatService{llm: llm, dao: dao}
}

func (s *ChatService) Chat(ctx context.Context, query string) (string, error) {
	// Create messages for the chat
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant that explains complex topics step by step"),
		llms.TextParts(llms.ChatMessageTypeHuman, query),
	}

	// Generate content with streaming to see both reasoning and final answer in real-time
	completion, err := s.llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(2000),
		llms.WithTemperature(0.7),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		return "", err
	}

	// Access the reasoning content and final answer separately
	if len(completion.Choices) == 0 {
		return "", errors.New("response is empty")
	}
	choice := completion.Choices[0]

	return choice.Content, nil
}
