package ioc

import (
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/builtin_apps"
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/file_extractor"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	apitools "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/categories"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/crazyfrankie/voidx/pkg/langchainx/embeddings"
)

func InitAgentManager(cmd redis.Cmdable) *agent.AgentQueueManager {
	return agent.NewAgentQueueManager(cmd)
}

func InitBuiltinAppManager() *builtin_apps.BuiltinAppManager {
	return builtin_apps.NewBuiltinAppManager()
}

func InitBuiltinToolsCategories() *categories.BuiltinCategoryManager {
	manager, err := categories.NewBuiltinCategoryManager()
	if err != nil {
		panic(err)
	}
	return manager
}

func InitEmbeddingService(cmd redis.Cmdable, embedder *embeddings.OpenAI) *embedding.EmbeddingService {
	return embedding.NewEmbeddingService(cmd, embedder)
}

func InitFileExtractor(uploadSvc *upload.Service) *file_extractor.FileExtractor {
	return file_extractor.NewFileExtractor(uploadSvc)
}

func InitJiebaService() *retrievers.JiebaService {
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

func InitLLMCore() *llm.LanguageModelManager {
	llmCore, err := llm.NewLanguageModelManager()
	if err != nil {
		return nil
	}

	return llmCore
}

func InitTokenBufMem(db *gorm.DB) *memory.TokenBufferMemory {
	return memory.NewTokenBufferMemory(db)
}

func InitApiToolsManager() *apitools.ApiProviderManager {
	return apitools.NewApiProviderManager()
}

func InitBuiltinToolsManager() *builtin.BuiltinProviderManager {
	manager, err := builtin.NewBuiltinProviderManager()
	if err != nil {
		panic(err)
	}

	return manager
}
