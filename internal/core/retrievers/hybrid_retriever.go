package retrievers

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
)

// HybridRetriever 混合检索器，结合全文检索和语义检索
type HybridRetriever struct {
	FullTextRetriever *FullTextRetriever
	SemanticRetriever *SemanticRetriever
	SearchOptions     map[string]any
	DatasetIDs        []uuid.UUID
}

// NewHybridRetriever 创建一个新的混合检索器
func NewHybridRetriever(fullTextRetriever *FullTextRetriever, semanticRetriever *SemanticRetriever,
	datasetIDs []uuid.UUID, options map[string]any) *HybridRetriever {
	return &HybridRetriever{
		FullTextRetriever: fullTextRetriever,
		SemanticRetriever: semanticRetriever,
		SearchOptions:     options,
		DatasetIDs:        datasetIDs,
	}
}

// GetRelevantDocuments 执行混合检索获取文档列表
func (r *HybridRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	// 获取检索参数
	k := 4 // 默认返回4条结果
	if kVal, ok := r.SearchOptions["k"]; ok {
		if kInt, ok := kVal.(int); ok {
			k = kInt
		}
	}

	// 执行全文检索
	fullTextDocs, err := r.FullTextRetriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("full text retrieval failed: %w", err)
	}

	// 执行语义检索
	semanticDocs, err := r.SemanticRetriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("semantic retrieval failed: %w", err)
	}

	// 合并结果并去重
	mergedDocs := mergeDocuments(fullTextDocs, semanticDocs)

	// 限制返回结果数量
	if len(mergedDocs) > k {
		mergedDocs = mergedDocs[:k]
	}

	return mergedDocs, nil
}

// mergeDocuments 合并两个文档列表并去重
func mergeDocuments(fullTextDocs, semanticDocs []schema.Document) []schema.Document {
	// 使用map去重
	docMap := make(map[string]schema.Document)

	// 添加全文检索结果
	for _, doc := range fullTextDocs {
		segmentID, ok := doc.Metadata["segment_id"].(string)
		if !ok {
			continue
		}
		doc.Metadata["retrieval_method"] = "full_text"
		docMap[segmentID] = doc
	}

	// 添加语义检索结果，如果已存在则更新分数和检索方法
	for _, doc := range semanticDocs {
		segmentID, ok := doc.Metadata["segment_id"].(string)
		if !ok {
			continue
		}

		if existingDoc, exists := docMap[segmentID]; exists {
			// 如果文档已存在，更新检索方法为混合
			existingDoc.Metadata["retrieval_method"] = "hybrid"

			// 如果语义检索有分数，则使用语义检索的分数
			if score, ok := doc.Metadata["score"].(float32); ok {
				existingDoc.Metadata["score"] = score
			}

			docMap[segmentID] = existingDoc
		} else {
			// 如果文档不存在，添加语义检索结果
			doc.Metadata["retrieval_method"] = "semantic"
			docMap[segmentID] = doc
		}
	}

	// 将map转换为slice
	result := make([]schema.Document, 0, len(docMap))
	for _, doc := range docMap {
		result = append(result, doc)
	}

	// 按分数降序排序
	sort.Slice(result, func(i, j int) bool {
		scoreI, okI := result[i].Metadata["score"].(float32)
		scoreJ, okJ := result[j].Metadata["score"].(float32)

		// 如果两个文档都有分数，按分数降序排序
		if okI && okJ {
			return scoreI > scoreJ
		}

		// 如果只有一个文档有分数，有分数的排在前面
		return okI && !okJ
	})

	return result
}
