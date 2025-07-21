package ioc

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/crazyfrankie/voidx/pkg/middlewares"
)

func InitWeb(mws []gin.HandlerFunc, app *app.Handler, auth *auth.Handler,
	account *account.Handler, llm *llm.Handler) *gin.Engine {
	srv := gin.Default()
	srv.Use(mws...)

	apiGroup := srv.Group("api")

	auth.RegisterRoute(apiGroup)
	app.RegisterRoute(apiGroup)
	account.RegisterRoute(apiGroup)
	llm.RegisterRoute(apiGroup)

	return srv
}

func InitMiddlewares(jwt *jwt.TokenService) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middlewares.CORS(),

		middlewares.NewAuthnHandler(jwt).
			IgnorePath("/api/auth/login").
			Auth(),
	}
}
