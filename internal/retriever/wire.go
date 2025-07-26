//go:build wireinject

package retriever

import (
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/retriever/repository"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/cache"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/dao"
	"github.com/crazyfrankie/voidx/internal/retriever/service"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/vectorstores/milvus"
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
	embedding *embedding.EmbeddingService) *RetrieverModule {
	wire.Build(
		// 初始化Jieba服务
		initJiebaService,
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

// initJiebaService 初始化Jieba服务
func initJiebaService() *retrievers.JiebaService {
	// 加载停用词
	stopwords := []string{
		"的", "了", "和", "是", "在", "我", "有", "不", "这", "也",
		"就", "都", "而", "要", "把", "但", "可以", "你", "会", "对",
		"能", "他", "说", "着", "那", "如果", "只", "因为", "所以", "还",
		"a", "an", "the", "and", "or", "but", "if", "of", "to", "in",
		"for", "with", "on", "at", "by", "from", "up", "about", "into", "over",
		"after", "beneath", "under", "above",
	}

	stopwordSet := retrievers.LoadStopwords(stopwords)
	return retrievers.NewJiebaService(stopwordSet)
}

// initRetrieverFactory 初始化检索器工厂
func initRetrieverFactory(db *gorm.DB, vectorStore *milvus.Store, embedding *embedding.EmbeddingService, jiebaService *retrievers.JiebaService) *retrievers.RetrieverFactory {
	return retrievers.NewRetrieverFactory(db, vectorStore, embedding, jiebaService)
}
