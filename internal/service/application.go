package service

import (
	"context"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"

	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/repository/dao"
	"github.com/crazyfrankie/voidx/pkg/langchainx"
)

type AppService struct {
	llm *openai.LLM
	dao *dao.AppDao
}

func NewAppService(llm *openai.LLM, dao *dao.AppDao) *AppService {
	return &AppService{llm: llm, dao: dao}
}

func (s *AppService) GetAppConfig(ctx context.Context, appid string) (resp.AppConfig, error) {
	// TODO
	panic("implement me")
}

func (s *AppService) UpdateAppDraftConfig(ctx context.Context, request req.UpdateAppConfReq) error {
	// TODO
	panic("implement me")
}

func (s *AppService) GetAppDebugLongMemory(ctx context.Context, appid string) (resp.GetAppLTMResp, error) {
	// TODO
	panic("implement me")
}

func (s *AppService) UpdateAppDebugLongMemory(ctx context.Context, request req.UpdateAppDebugLTMReq) error {
	// TODO
	panic("implement me")
}

func (s *AppService) AppDebugChat(ctx context.Context, query string) (resp.AppDebugChatResp, error) {
	chatHis, err := langchainx.NewFileChatMessageHistory(langchainx.WithFilePath("storage/memory/chat_history.json"))
	if err != nil {
		return resp.AppDebugChatResp{}, err
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
		return resp.AppDebugChatResp{}, err
	}

	return resp.AppDebugChatResp{Content: res["text"].(string)}, nil
}

func (s *AppService) GetAppDebugHistoryList(ctx context.Context, appid string) (resp.AppDebugChatResp, error) {
	// TODO
	panic("implement me")
}

func (s *AppService) DeleteDebugMessage(ctx context.Context, appid string, messageId string) error {
	// TODO
	panic("implement me")
}
