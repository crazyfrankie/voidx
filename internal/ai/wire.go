//go:build wireinject
// +build wireinject

package ai

import (
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/ai/handler"
	"github.com/crazyfrankie/voidx/internal/ai/repository"
	"github.com/crazyfrankie/voidx/internal/ai/repository/dao"
	"github.com/crazyfrankie/voidx/internal/ai/service"
)

type Handler = handler.AIHandler

type AIModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewAIDao,
	repository.NewAIRepo,
	service.NewAIService,
	handler.NewAIHandler,
)

func InitAIModule(db *gorm.DB, conversationModule *conversation.ConversationModule) *AIModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(AIModule), "*"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Service"),
	)
	return new(AIModule)
}
