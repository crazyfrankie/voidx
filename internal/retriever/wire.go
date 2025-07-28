//go:build wireinject

package retriever

import (
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/retriever/repository"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/cache"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/dao"
	"github.com/crazyfrankie/voidx/internal/retriever/service"
	"github.com/crazyfrankie/voidx/pkg/langchainx/milvus"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type KeyWordService = service.KeywordService
type Service = service.RetrievalService

type RetrieverModule struct {
	KeyWord *KeyWordService
	Service *Service
}

// InitRetrieverModule 初始化检索模块
func InitRetrieverModule(db *gorm.DB, cmd redis.Cmdable, vectorStore *milvus.Store,
	embedding *embedding.EmbeddingService, jiebaService *retrievers.JiebaService) *RetrieverModule {
	wire.Build(
		cache.NewKeyWordCache,
		dao.NewKeywordDao,
		// 初始化关键词表存储库
		repository.NewKeywordRepository,
		// 初始化关键词表服务
		service.NewKeywordService,
		// 初始化检索器工厂
		initRetrieverFactory,
		// 初始化检索服务
		service.NewRetrievalService,

		wire.Struct(new(RetrieverModule), "*"),
	)
	return new(RetrieverModule)
}

// initRetrieverFactory 初始化检索器工厂
func initRetrieverFactory(db *gorm.DB, vectorStore *milvus.Store, embedding *embedding.EmbeddingService, jiebaService *retrievers.JiebaService) *retrievers.RetrieverFactory {
	return retrievers.NewRetrieverFactory(db, vectorStore, embedding, jiebaService)
}
