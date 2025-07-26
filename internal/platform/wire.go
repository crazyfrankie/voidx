//go:build wireinject
// +build wireinject

package platform

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/platform/handler"
	"github.com/crazyfrankie/voidx/internal/platform/repository"
	"github.com/crazyfrankie/voidx/internal/platform/repository/dao"
	"github.com/crazyfrankie/voidx/internal/platform/service"
)

type Handler = handler.PlatformHandler

type PlatformModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewPlatformDao,
	repository.NewPlatformRepo,
	service.NewPlatformService,
	handler.NewPlatformHandler,
)

func InitPlatformModule(db *gorm.DB) *PlatformModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(PlatformModule), "*"),
	)
	return new(PlatformModule)
}
