//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitCache, InitDB, InitLLM, InitJWT)

func InitEngine() *gin.Engine {
	wire.Build(
		BaseSet,

		auth.InitAuthModule,
		app.InitAppModule,
		account.InitAccountModule,

		InitMiddlewares,
		InitWeb,

		wire.FieldsOf(new(*app.AppModule), "Handler"),
		wire.FieldsOf(new(*auth.AuthModule), "Handler"),
		wire.FieldsOf(new(*account.AccountModule), "Handler"),
	)

	return new(gin.Engine)
}
