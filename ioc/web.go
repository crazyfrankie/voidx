package ioc

import (
	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/middlewares"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func InitWeb(middlewares []gin.HandlerFunc, app *app.Handler, auth *auth.Handler, account *account.Handler) *gin.Engine {
	srv := gin.Default()
	srv.Use(middlewares...)

	apiGroup := srv.Group("api")

	auth.RegisterRoute(apiGroup)
	app.RegisterRoute(apiGroup)
	account.RegisterRoute(apiGroup)

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
