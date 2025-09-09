package retrievers

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// FullTextRetriever 全文检索器
type FullTextRetriever struct {
	db           *gorm.DB
	datasetIDs   []uuid.UUID
	jiebaService *JiebaService
}

// NewFullTextRetriever 创建一个新的全文检索器，支持多个数据集
func NewFullTextRetriever(db *gorm.DB, datasetIDs []uuid.UUID, jiebaService *JiebaService) *FullTextRetriever {
	return &FullTextRetriever{
		db:           db,
		datasetIDs:   datasetIDs,
		jiebaService: jiebaService,
	}
}

func (r *FullTextRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1.将查询query转换成关键词列表
	keywords := r.jiebaService.ExtractKeywords(query, 10)
	if len(keywords) == 0 {
		return []*schema.Document{}, nil
	}

	keywordTables, err := r.getKeywordTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyword tables: %w", err)
	}

	// 3.遍历所有的知识库关键词表，找到匹配query关键词的id列表
	var allIDs []string
	for _, keywordTable := range keywordTables {
		// 4.遍历每一个关键词表的每一项
		for keyword, segmentIDs := range keywordTable {
			// 5.如果数据存在则提取关键词对应的片段id列表
			for _, kw := range keywords {
				if keyword == kw {
					allIDs = append(allIDs, segmentIDs...)
				}
			}
		}
	}

	if len(allIDs) == 0 {
		return []*schema.Document{}, nil
	}
	// 6.统计segment_id出现的频率
	idCounter := make(map[string]int)
	for _, id := range allIDs {
		idCounter[id]++
	}

	// 7. 根据得到的id列表检索数据库得到片段列表信息
	segments, err := r.getSegmentsByIDs(ctx, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get segments: %w", err)
	}

	// 8. 根据频率进行排序
	sortedSegments := r.sortSegmentsByFrequency(segments, allIDs, idCounter)

	// 9. 构建 Document 列表
	documents := make([]*schema.Document, 0, len(sortedSegments))
	for _, segment := range sortedSegments {
		doc := &schema.Document{
			Content: segment.Content,
			MetaData: map[string]any{
				"account_id":       segment.AccountID.String(),
				"dataset_id":       segment.DatasetID.String(),
				"document_id":      segment.DocumentID.String(),
				"segment_id":       segment.ID.String(),
				"node_id":          segment.NodeID,
				"document_enabled": true,
				"segment_enabled":  true,
				"score":            0,
			},
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// getKeywordTables 获取知识库关键词表
func (r *FullTextRetriever) getKeywordTables(ctx context.Context) ([]map[string][]string, error) {
	var keywords []entity.Keyword

	// 查询指定数据集的关键词表
	err := r.db.Where("dataset_id IN ?", r.datasetIDs).Find(&keywords).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query keyword tables: %w", err)
	}

	// 转换为 map 格式
	result := make([]map[string][]string, 0, len(keywords))
	for _, kt := range keywords {
		if kt.KeywordMap != nil {
			result = append(result, kt.KeywordMap)
		}
	}

	return result, nil
}

// getSegmentsByIDs 根据ID获取片段
func (r *FullTextRetriever) getSegmentsByIDs(ctx context.Context, ids []string) ([]entity.Segment, error) {
	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if parsedUUID, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, parsedUUID)
		}
	}

	if len(uuids) == 0 {
		return []entity.Segment{}, nil
	}

	var segments []entity.Segment
	err := r.db.WithContext(ctx).
		Where("id IN ? AND dataset_id IN ?", uuids, r.datasetIDs).
		Find(&segments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query segments: %w", err)
	}

	return segments, nil
}

// sortSegmentsByFrequency 根据频率对片段进行排序
func (r *FullTextRetriever) sortSegmentsByFrequency(segments []entity.Segment, topKIDs []string, idCounter map[string]int) []entity.Segment {
	// 创建片段映射
	segmentMap := make(map[string]entity.Segment)
	for _, segment := range segments {
		segmentMap[segment.ID.String()] = segment
	}

	// 按照频率排序返回片段
	sortedSegments := make([]entity.Segment, 0, len(topKIDs))
	for _, id := range topKIDs {
		if segment, exists := segmentMap[id]; exists {
			sortedSegments = append(sortedSegments, segment)
		}
	}

	return sortedSegments
}
