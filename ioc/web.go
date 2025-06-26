package ioc

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/controller"
)

func InitWeb(chat *controller.ChatHandler) *gin.Engine {
	srv := gin.Default()

	apiGroup := srv.Group("api")

	chat.RegisterRoute(apiGroup)

	return srv
}
