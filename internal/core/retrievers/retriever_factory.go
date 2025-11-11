package retrievers

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/infra/contract/document/vecstore"
)

// RetrieverType 检索器类型
type RetrieverType string

const (
	// RetrieverTypeFullText 全文检索器
	RetrieverTypeFullText RetrieverType = "full_text"
	// RetrieverTypeSemantic 语义检索器
	RetrieverTypeSemantic RetrieverType = "semantic"
	// RetrieverTypeHybrid 混合检索器
	RetrieverTypeHybrid RetrieverType = "hybrid"
)

// RetrieverFactory 检索器工厂，负责创建不同类型的检索器
type RetrieverFactory struct {
	db           *gorm.DB
	vectorStore  vecstore.SearchStore
	embedder     *embedding.EmbeddingService
	jiebaService *JiebaService
}

// NewRetrieverFactory 创建一个新的检索器工厂
func NewRetrieverFactory(db *gorm.DB, vectorStore vecstore.SearchStore, embedder *embedding.EmbeddingService, jiebaService *JiebaService) *RetrieverFactory {
	return &RetrieverFactory{
		db:           db,
		vectorStore:  vectorStore,
		embedder:     embedder,
		jiebaService: jiebaService,
	}
}

// CreateRetriever 根据类型创建检索器，支持多个数据集ID
func (f *RetrieverFactory) CreateRetriever(ctx context.Context, retrieverType RetrieverType,
	datasetIDs []uuid.UUID, options map[string]any) (retriever.Retriever, error) {
	// 验证参数
	if len(datasetIDs) == 0 {
		return nil, fmt.Errorf("datasetIDs cannot be empty")
	}

	switch retrieverType {
	case RetrieverTypeFullText:
		return f.createFullTextRetriever(ctx, datasetIDs)
	case RetrieverTypeSemantic:
		return f.createSemanticRetriever(ctx, datasetIDs, options)
	case RetrieverTypeHybrid:
		return f.createHybridRetriever(ctx, datasetIDs, options)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}

// createFullTextRetriever 创建全文检索器，支持多个数据集
func (f *RetrieverFactory) createFullTextRetriever(ctx context.Context, datasetIDs []uuid.UUID) (retriever.Retriever, error) {
	return NewFullTextRetriever(f.db, datasetIDs, f.jiebaService), nil
}

// createSemanticRetriever 创建语义检索器，支持多个数据集
func (f *RetrieverFactory) createSemanticRetriever(ctx context.Context, datasetIDs []uuid.UUID, options map[string]any) (retriever.Retriever, error) {
	return NewSemanticRetriever(f.vectorStore, f.embedder, datasetIDs), nil
}

// createHybridRetriever 创建混合检索器，支持多个数据集
func (f *RetrieverFactory) createHybridRetriever(ctx context.Context, datasetIDs []uuid.UUID, options map[string]any) (retriever.Retriever, error) {
	fullTextRetriever, err := f.createFullTextRetriever(ctx, datasetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create full text retriever: %w", err)
	}

	semanticRetriever, err := f.createSemanticRetriever(ctx, datasetIDs, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create semantic retriever: %w", err)
	}

	return NewHybridRetriever(fullTextRetriever, semanticRetriever, datasetIDs, options), nil
}

// ValidateDatasetAccess 验证用户对数据集的访问权限
func (f *RetrieverFactory) ValidateDatasetAccess(userID uuid.UUID, datasetIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(datasetIDs) == 0 {
		return nil, fmt.Errorf("dataset IDs cannot be empty")
	}

	var validDatasetIDs []uuid.UUID

	// 查询用户有权限访问的数据集
	err := f.db.Model(&entity.Dataset{}).
		Select("id").
		Where("id IN ? AND account_id = ?", datasetIDs, userID).
		Scan(&validDatasetIDs).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to validate dataset access: %w", err)
	}

	return validDatasetIDs, nil
}

// RecordDatasetQuery 记录数据集查询历史
func (f *RetrieverFactory) RecordDatasetQuery(userID uuid.UUID, datasetID uuid.UUID, query string, source string) error {
	datasetQuery := &entity.DatasetQuery{
		DatasetID: datasetID,
		Query:     query,
		Source:    source,
		CreatedBy: userID,
	}

	err := f.db.Create(datasetQuery).Error
	if err != nil {
		return fmt.Errorf("failed to record dataset query: %w", err)
	}

	return nil
}

// UpdateSegmentHitCount 批量更新片段命中次数
func (f *RetrieverFactory) UpdateSegmentHitCount(segmentIDs []uuid.UUID) error {
	if len(segmentIDs) == 0 {
		return nil
	}

	err := f.db.Model(&entity.Segment{}).
		Where("id IN ?", segmentIDs).
		Update("hit_count", gorm.Expr("hit_count + 1")).
		Error

	if err != nil {
		return fmt.Errorf("failed to update segment hit count: %w", err)
	}

	return nil
}
