//go:build wireinject
// +build wireinject

package app_config

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/app_config/repository"
	"github.com/crazyfrankie/voidx/internal/app_config/repository/dao"
	"github.com/crazyfrankie/voidx/internal/app_config/service"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	apitools "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
)

type Service = service.AppConfigService
type AppConfigModule struct {
	Service *Service
}

var ProviderSet = wire.NewSet(
	dao.NewAppConfigDao,
	repository.NewAppConfigRepo,
	service.NewAppConfigService,
)

func InitAppConfigModule(db *gorm.DB, llmMgr *llm.LanguageModelManager,
	builtinProvider *builtin.BuiltinProviderManager,
	apiProvider *apitools.ApiProviderManager) *AppConfigModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(AppConfigModule), "*"),
	)
	return new(AppConfigModule)
}
