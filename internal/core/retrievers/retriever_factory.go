package retrievers

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/vectorstores"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
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
	VectorStore  vectorstores.VectorStore
	Embedder     *embedding.EmbeddingService
	JiebaService *JiebaService
}

// NewRetrieverFactory 创建一个新的检索器工厂
func NewRetrieverFactory(db *gorm.DB, vectorStore vectorstores.VectorStore, embedder *embedding.EmbeddingService, jiebaService *JiebaService) *RetrieverFactory {
	return &RetrieverFactory{
		db:           db,
		VectorStore:  vectorStore,
		Embedder:     embedder,
		JiebaService: jiebaService,
	}
}

// CreateRetriever 根据类型创建检索器
func (f *RetrieverFactory) CreateRetriever(retrieverType RetrieverType, datasetIDs []uuid.UUID,
	options map[string]any) (interface{}, error) {
	switch retrieverType {
	case RetrieverTypeFullText:
		return f.createFullTextRetriever(f.db, datasetIDs, options), nil
	case RetrieverTypeSemantic:
		return f.createSemanticRetriever(datasetIDs, options), nil
	case RetrieverTypeHybrid:
		return f.createHybridRetriever(f.db, datasetIDs, options), nil
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}

// createFullTextRetriever 创建全文检索器
func (f *RetrieverFactory) createFullTextRetriever(db *gorm.DB, datasetIDs []uuid.UUID, options map[string]any) *FullTextRetriever {
	return NewFullTextRetriever(db, datasetIDs, f.JiebaService, options)
}

// createSemanticRetriever 创建语义检索器
func (f *RetrieverFactory) createSemanticRetriever(datasetIDs []uuid.UUID, options map[string]any) *SemanticRetriever {
	return NewSemanticRetriever(f.VectorStore, f.Embedder, datasetIDs, options)
}

// createHybridRetriever 创建混合检索器
func (f *RetrieverFactory) createHybridRetriever(db *gorm.DB, datasetIDs []uuid.UUID, options map[string]any) *HybridRetriever {
	fullTextRetriever := f.createFullTextRetriever(db, datasetIDs, options)
	semanticRetriever := f.createSemanticRetriever(datasetIDs, options)
	return NewHybridRetriever(fullTextRetriever, semanticRetriever, datasetIDs, options)
}

// ValidateDatasetAccess 验证用户对数据集的访问权限
func (f *RetrieverFactory) ValidateDatasetAccess(userID uuid.UUID, datasetIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(datasetIDs) == 0 {
		return nil, fmt.Errorf("dataset IDs cannot be empty")
	}

	var validDatasetIDs []uuid.UUID

	// 查询用户有权限访问的数据集
	query := `SELECT id FROM datasets WHERE id IN ? AND account_id = ?`
	err := f.db.Raw(query, datasetIDs, userID).Scan(&validDatasetIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to validate dataset access: %w", err)
	}

	return validDatasetIDs, nil
}

// RecordDatasetQuery 记录数据集查询历史
func (f *RetrieverFactory) RecordDatasetQuery(userID uuid.UUID, datasetID uuid.UUID, query string, source string) error {
	// 插入查询记录
	insertQuery := `
		INSERT INTO dataset_queries (id, dataset_id, content, source, created_by, created_at) 
		VALUES (gen_random_uuid(), ?, ?, ?, ?, NOW())
	`

	err := f.db.Exec(insertQuery, datasetID, query, source, userID).Error
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

	// 批量更新片段命中次数
	updateQuery := `UPDATE segments SET hit_count = hit_count + 1 WHERE id IN ?`
	err := f.db.Exec(updateQuery, segmentIDs).Error
	if err != nil {
		return fmt.Errorf("failed to update segment hit count: %w", err)
	}

	return nil
}
