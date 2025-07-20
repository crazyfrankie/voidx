package vecstore

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/milvus"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
)

type VecStoreService struct {
	vecStore     *milvus.Store
	embedService *embedding.EmbeddingService
}

func NewVecStoreService(vecStore *milvus.Store, embedService *embedding.EmbeddingService) *VecStoreService {
	return &VecStoreService{vecStore: vecStore, embedService: embedService}
}

func (s *VecStoreService) SimilaritySearch(ctx context.Context, query string, numDocs int, opts ...vectorstores.Option) ([]schema.Document, error) {
	return s.getRetriever(numDocs, opts...).GetRelevantDocuments(ctx, query)
}

func (s *VecStoreService) AddDocument(ctx context.Context, documents []schema.Document) ([]string, error) {
	return s.vecStore.AddDocuments(ctx, documents)
}

func (s *VecStoreService) getRetriever(numDocs int, opts ...vectorstores.Option) vectorstores.Retriever {
	opts = append(opts, vectorstores.WithEmbedder(s.embedService.Embedder))
	return vectorstores.ToRetriever(s.vecStore, numDocs, opts...)
}
