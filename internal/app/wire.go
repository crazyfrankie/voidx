//go:build wireinject

package app

import (
	"github.com/crazyfrankie/voidx/internal/app/handler"
	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app/service"
	llmcore "github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.AppHandler

type AppModule struct {
	Handler *Handler
}

func InitAppModule(db *gorm.DB, vecStore *vecstore.VecStoreService,
	llmCore *llmcore.LanguageModelManager) *AppModule {
	wire.Build(
		dao.NewAppDao,
		repository.NewAppRepo,
		service.NewAppService,
		handler.NewAppHandler,

		wire.Struct(new(AppModule), "*"),
	)
	return new(AppModule)
}
