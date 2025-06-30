package service

import (
	"context"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"

	"github.com/crazyfrankie/voidx/internal/repository/dao"
	"github.com/crazyfrankie/voidx/pkg/langchainx"
)

type ChatService struct {
	llm *openai.LLM
	dao *dao.ChatDao
}

func NewChatService(llm *openai.LLM, dao *dao.ChatDao) *ChatService {
	return &ChatService{llm: llm, dao: dao}
}

func (s *ChatService) Chat(ctx context.Context, query string) (string, error) {
	chatHis, err := langchainx.NewFileChatHistory("storage/memory/chat_history.json")
	if err != nil {
		return "", err
	}

	mem := memory.NewConversationWindowBuffer(3,
		memory.WithInputKey("input"),
		memory.WithOutputKey("text"),
		memory.WithReturnMessages(true),
		memory.WithChatHistory(chatHis))

	chain := chains.NewConversation(s.llm, mem)
	res, err := chains.Call(ctx, chain, map[string]any{
		"input": query,
	})
	if err != nil {
		return "", err
	}

	return res["text"].(string), nil
}
