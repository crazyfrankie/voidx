package service

import (
	"context"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"

	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/crazyfrankie/voidx/pkg/langchainx"
)

type AppService struct {
	repo     *repository.AppRepo
	llm      entity.BaseLanguageModel
	vecStore *vecstore.VecStoreService
}

func NewAppService(repo *repository.AppRepo, vecStore *vecstore.VecStoreService, llmManager *llm.LanguageModelManager) *AppService {
	model, err := llmManager.CreateModel("tongyi", "qwen-max", map[string]any{
		"base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
	})
	if err != nil {
		panic(err)
	}

	return &AppService{
		repo:     repo,
		llm:      model,
		vecStore: vecStore,
	}
}

func (s *AppService) DebugChat(ctx context.Context, appID uuid.UUID, chatReq req.DebugChatReq) (resp.AppDebugChatResp, error) {
	chatHis, err := langchainx.NewFileChatMessageHistory(langchainx.WithFilePath("storage/memory/chat_history.json"))
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
