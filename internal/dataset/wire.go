//go:build wireinject
// +build wireinject

package dataset

import (
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/segment"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/dataset/handler"
	"github.com/crazyfrankie/voidx/internal/dataset/repository"
	"github.com/crazyfrankie/voidx/internal/dataset/repository/dao"
	"github.com/crazyfrankie/voidx/internal/dataset/service"
)

type Handler = handler.DatasetHandler
type Service = service.DatasetService
type DataSetModule struct {
	Handler *Handler
	Service *Service
}

var DatasetSet = wire.NewSet(
	dao.NewDatasetDao,
	repository.NewDatasetRepo,
	service.NewDatasetService,
	handler.NewDatasetHandler,
)

func InitDatasetHandler(db *gorm.DB,
	retrieverModule *retriever.RetrieverModule, segmentModule *segment.SegmentModule) *DataSetModule {
	wire.Build(
		DatasetSet,

		wire.Struct(new(DataSetModule), "*"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Service"),
		wire.FieldsOf(new(*segment.SegmentModule), "Service"),
	)
	return new(DataSetModule)
}
