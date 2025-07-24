package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"

	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	file "github.com/crazyfrankie/voidx/pkg/langchainx/memory"
)

type AppService struct {
	repo     *repository.AppRepo
	llm      entity.BaseLanguageModel
	vecStore *vecstore.VecStoreService
}

func NewAppService(repo *repository.AppRepo, vecStore *vecstore.VecStoreService, llm entity.BaseLanguageModel) *AppService {
	return &AppService{
		repo:     repo,
		llm:      llm,
		vecStore: vecStore,
	}
}

func (s *AppService) DebugChat(ctx context.Context, appID uuid.UUID, chatReq req.DebugChatReq) (resp.AppDebugChatResp, error) {
	chatHis, err := file.NewFileChatMessageHistory(file.WithFilePath("storage/memory/chat_history.json"))
	if err != nil {
		return resp.AppDebugChatResp{}, err
	}

	mem := memory.NewConversationWindowBuffer(3,
		memory.WithInputKey("question"),
		memory.WithOutputKey("text"),
		memory.WithReturnMessages(true),
		memory.WithChatHistory(chatHis))

	chain := chains.NewConversationalRetrievalQAFromLLM(s.llm, s.vecStore.GetRetriever(10), mem)
	res, err := chains.Call(ctx, chain, map[string]any{
		"question": chatReq.Query,
	})
	if err != nil {
		return resp.AppDebugChatResp{}, err
	}

	return resp.AppDebugChatResp{Content: res["text"].(string)}, nil
}
