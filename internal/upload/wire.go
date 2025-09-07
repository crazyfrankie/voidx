//go:build wireinject
// +build wireinject

package upload

import (
	"github.com/crazyfrankie/voidx/infra/contract/storage"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/upload/handler"
	"github.com/crazyfrankie/voidx/internal/upload/repository"
	"github.com/crazyfrankie/voidx/internal/upload/repository/dao"
	"github.com/crazyfrankie/voidx/internal/upload/service"
)

type Handler = handler.UploadFileHandler
type Service = service.OssService
type UploadModule struct {
	Handler *Handler
	Service *Service
}

var ProviderSet = wire.NewSet(
	dao.NewUploadFileDao,
	repository.NewUploadFileRepo,
	service.NewOssService,
	handler.NewUploadFileHandler,
)

func InitUploadModule(db *gorm.DB, minioCli storage.Storage) *UploadModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(UploadModule), "*"),
	)
	return new(UploadModule)
}
