package retrievers

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/types/consts"
)

// RetrieverService 检索服务，提供统一的检索接口，支持多个数据集
type RetrieverService struct {
	factory *RetrieverFactory
}

// NewRetrieverService 创建一个新的检索服务
func NewRetrieverService(factory *RetrieverFactory) *RetrieverService {
	return &RetrieverService{
		factory: factory,
	}
}

// Search 执行检索，基于指定的检索策略，支持多个数据集
func (s *RetrieverService) Search(ctx context.Context, query string, datasetIDs []uuid.UUID,
	strategy consts.RetrievalStrategy,
	options map[string]any,
) ([]*schema.Document, error) {
	if len(datasetIDs) == 0 { // 修复：检查多个数据集
		return []*schema.Document{}, nil
	}

	if query == "" {
		return []*schema.Document{}, nil
	}

	// 根据检索策略创建相应的检索器
	var retrieverType RetrieverType
	switch strategy {
	case consts.RetrievalStrategySemantic:
		retrieverType = RetrieverTypeSemantic
	case consts.RetrievalStrategyFullText:
		retrieverType = RetrieverTypeFullText
	case consts.RetrievalStrategyHybrid:
		retrieverType = RetrieverTypeHybrid
	default:
		retrieverType = RetrieverTypeSemantic // 默认使用语义检索
	}

	ret, err := s.factory.CreateRetriever(ctx, retrieverType, datasetIDs, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create retriever: %w", err)
	}

	// 执行检索
	documents, err := ret.Retrieve(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	return documents, nil
}

// SearchWithMultipleStrategies 使用多种检索策略并合并结果，支持多个数据集
func (s *RetrieverService) SearchWithMultipleStrategies(
	ctx context.Context,
	query string,
	datasetIDs []uuid.UUID, // 修复：使用复数形式，支持多个数据集
	strategies []consts.RetrievalStrategy,
	options map[string]any,
) ([]*schema.Document, error) {
	if len(strategies) == 0 {
		// 默认使用语义检索
		return s.Search(ctx, query, datasetIDs, consts.RetrievalStrategySemantic, options)
	}

	if len(strategies) == 1 {
		return s.Search(ctx, query, datasetIDs, strategies[0], options)
	}

	// 多策略检索
	allResults := make([][]*schema.Document, 0, len(strategies))

	for _, strategy := range strategies {
		results, err := s.Search(ctx, query, datasetIDs, strategy, options)
		if err != nil {
			// 记录错误但继续其他策略
			continue
		}
		allResults = append(allResults, results)
	}

	if len(allResults) == 0 {
		return []*schema.Document{}, nil
	}

	// 合并结果
	return s.mergeResults(allResults, options), nil
}

// mergeResults 合并多个检索结果
func (s *RetrieverService) mergeResults(
	allResults [][]*schema.Document,
	options map[string]any,
) []*schema.Document {
	if len(allResults) == 0 {
		return []*schema.Document{}
	}

	if len(allResults) == 1 {
		return allResults[0]
	}

	// 使用 segment_id 去重
	docMap := make(map[string]*schema.Document)

	for _, results := range allResults {
		for _, doc := range results {
			segmentID := s.getSegmentID(doc)
			if segmentID != "" {
				// 如果已存在，保留分数更高的文档
				if existingDoc, exists := docMap[segmentID]; exists {
					if s.getScore(doc) > s.getScore(existingDoc) {
						docMap[segmentID] = doc
					}
				} else {
					docMap[segmentID] = doc
				}
			}
		}
	}

	// 转换为切片
	mergedDocs := make([]*schema.Document, 0, len(docMap))
	for _, doc := range docMap {
		mergedDocs = append(mergedDocs, doc)
	}

	// 应用k限制
	k := 4
	if kVal, ok := options["k"]; ok {
		if kInt, ok := kVal.(int); ok {
			k = kInt
		}
	}

	if len(mergedDocs) > k {
		mergedDocs = mergedDocs[:k]
	}

	return mergedDocs
}

// SearchInDatasets 在指定数据集中搜索 - 修复：明确支持多个数据集的方法名
func (s *RetrieverService) SearchInDatasets(
	ctx context.Context,
	query string,
	datasetIDs []uuid.UUID, // 修复：使用复数形式，支持多个数据集
	strategy consts.RetrievalStrategy,
	options map[string]any,
) ([]*schema.Document, error) {
	return s.Search(ctx, query, datasetIDs, strategy, options)
}

// CompareRetrievers 比较不同检索器类型的结果，支持多个数据集
func (s *RetrieverService) CompareRetrievers(
	ctx context.Context,
	query string,
	datasetIDs []uuid.UUID, // 修复：使用复数形式，支持多个数据集
	options map[string]any,
) (map[string][]*schema.Document, error) {
	results := make(map[string][]*schema.Document)

	strategies := []consts.RetrievalStrategy{
		consts.RetrievalStrategySemantic,
		consts.RetrievalStrategyFullText,
		consts.RetrievalStrategyHybrid,
	}

	for _, strategy := range strategies {
		docs, err := s.Search(ctx, query, datasetIDs, strategy, options)
		if err != nil {
			results[string(strategy)] = []*schema.Document{}
		} else {
			results[string(strategy)] = docs
		}
	}

	return results, nil
}

// getSegmentID 从文档元数据中获取segment_id
func (s *RetrieverService) getSegmentID(doc *schema.Document) string {
	if segmentID, ok := doc.MetaData["segment_id"].(string); ok {
		return segmentID
	}
	return ""
}

// getScore 从文档元数据中获取分数
func (s *RetrieverService) getScore(doc *schema.Document) float64 {
	if score, ok := doc.MetaData["score"].(float64); ok {
		return score
	}
	if score, ok := doc.MetaData["score"].(float32); ok {
		return float64(score)
	}
	return 0.0
}

// GetFactory 获取检索器工厂
func (s *RetrieverService) GetFactory() *RetrieverFactory {
	return s.factory
}

// CreateRetrieverTool 创建检索工具，用于 Agent 调用，支持多个数据集
func (s *RetrieverService) CreateRetrieverTool(
	ctx context.Context,
	datasetIDs []uuid.UUID, // 修复：使用复数形式，支持多个数据集
	strategy consts.RetrievalStrategy,
	options map[string]any,
) (retriever.Retriever, error) {
	var retrieverType RetrieverType
	switch strategy {
	case consts.RetrievalStrategySemantic:
		retrieverType = RetrieverTypeSemantic
	case consts.RetrievalStrategyFullText:
		retrieverType = RetrieverTypeFullText
	case consts.RetrievalStrategyHybrid:
		retrieverType = RetrieverTypeHybrid
	default:
		retrieverType = RetrieverTypeSemantic
	}

	return s.factory.CreateRetriever(ctx, retrieverType, datasetIDs, options)
}

// ValidateDatasetAccess 验证数据集访问权限
func (s *RetrieverService) ValidateDatasetAccess(userID uuid.UUID, datasetIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.factory.ValidateDatasetAccess(userID, datasetIDs)
}

// RecordSearchHistory 记录搜索历史，支持多个数据集
func (s *RetrieverService) RecordSearchHistory(
	userID uuid.UUID,
	datasetIDs []uuid.UUID, // 修复：使用复数形式，支持多个数据集
	query string,
	source string,
) error {
	return s.factory.RecordDatasetQuery(userID, datasetIDs, query, source)
}

// UpdateSegmentHitCount 更新片段命中次数
func (s *RetrieverService) UpdateSegmentHitCount(segmentIDs []uuid.UUID) error {
	return s.factory.UpdateSegmentHitCount(segmentIDs)
}
