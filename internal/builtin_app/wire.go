//go:build wireinject

package builtin_app

import (
	"github.com/crazyfrankie/voidx/internal/builtin_app/handler"
	"github.com/crazyfrankie/voidx/internal/builtin_app/repository"
	"github.com/crazyfrankie/voidx/internal/builtin_app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/builtin_app/service"
	"github.com/crazyfrankie/voidx/internal/core/builtin_apps"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.BuiltinAppHandler

type BuiltinModule struct {
	Handler *Handler
}

func InitBuiltinAppModule(db *gorm.DB, builtinManager *builtin_apps.BuiltinAppManager) *BuiltinModule {
	wire.Build(
		dao.NewBuiltinDao,
		repository.NewBuiltinRepository,
		service.NewBuiltinService,
		handler.NewBuiltinAppHandler,

		wire.Struct(new(BuiltinModule), "*"),
	)
	return new(BuiltinModule)
}
