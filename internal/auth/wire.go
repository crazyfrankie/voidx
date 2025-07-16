//go:build wireinject

package auth

import (
	"github.com/crazyfrankie/voidx/internal/auth/handler"
	"github.com/crazyfrankie/voidx/internal/auth/repository"
	"github.com/crazyfrankie/voidx/internal/auth/repository/dao"
	"github.com/crazyfrankie/voidx/internal/auth/service"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler = handler.AuthHandler

type AuthModule struct {
	Handler *Handler
}

func InitAuthModule(db *gorm.DB, cmd redis.Cmdable, token *jwt.TokenService) *AuthModule {
	wire.Build(
		dao.NewAuthDao,
		repository.NewAuthRepo,
		service.NewAuthService,
		handler.NewAuthHandler,

		wire.Struct(new(AuthModule), "*"),
	)
	return new(AuthModule)
}
