package retrievers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
)

// SemanticRetriever 相似性检索器/向量检索器
type SemanticRetriever struct {
	VectorStore    vectorstores.VectorStore
	Embedder       *embedding.EmbeddingService
	SearchOptions  map[string]any
	DatasetIDs     []uuid.UUID
	ScoreThreshold float32
}

// NewSemanticRetriever 创建一个新的语义检索器
func NewSemanticRetriever(vectorStore vectorstores.VectorStore, embedder *embedding.EmbeddingService,
	datasetIDs []uuid.UUID, options map[string]any) *SemanticRetriever {
	// 默认相似度阈值
	scoreThreshold := float32(0.0)
	if threshold, ok := options["score_threshold"].(float32); ok {
		scoreThreshold = threshold
	}

	return &SemanticRetriever{
		VectorStore:    vectorStore,
		Embedder:       embedder,
		DatasetIDs:     datasetIDs,
		SearchOptions:  options,
		ScoreThreshold: scoreThreshold,
	}
}

// GetRelevantDocuments 根据传递的query执行相似性检索
func (r *SemanticRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	// 1.提取最大搜索条件k，默认值为4
	k := 4
	if kVal, ok := r.SearchOptions["k"]; ok {
		if kInt, ok := kVal.(int); ok {
			k = kInt
		}
	}

	// 2.构建过滤条件 - 根据Python实现使用Weaviate的过滤格式
	filters := make(map[string]any)

	// 添加数据集ID过滤
	if len(r.DatasetIDs) > 0 {
		datasetIDStrs := make([]string, 0, len(r.DatasetIDs))
		for _, id := range r.DatasetIDs {
			datasetIDStrs = append(datasetIDStrs, id.String())
		}
		// 使用contains_any过滤器，类似Python中的Filter.by_property("dataset_id").contains_any()
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

	// 3.执行相似性检索并获取得分信息
	searchOptions := []vectorstores.Option{
		vectorstores.WithFilters(filters),
	}

	// 添加其他搜索选项
	for key, value := range r.SearchOptions {
		if key != "k" && key != "score_threshold" {
			searchOptions = append(searchOptions, vectorstores.WithFilters(map[string]any{key: value}))
		}
	}

	// 使用SimilaritySearch方法
	docs, err := r.VectorStore.SimilaritySearch(ctx, query, k, searchOptions...)
	if err != nil {
		return nil, fmt.Errorf("similarity search failed: %w", err)
	}

	if len(docs) == 0 {
		return []schema.Document{}, nil
	}

	// 4.过滤低于阈值的结果
	if r.ScoreThreshold > 0 {
		filteredDocs := make([]schema.Document, 0, len(docs))
		for _, doc := range docs {
			if score, ok := doc.Metadata["score"].(float32); ok && score >= r.ScoreThreshold {
				filteredDocs = append(filteredDocs, doc)
			} else if score, ok := doc.Metadata["score"].(float64); ok && float32(score) >= r.ScoreThreshold {
				filteredDocs = append(filteredDocs, doc)
			} else {
				// 如果没有分数信息，保留文档
				filteredDocs = append(filteredDocs, doc)
			}
		}
		return filteredDocs, nil
	}

	return docs, nil
}
