//go:build wireinject

package builtin_tools

import (
	"github.com/crazyfrankie/voidx/internal/builtin_tools/handler"
	"github.com/crazyfrankie/voidx/internal/builtin_tools/service"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/categories"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/google/wire"
)

type Handler = handler.BuiltinToolsHandler

type BuiltinToolsModule struct {
	Handler *Handler
}

func InitBuiltinToolsModule(toolManager *categories.BuiltinCategoryManager,
	providerManager *providers.BuiltinProviderManager) *BuiltinToolsModule {
	wire.Build(
		service.NewBuiltinToolsService,
		handler.NewBuiltinToolsHandler,

		wire.Struct(new(BuiltinToolsModule), "*"),
	)
	return new(BuiltinToolsModule)
}
