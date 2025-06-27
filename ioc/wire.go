//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/voidx/internal/repository/dao"
	"github.com/crazyfrankie/voidx/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/crazyfrankie/voidx/internal/controller"
)

func InitEngine() *gin.Engine {
	wire.Build(
		InitDB,

		dao.NewChatDao,
		service.NewChatService,
		controller.NewChatHandler,

		InitMiddlewares,
		InitWeb,
	)

	return new(gin.Engine)
}
