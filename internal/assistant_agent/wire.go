//go:build wireinject
// +build wireinject

package assistant_agent

import (
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/internal/conversation"
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

func InitProducer() *task.AppProducer {
	producer, err := task.NewAppProducer([]string{})
	if err != nil {
		panic(err)
	}

	return producer
}

func InitAssistantModule(db *gorm.DB, conversationSvc *conversation.ConversationModule) *AssistantModule {
	wire.Build(
		InitProducer,
		ProviderSet,

		wire.Struct(new(AssistantModule), "*"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
	)
	return new(AssistantModule)
}
