//go:build wireinject
// +build wireinject

package document

import (
	"github.com/crazyfrankie/voidx/internal/document/task"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/document/handler"
	"github.com/crazyfrankie/voidx/internal/document/repository"
	"github.com/crazyfrankie/voidx/internal/document/repository/dao"
	"github.com/crazyfrankie/voidx/internal/document/service"
)

type Handler = handler.DocumentHandler

type DocumentModule struct {
	Handler *Handler
}

var DocumentSet = wire.NewSet(
	dao.NewDocumentDao,
	repository.NewDocumentRepo,
	service.NewDocumentService,
	handler.NewDocumentHandler,
)

func InitProducer() *task.DocumentProducer {
	producer, err := task.NewDocumentProducer([]string{})
	if err != nil {
		panic(err)
	}

	return producer
}

func InitDocumentModule(db *gorm.DB) *DocumentModule {
	wire.Build(
		InitProducer,
		DocumentSet,

		wire.Struct(new(DocumentModule), "*"),
	)
	return new(DocumentModule)
}
