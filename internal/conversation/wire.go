//go:build wireinject
// +build wireinject

package conversation

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/conversation/handler"
	"github.com/crazyfrankie/voidx/internal/conversation/repository"
	"github.com/crazyfrankie/voidx/internal/conversation/repository/dao"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
)

type Handler = handler.ConversationHandler
type Service = service.ConversationService

type ConversationModule struct {
	Handler *Handler
	Service *Service
}

var ConversationSet = wire.NewSet(
	dao.NewConversationDao,
	repository.NewConversationRepo,
	service.NewConversationService,
	handler.NewConversationHandler,
)

func InitConversationModule(db *gorm.DB) *ConversationModule {
	wire.Build(
		ConversationSet,

		wire.Struct(new(ConversationModule), "*"),
	)
	return new(ConversationModule)
}
