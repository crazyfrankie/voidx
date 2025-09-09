//go:build wireinject

package app

import (
	"context"

	"github.com/crazyfrankie/voidx/internal/app/handler"
	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	llmcore "github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/llm/entities"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.AppHandler
type Service = service.AppService

type AppModule struct {
	Handler *Handler
	Service *Service
}

func InitModel(llmManager *llmcore.LanguageModelManager) entities.BaseLanguageModel {
	model, err := llmManager.CreateModel(context.Background(), "tongyi", "qwen-max", map[string]any{
		"base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
	})
	if err != nil {
		panic(err)
	}

	return model
}

func InitAppModule(db *gorm.DB, memory *memory.TokenBufferMemory,
	llmCore *llmcore.LanguageModelManager, appConfig *app_config.AppConfigModule, agentSvc *agent.AgentQueueManagerFactory,
	ossSvc *upload.UploadModule, retrieverSvc *retriever.RetrieverModule,
	apiProvider *providers.APIProviderManager, builtinProvider *builtin.BuiltinProviderManager,
	convers *conversation.ConversationModule) *AppModule {
	wire.Build(
		InitModel,
		dao.NewAppDao,
		repository.NewAppRepo,
		service.NewAppService,
		handler.NewAppHandler,

		wire.Struct(new(AppModule), "*"),
		wire.FieldsOf(new(*app_config.AppConfigModule), "Service"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
		wire.FieldsOf(new(*upload.UploadModule), "Service"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Service"),
	)
	return new(AppModule)
}
