//go:build wireinject
// +build wireinject

package webapp

import (
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/webapp/handler"
	"github.com/crazyfrankie/voidx/internal/webapp/repository"
	"github.com/crazyfrankie/voidx/internal/webapp/repository/dao"
	"github.com/crazyfrankie/voidx/internal/webapp/service"
)

type Handler = handler.WebAppHandler

type WebAppModule struct {
	Handler *Handler
}

var WebAppSet = wire.NewSet(
	dao.NewWebAppDao,
	repository.NewWebAppRepo,
	service.NewWebAppService,
	handler.NewWebAppHandler,
)

func InitWebAppModule(db *gorm.DB, tokenBufMem *memory.TokenBufferMemory,
	appConfigModule *app_config.AppConfigModule,
	conversationModule *conversation.ConversationModule,
	llmModule *llm.LLMModule,
	agentManager *agent.AgentQueueManager,
	retrievalModule *retriever.RetrieverModule,
) *WebAppModule {
	wire.Build(
		WebAppSet,

		wire.Struct(new(WebAppModule), "*"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
		wire.FieldsOf(new(*app_config.AppConfigModule), "Service"),
		wire.FieldsOf(new(*llm.LLMModule), "Service"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Service"),
	)
	return new(WebAppModule)
}
