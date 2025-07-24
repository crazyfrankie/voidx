//go:build wireinject
// +build wireinject

package upload

import (
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/upload/handler"
	"github.com/crazyfrankie/voidx/internal/upload/repository"
	"github.com/crazyfrankie/voidx/internal/upload/repository/dao"
	"github.com/crazyfrankie/voidx/internal/upload/service"
)

type Handler = handler.UploadFileHandler

type UploadModule struct {
	Handler *Handler
}

var ProviderSet = wire.NewSet(
	dao.NewUploadFileDao,
	repository.NewUploadFileRepo,
	service.NewUploadFileService,
	handler.NewUploadFileHandler,
)

func InitUploadModule(db *gorm.DB, minioCli *minio.Client) *UploadModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(UploadModule), "*"),
	)
	return new(UploadModule)
}
