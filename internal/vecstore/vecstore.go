package vecstore

import (
	"context"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"

	"github.com/crazyfrankie/voidx/pkg/langchainx/milvus"
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

func (s *VecStoreService) DeleteDocumentsByID(ctx context.Context, key string, value any) error {
	return s.vecStore.DeleteDocument(ctx, key, value)
}

func (s *VecStoreService) UpdateDocument(ctx context.Context, column ...entity.Column) error {
	return s.vecStore.UpdateDocument(ctx, column...)
}
