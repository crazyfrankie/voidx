//go:build wireinject

package openapi

import (
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/openapi/handler"
	"github.com/crazyfrankie/voidx/internal/openapi/repository"
	"github.com/crazyfrankie/voidx/internal/openapi/repository/dao"
	"github.com/crazyfrankie/voidx/internal/openapi/service"
)

type Handler = handler.OpenAPIHandler

type OpenAPIModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewOpenAPIDao,
	repository.NewOpenAPIRepo,
	service.NewOpenAPIService,
	handler.NewOpenAPIHandler,
)

func InitOpenAIModule(db *gorm.DB, conversationSvc *conversation.ConversationModule,
	retrieverModule *retriever.RetrieverModule, llmModule *llm.LLMModule, appConfig *app_config.AppConfigModule,
	appModule *app.AppModule, agent *agent.AgentQueueManager, token *memory.TokenBufferMemory) *OpenAPIModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(OpenAPIModule), "*"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Service"),
		wire.FieldsOf(new(*llm.LLMModule), "Service"),
		wire.FieldsOf(new(*app.AppModule), "Service"),
		wire.FieldsOf(new(*app_config.AppConfigModule), "Service"),
	)
	return new(OpenAPIModule)
}
