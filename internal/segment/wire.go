//go:build wireinject
// +build wireinject

package segment

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/segment/handler"
	"github.com/crazyfrankie/voidx/internal/segment/repository"
	"github.com/crazyfrankie/voidx/internal/segment/repository/dao"
	"github.com/crazyfrankie/voidx/internal/segment/service"
	"github.com/crazyfrankie/voidx/internal/vecstore"
)

type Handler = handler.SegmentHandler
type Service = service.SegmentService

type SegmentModule struct {
	Handler *Handler
	Service *Service
}

var SegmentSet = wire.NewSet(
	dao.NewSegmentDao,
	repository.NewSegmentRepo,
	service.NewSegmentService,
	handler.NewSegmentHandler,
)

func InitSegmentModule(db *gorm.DB, embeddings *embedding.EmbeddingService,
	jiebaSvc *retrievers.JiebaService,
	vecSvc *vecstore.VecStoreService,
	keywordSvc *retriever.KeyWordService) *SegmentModule {
	wire.Build(
		SegmentSet,

		wire.Struct(new(SegmentModule), "*"),
	)
	return new(SegmentModule)
}
