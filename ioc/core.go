package ioc

import (
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/builtin_apps"
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/file_extractor"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	apitools "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/categories"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/upload"
)

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

func InitEmbeddingService() *embedding.EmbeddingService {
	eb, err := embedding.NewEmbeddingService("https://dashscope.aliyuncs.com/compatible-mode/v1", "text-embedding-v4")
	if err != nil {
		panic(err)
	}
	return eb
}

func InitFileExtractor(uploadSvc *upload.Service) *file_extractor.FileExtractor {
	fileExtra, err := file_extractor.NewFileExtractor(uploadSvc)
	if err != nil {
		panic(err)
	}
	return fileExtra
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
	return memory.NewTokenBufferMemory(db, 2000)
}

func InitApiToolsManager() *apitools.APIProviderManager {
	return apitools.NewAPIProviderManager()
}

func InitBuiltinToolsManager() *builtin.BuiltinProviderManager {
	builtinMan, err := builtin.NewBuiltinProviderManager()
	if err != nil {
		panic(err)
	}

	return builtinMan
}
