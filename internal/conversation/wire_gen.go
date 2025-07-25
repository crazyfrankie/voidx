// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package conversation

import (
	"github.com/crazyfrankie/voidx/internal/conversation/handler"
	"github.com/crazyfrankie/voidx/internal/conversation/repository"
	"github.com/crazyfrankie/voidx/internal/conversation/repository/dao"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// Injectors from wire.go:

func InitConversationModule(db *gorm.DB) *ConversationModule {
	conversationDao := dao.NewConversationDao(db)
	conversationRepo := repository.NewConversationRepo(conversationDao)
	conversationService := service.NewConversationService(conversationRepo)
	conversationHandler := handler.NewConversationHandler(conversationService)
	conversationModule := &ConversationModule{
		Handler: conversationHandler,
		Service: conversationService,
	}
	return conversationModule
}

// wire.go:

type Handler = handler.ConversationHandler

type Service = service.ConversationService

type ConversationModule struct {
	Handler *Handler
	Service *Service
}

var ConversationSet = wire.NewSet(dao.NewConversationDao, repository.NewConversationRepo, service.NewConversationService, handler.NewConversationHandler)
