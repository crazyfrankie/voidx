//go:build wireinject
// +build wireinject

package api_key

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/api_key/handler"
	"github.com/crazyfrankie/voidx/internal/api_key/repository"
	"github.com/crazyfrankie/voidx/internal/api_key/repository/dao"
	"github.com/crazyfrankie/voidx/internal/api_key/service"
)

type Handler = handler.ApiKeyHandler

type ApiKeyModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewApiKeyDao,
	repository.NewApiKeyRepo,
	service.NewApiKeyService,
	handler.NewApiKeyHandler,
)

func InitApiKeyModule(db *gorm.DB) *ApiKeyModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(ApiKeyModule), "*"),
	)
	return new(ApiKeyModule)
}
