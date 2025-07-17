//go:build wireinject

package account

import (
	"github.com/crazyfrankie/voidx/internal/account/handler"
	"github.com/crazyfrankie/voidx/internal/account/repository"
	"github.com/crazyfrankie/voidx/internal/account/repository/dao"
	"github.com/crazyfrankie/voidx/internal/account/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.AccountHandler

type AccountModule struct {
	Handler *Handler
}

func InitAccountModule(db *gorm.DB) *AccountModule {
	wire.Build(
		dao.NewAccountDao,
		repository.NewAccountRepo,
		service.NewAccountService,
		handler.NewAccountHandler,

		wire.Struct(new(AccountModule), "*"),
	)
	return new(AccountModule)
}
