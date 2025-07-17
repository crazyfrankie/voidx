//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitCache, InitDB, InitLLMCore, InitJWT)

func InitEngine() *gin.Engine {
	wire.Build(
		BaseSet,

		auth.InitAuthModule,
		app.InitAppModule,
		account.InitAccountModule,
		llm.InitLLMModule,

		InitMiddlewares,
		InitWeb,

		wire.FieldsOf(new(*app.AppModule), "Handler"),
		wire.FieldsOf(new(*auth.AuthModule), "Handler"),
		wire.FieldsOf(new(*account.AccountModule), "Handler"),
		wire.FieldsOf(new(*llm.LLMModule), "Handler"),
	)

	return new(gin.Engine)
}
