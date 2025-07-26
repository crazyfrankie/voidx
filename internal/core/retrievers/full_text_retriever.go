package retrievers

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// FullTextRetriever 全文检索器
type FullTextRetriever struct {
	db            *gorm.DB
	DatasetIDs    []uuid.UUID
	JiebaService  *JiebaService
	SearchOptions map[string]any
}

// NewFullTextRetriever 创建一个新的全文检索器
func NewFullTextRetriever(
	db *gorm.DB,
	datasetIDs []uuid.UUID,
	jiebaService *JiebaService,
	options map[string]any,
) *FullTextRetriever {
	return &FullTextRetriever{
		db:            db,
		DatasetIDs:    datasetIDs,
		JiebaService:  jiebaService,
		SearchOptions: options,
	}
}

// GetRelevantDocuments 根据传递的query执行关键词检索获取LangChain文档列表
func (r *FullTextRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	// 1.将查询query转换成关键词列表
	keywords := r.JiebaService.ExtractKeywords(query, 10)

	// 2.查找指定知识库的关键词表
	keywordTables, err := r.getKeywordTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get keyword tables: %w", err)
	}

	// 3.遍历所有的知识库关键词表，找到匹配query关键词的id列表
	allIDs := []string{}
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

	// 6.统计segment_id出现的频率
	idCounter := make(map[string]int)
	for _, id := range allIDs {
		idCounter[id]++
	}

	// 7.获取频率最高的前k条数据
	k := 4
	if kVal, ok := r.SearchOptions["k"]; ok {
		if kInt, ok := kVal.(int); ok {
			k = kInt
		}
	}

	topKIDs := getTopKIDs(idCounter, k)

	// 8.根据得到的id列表检索数据库得到片段列表信息
	segments, err := r.getSegmentsByIDs(topKIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get segments: %w", err)
	}

	// 9.构建LangChain文档列表
	documents := make([]schema.Document, 0, len(segments))
	for _, segment := range segments {
		documents = append(documents, schema.Document{
			PageContent: segment.Content,
			Metadata: map[string]any{
				"account_id":       segment.DatasetID.String(), // 修正：使用DatasetID作为account_id
				"dataset_id":       segment.DatasetID.String(),
				"document_id":      segment.DocumentID.String(),
				"segment_id":       segment.ID.String(),
				"node_id":          segment.NodeID,
				"document_enabled": true,
				"segment_enabled":  segment.Status,
				"score":            0,
			},
		})
	}

	return documents, nil
}

// getKeywordTables 获取知识库关键词表
func (r *FullTextRetriever) getKeywordTables() ([]map[string][]string, error) {
	var keywords []entity.Keyword

	err := r.db.Where("dataset_id IN ?", r.DatasetIDs).Find(&keywords).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query keyword tables: %w", err)
	}

	result := make([]map[string][]string, 0, len(keywords))
	for _, kt := range keywords {
		if kt.KeywordMap != nil {
			result = append(result, kt.KeywordMap)
		}
	}

	return result, nil
}

// getSegmentsByIDs 根据ID获取片段
func (r *FullTextRetriever) getSegmentsByIDs(ids []string) ([]entity.Segment, error) {
	var segments []entity.Segment

	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if parsedUUID, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, parsedUUID)
		}
	}

	err := r.db.Where("id IN ?", uuids).Find(&segments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query segments: %w", err)
	}

	// Sort segments by the order of IDs
	segmentMap := make(map[string]entity.Segment)
	for _, segment := range segments {
		segmentMap[segment.ID.String()] = segment
	}

	sortedSegments := make([]entity.Segment, 0, len(ids))
	for _, id := range ids {
		if segment, exists := segmentMap[id]; exists {
			sortedSegments = append(sortedSegments, segment)
		}
	}

	return sortedSegments, nil
}

// getTopKIDs 获取出现频率最高的前K个ID
func getTopKIDs(idCounter map[string]int, k int) []string {
	type idFreq struct {
		id   string
		freq int
	}

	idFreqs := make([]idFreq, 0, len(idCounter))
	for id, freq := range idCounter {
		idFreqs = append(idFreqs, idFreq{id: id, freq: freq})
	}

	// 按频率降序排序
	sort.Slice(idFreqs, func(i, j int) bool {
		return idFreqs[i].freq > idFreqs[j].freq
	})

	// 获取前k个
	result := make([]string, 0, k)
	for i := 0; i < len(idFreqs) && i < k; i++ {
		result = append(result, idFreqs[i].id)
	}

	return result
}
