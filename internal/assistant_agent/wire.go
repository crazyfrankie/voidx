//go:build wireinject
// +build wireinject

package assistant_agent

import (
	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	llmcore "github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/handler"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository/dao"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/service"
)

type Handler = handler.AssistantAgentHandler

type AssistantModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewAssistantAgentDao,
	repository.NewAssistantAgentRepo,
	service.NewAssistantAgentService,
	handler.NewAssistantAgentHandler,
)

func InitModel(llmManager *llmcore.LanguageModelManager) entity.BaseLanguageModel {
	model, err := llmManager.CreateModel("tongyi", "qwen-max", map[string]any{
		"base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
	})
	if err != nil {
		panic(err)
	}

	return model
}

func InitProducer() *task.AppProducer {
	producer, err := task.NewAppProducer(conf.GetConf().Kafka.Brokers)
	if err != nil {
		panic(err)
	}

	return producer
}

func InitAssistantModule(db *gorm.DB, conversationSvc *conversation.ConversationModule,
	agentManager *agent.AgentQueueManager, llmManager *llmcore.LanguageModelManager,
	tokeBufMem *memory.TokenBufferMemory) *AssistantModule {
	wire.Build(
		InitProducer,
		ProviderSet,
		InitModel,

		wire.Struct(new(AssistantModule), "*"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
	)
	return new(AssistantModule)
}
