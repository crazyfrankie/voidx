//go:build wireinject

package wechat

import (
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/wechat/handler"
	"github.com/crazyfrankie/voidx/internal/wechat/repository"
	"github.com/crazyfrankie/voidx/internal/wechat/repository/dao"
	"github.com/crazyfrankie/voidx/internal/wechat/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.WechatHandler

type WechatModule struct {
	Handler *Handler
}

func InitWechatModule(db *gorm.DB, wec *wechat.Wechat, retrieval *retriever.RetrieverModule, appConfigSvc *app_config.AppConfigModule,
	conversationSvc *conversation.ConversationModule, llmSvc *llm.LLMModule, agentManager *agent.AgentQueueManager,
	tokenBufMem *memory.TokenBufferMemory) *WechatModule {
	wire.Build(
		dao.NewWechatDao,
		repository.NewWechatRepository,
		service.NewWechatService,
		handler.NewWechatHandler,

		wire.Struct(new(WechatModule), "*"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Service"),
		wire.FieldsOf(new(*app_config.AppConfigModule), "Service"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
		wire.FieldsOf(new(*llm.LLMModule), "Service"),
	)
	return new(WechatModule)
}
