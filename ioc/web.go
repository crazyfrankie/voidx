package ioc

import (
	"github.com/crazyfrankie/voidx/internal/middlewares"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/controller"
)

func InitWeb(middlewares []gin.HandlerFunc, chat *controller.ChatHandler) *gin.Engine {
	srv := gin.Default()
	srv.Use(middlewares...)

	apiGroup := srv.Group("api")

	chat.RegisterRoute(apiGroup)

	return srv
}

func InitMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middlewares.CORS(),
	}
}
