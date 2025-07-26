package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
)

// RetrievalService 检索服务，提供统一的检索接口
type RetrievalService struct {
	RetrieverFactory *retrievers.RetrieverFactory
}

// NewRetrievalService 创建一个新的检索服务
func NewRetrievalService(retrieverFactory *retrievers.RetrieverFactory) *RetrievalService {
	return &RetrievalService{
		RetrieverFactory: retrieverFactory,
	}
}

// CreateLangchainToolFromSearch 根据传递的参数构建一个LangChain知识库搜索工具
func (s *RetrievalService) CreateLangchainToolFromSearch(ctx context.Context, userID uuid.UUID, datasets []uuid.UUID,
	resource consts.RetrievalSource, retrievalConfig map[string]any) (tools.Tool, error) {
	// 解析检索配置参数
	k, ok := retrievalConfig["k"].(int)
	if !ok {
		k = 4 // 默认值
	}
	score, ok := retrievalConfig["score"].(float64)
	if !ok {
		score = 0 // 默认值
	}
	retrievalStrategy, ok := retrievalConfig["retrieval_strategy"].(consts.RetrievalStrategy)
	if !ok {
		retrievalStrategy = consts.RetrievalStrategySemantic // 默认语义检索
	}

	// 定义工具输入结构
	inputSchema := DatasetRetrievalInput{}

	// 创建工具实例
	tool := &datasetRetrievalTool{
		service:           s,
		userID:            userID,
		datasetIDs:        datasets,
		resource:          resource,
		retrievalConfig:   retrievalConfig,
		k:                 k,
		score:             float32(score),
		retrievalStrategy: retrievalStrategy,
		inputSchema:       inputSchema,
	}

	return tool, nil
}

// SearchInDatasets 在指定数据集中执行检索
func (s *RetrievalService) SearchInDatasets(ctx context.Context, userID uuid.UUID, searchReq req.SearchRequest) ([]resp.SearchResult, error) {
	// 1. 验证数据集权限
	datasets, err := s.validateDatasetAccess(ctx, userID, searchReq.DatasetIDs)
	if err != nil {
		return nil, err
	}

	if len(datasets) == 0 {
		return nil, fmt.Errorf("当前无知识库可执行检索")
	}

	// 2. 设置默认参数
	if searchReq.K <= 0 {
		searchReq.K = 4
	}
	if searchReq.RetrieverType == "" {
		searchReq.RetrieverType = "semantic"
	}

	// 3. 创建检索选项
	options := map[string]any{
		"k":               searchReq.K,
		"score_threshold": searchReq.ScoreThreshold,
	}

	// 4. 创建检索器
	retriever, err := s.createRetriever(searchReq.RetrieverType, searchReq.DatasetIDs, options)
	if err != nil {
		return nil, err
	}

	// 5. 执行检索
	documents, err := s.retrieveDocuments(ctx, retriever, searchReq.Query)
	if err != nil {
		return nil, err
	}

	// 6. 转换为检索结果
	results := make([]resp.SearchResult, 0, len(documents))
	for _, doc := range documents {
		result := s.convertDocumentToSearchResult(doc)
		results = append(results, result)
	}

	// 7. 记录查询历史和更新命中次数
	go s.recordSearchHistory(ctx, userID, searchReq, results)

	return results, nil
}

// createRetriever 创建检索器
func (s *RetrievalService) createRetriever(retrieverType string, datasetIDs []uuid.UUID, options map[string]any) (interface{}, error) {
	// 默认使用混合检索
	if retrieverType == "" {
		retrieverType = string(retrievers.RetrieverTypeHybrid)
	}

	return s.RetrieverFactory.CreateRetriever(retrievers.RetrieverType(retrieverType), datasetIDs, options)
}

// retrieveDocuments 执行文档检索
func (s *RetrievalService) retrieveDocuments(ctx context.Context, retriever interface{}, query string) ([]schema.Document, error) {
	switch r := retriever.(type) {
	case *retrievers.FullTextRetriever:
		return r.GetRelevantDocuments(ctx, query)
	case *retrievers.SemanticRetriever:
		return r.GetRelevantDocuments(ctx, query)
	case *retrievers.HybridRetriever:
		return r.GetRelevantDocuments(ctx, query)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %T", retriever)
	}
}

