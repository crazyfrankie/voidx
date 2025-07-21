package vecstore

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/milvus"
)

type VecStoreService struct {
	vecStore *milvus.Store
}

func NewVecStoreService(vecStore *milvus.Store) *VecStoreService {
	return &VecStoreService{vecStore: vecStore}
}

func (s *VecStoreService) SimilaritySearch(ctx context.Context, query string, numDocs int, opts ...vectorstores.Option) ([]schema.Document, error) {
	return s.GetRetriever(numDocs, opts...).GetRelevantDocuments(ctx, query)
}

func (s *VecStoreService) AddDocument(ctx context.Context, documents []schema.Document) ([]string, error) {
	return s.vecStore.AddDocuments(ctx, documents)
}

func (s *VecStoreService) GetRetriever(numDocs int, opts ...vectorstores.Option) vectorstores.Retriever {
	return vectorstores.ToRetriever(s.vecStore, numDocs, opts...)
}
