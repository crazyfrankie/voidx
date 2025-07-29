//go:build wireinject

package index

import (
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/file_extractor"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/index/repository"
	"github.com/crazyfrankie/voidx/internal/index/repository/dao"
	"github.com/crazyfrankie/voidx/internal/index/service"
	"github.com/crazyfrankie/voidx/internal/process_rule"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service = service.IndexingService

type IndexModule struct {
	Service *Service
}

func InitIndexModule(db *gorm.DB, cmd redis.Cmdable, fileExtractor *file_extractor.FileExtractor,
	embeddingsSvc *embedding.EmbeddingService, jiebaService *retrievers.JiebaService, processRuleService *process_rule.ProcessRuleModule,
	keywordSvc *retriever.RetrieverModule, vectorDatabaseService *vecstore.VecStoreService) *IndexModule {
	wire.Build(
		dao.NewIndexingDao,
		repository.NewIndexingRepo,
		service.NewIndexingService,

		wire.Struct(new(IndexModule), "*"),
		wire.FieldsOf(new(*retriever.RetrieverModule), "Keyword"),
		wire.FieldsOf(new(*process_rule.ProcessRuleModule), "Service"),
	)
	return new(IndexModule)
}
