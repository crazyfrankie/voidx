//go:build wireinject

package app

import (
	"github.com/crazyfrankie/voidx/internal/app/handler"
	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app/service"
	llmcore "github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.AppHandler
type Service = service.AppService

type AppModule struct {
	Handler *Handler
	Service *Service
}

func InitModel(llmManager *llmcore.LanguageModelManager) entity.BaseLanguageModel {
	model, err := llmManager.CreateModel("tongyi", "qwen-max", map[string]any{
		"base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
	})
	if err != nil {
		panic(err)
	}

	return model
}

func InitAppModule(db *gorm.DB, vecStore *vecstore.VecStoreService,
	llmCore *llmcore.LanguageModelManager) *AppModule {
	wire.Build(
		InitModel,
		dao.NewAppDao,
		repository.NewAppRepo,
		service.NewAppService,
		handler.NewAppHandler,

		wire.Struct(new(AppModule), "*"),
	)
	return new(AppModule)
}
