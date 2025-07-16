//go:build wireinject

package app

import (
	"github.com/crazyfrankie/voidx/internal/app/handler"
	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/google/wire"
	"github.com/tmc/langchaingo/llms/openai"
	"gorm.io/gorm"
)

type Handler = handler.AppHandler

type AppModule struct {
	Handler *Handler
}

func InitAppModule(db *gorm.DB, llm *openai.LLM) *AppModule {
	wire.Build(
		dao.NewAppDao,
		repository.NewAppRepo,
		service.NewAppService,
		handler.NewAppHandler,

		wire.Struct(new(AppModule), "*"),
	)
	return new(AppModule)
}
