//go:build wireinject
// +build wireinject

package analysis

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/analysis/handler"
	"github.com/crazyfrankie/voidx/internal/analysis/repository"
	"github.com/crazyfrankie/voidx/internal/analysis/repository/cache"
	"github.com/crazyfrankie/voidx/internal/analysis/repository/dao"
	"github.com/crazyfrankie/voidx/internal/analysis/service"
	"github.com/crazyfrankie/voidx/internal/app"
)

type Handler = handler.AnalysisHandler

type AnalysisModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewAnalysisDao,
	cache.NewAnalysisCache,
	repository.NewAnalysisRepo,
	service.NewAnalysisService,
	handler.NewAnalysisHandler,
)

func InitAnalysisModule(db *gorm.DB, rdb redis.Cmdable, appModule *app.AppModule) *AnalysisModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(AnalysisModule), "*"),
		wire.FieldsOf(new(*app.AppModule), "Service"),
	)
	return new(AnalysisModule)
}
