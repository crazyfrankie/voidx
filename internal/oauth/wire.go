//go:build wireinject

package oauth

import (
	"github.com/crazyfrankie/voidx/internal/oauth/handler"
	"github.com/crazyfrankie/voidx/internal/oauth/repository"
	"github.com/crazyfrankie/voidx/internal/oauth/repository/dao"
	"github.com/crazyfrankie/voidx/internal/oauth/service"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.OAuthHandler

type OAuthModule struct {
	Handler *Handler
}

func InitOAuthModule(db *gorm.DB, token *jwt.TokenService) *OAuthModule {
	wire.Build(
		dao.NewOAuthDao,
		repository.NewOAuthRepo,
		service.NewOAuthService,
		handler.NewOAuthHandler,

		wire.Struct(new(OAuthModule), "*"),
	)
	return new(OAuthModule)
}
