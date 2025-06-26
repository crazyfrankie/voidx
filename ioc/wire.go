//go:build wireinject

package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/crazyfrankie/voidx/internal/controller"
)

func InitEngine() *gin.Engine {
	wire.Build(
		controller.NewChatHandler,

		InitWeb,
	)

	return new(gin.Engine)
}
