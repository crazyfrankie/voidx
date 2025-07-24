//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var BaseSet = wire.NewSet(InitCache, InitDB, InitLLMCore, InitJWT, InitEmbedding, InitVectorStore, InitMinIO)

func InitEngine() *gin.Engine {
	wire.Build(
		BaseSet,

		vecstore.NewVecStoreService,

		auth.InitAuthModule,
		app.InitAppModule,
		account.InitAccountModule,
		llm.InitLLMModule,
		upload.InitUploadModule,

		InitMiddlewares,
		InitWeb,

		wire.FieldsOf(new(*app.AppModule), "Handler"),
		wire.FieldsOf(new(*auth.AuthModule), "Handler"),
		wire.FieldsOf(new(*account.AccountModule), "Handler"),
		wire.FieldsOf(new(*llm.LLMModule), "Handler"),
		wire.FieldsOf(new(*upload.UploadModule), "Handler"),
	)

	return new(gin.Engine)
}
