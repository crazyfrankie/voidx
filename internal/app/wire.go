//go:build wireinject

package app

import (
	"github.com/crazyfrankie/voidx/internal/app/handler"
	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app/service"
	llmcore "github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.AppHandler

type AppModule struct {
	Handler *Handler
}

func InitAppModule(db *gorm.DB, llmCore *llmcore.LanguageModelManager, llmModule *llm.LLMModule) *AppModule {
	wire.Build(
		dao.NewAppDao,
		repository.NewAppRepo,
		service.NewAppService,
		handler.NewAppHandler,

		wire.Struct(new(AppModule), "*"),
		wire.FieldsOf(new(*llm.LLMModule), "Service"),
	)
	return new(AppModule)
}
