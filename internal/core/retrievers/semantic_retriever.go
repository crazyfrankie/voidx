package retrievers

import (
	"context"
	"fmt"

	"github.com/crazyfrankie/voidx/infra/contract/document/vecstore"
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// SemanticRetriever implements semantic search using vector embeddings with eino
type SemanticRetriever struct {
	vecstore   vecstore.SearchStore
	embedder   *embedding.EmbeddingService
	datasetIDs []uuid.UUID
}

// NewSemanticRetriever creates a new semantic retriever
func NewSemanticRetriever(vecstore vecstore.SearchStore, embedder *embedding.EmbeddingService, datasetIDs []uuid.UUID) *SemanticRetriever {
	return &SemanticRetriever{
		vecstore:   vecstore,
		embedder:   embedder,
		datasetIDs: datasetIDs,
	}
}

// Retrieve implements the eino retriever interface
func (r *SemanticRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	filters := make(map[string]any)
	// 添加数据集ID过滤
	if len(r.datasetIDs) > 0 {
		datasetIDStrs := make([]string, 0, len(r.datasetIDs))
		for _, id := range r.datasetIDs {
			datasetIDStrs = append(datasetIDStrs, id.String())
		}
		filters["dataset_id"] = map[string]any{
			"contains_any": datasetIDStrs,
		}
	}

	// 添加其他过滤条件
	filters["document_enabled"] = map[string]any{
		"equal": true,
	}
	filters["segment_enabled"] = map[string]any{
		"equal": true,
	}

	docs, err := r.vecstore.Retrieve(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("retrieve failed, %w", err)
	}

	if len(docs) == 0 {
		return []*schema.Document{}, nil
	}

	return docs, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func (r *SemanticRetriever) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}
