//go:build wireinject
// +build wireinject

package apitool

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/apitool/handler"
	"github.com/crazyfrankie/voidx/internal/apitool/repository"
	"github.com/crazyfrankie/voidx/internal/apitool/repository/dao"
	"github.com/crazyfrankie/voidx/internal/apitool/service"
)

type Handler = handler.ApiToolHandler

type ApiToolModule struct {
	Handler *Handler
}

var ApiToolSet = wire.NewSet(
	dao.NewApiToolDao,
	repository.NewApiToolRepo,
	service.NewApiToolService,
	handler.NewApiToolHandler,
)

func InitApiToolHandler(db *gorm.DB) *ApiToolModule {
	wire.Build(
		ApiToolSet,

		wire.Struct(new(ApiToolModule), "*"),
	)
	return new(ApiToolModule)
}