// validateDatasetAccess 验证数据集访问权限
func (s *RetrievalService) validateDatasetAccess(ctx context.Context, userID uuid.UUID, datasetIDs []uuid.UUID) ([]uuid.UUID, error) {
	// 使用RetrieverFactory验证数据集访问权限
	validDatasetIDs, err := s.RetrieverFactory.ValidateDatasetAccess(userID, datasetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to validate dataset access: %w", err)
	}

	if len(validDatasetIDs) == 0 {
		return nil, fmt.Errorf("no accessible datasets found")
	}

	return validDatasetIDs, nil
}

// convertDocumentToSearchResult 将文档转换为检索结果
func (s *RetrievalService) convertDocumentToSearchResult(doc schema.Document) resp.SearchResult {
	result := resp.SearchResult{
		Content: doc.PageContent,
	}

	// 从元数据中提取信息
	if segmentID, ok := doc.Metadata["segment_id"].(string); ok {
		if id, err := uuid.Parse(segmentID); err == nil {
			result.SegmentID = id
		}
	}

	if documentID, ok := doc.Metadata["document_id"].(string); ok {
		if id, err := uuid.Parse(documentID); err == nil {
			result.DocumentID = id
		}
	}

	if datasetID, ok := doc.Metadata["dataset_id"].(string); ok {
		if id, err := uuid.Parse(datasetID); err == nil {
			result.DatasetID = id
		}
	}

	if documentName, ok := doc.Metadata["document_name"].(string); ok {
		result.DocumentName = documentName
	}

	if score, ok := doc.Metadata["score"].(float64); ok {
		result.Score = score
	}

	if position, ok := doc.Metadata["position"].(int); ok {
		result.Position = position
	}

	if keywords, ok := doc.Metadata["keywords"].([]string); ok {
		result.Keywords = keywords
	}

	if characterCount, ok := doc.Metadata["character_count"].(int); ok {
		result.CharacterCount = characterCount
	}

	if tokenCount, ok := doc.Metadata["token_count"].(int); ok {
		result.TokenCount = tokenCount
	}

	if hitCount, ok := doc.Metadata["hit_count"].(int); ok {
		result.HitCount = hitCount
	}

	if enabled, ok := doc.Metadata["enabled"].(bool); ok {
		result.Enabled = enabled
	}

	if status, ok := doc.Metadata["status"].(string); ok {
		result.Status = status
	}

	if ctime, ok := doc.Metadata["ctime"].(int64); ok {
		result.Ctime = ctime
	}

	if utime, ok := doc.Metadata["utime"].(int64); ok {
		result.Utime = utime
	}

	return result
}

// recordSearchHistory 记录搜索历史和更新命中次数
func (s *RetrievalService) recordSearchHistory(ctx context.Context, userID uuid.UUID, searchReq req.SearchRequest, results []resp.SearchResult) {
	// 1. 记录数据集查询历史
	uniqueDatasetIDs := make(map[uuid.UUID]bool)
	for _, result := range results {
		uniqueDatasetIDs[result.DatasetID] = true
	}

	// 为每个唯一的数据集记录查询历史
	for datasetID := range uniqueDatasetIDs {
		err := s.RetrieverFactory.RecordDatasetQuery(userID, datasetID, searchReq.Query, "hit_testing")
		if err != nil {
			fmt.Printf("Failed to record search history for dataset %s: %v\n", datasetID, err)
		}
	}

	// 2. 更新片段命中次数
	segmentIDs := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		segmentIDs = append(segmentIDs, result.SegmentID)
	}

	if len(segmentIDs) > 0 {
		err := s.RetrieverFactory.UpdateSegmentHitCount(segmentIDs)
		if err != nil {
			fmt.Printf("Failed to update segment hit count: %v\n", err)
		}
	}
}
