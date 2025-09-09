package retrievers

import (
	"context"
	"fmt"
	"sort"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// HybridRetriever 混合检索器，结合全文检索和语义检索
type HybridRetriever struct {
	fullTextRetriever retriever.Retriever
	semanticRetriever retriever.Retriever
	datasetIDs        []uuid.UUID
	fullTextWeight    float64
	semanticWeight    float64
}

// NewHybridRetriever 创建一个新的混合检索器，支持多个数据集
func NewHybridRetriever(fullTextRetriever retriever.Retriever, semanticRetriever retriever.Retriever,
	datasetIDs []uuid.UUID, options map[string]any) *HybridRetriever {
	// 默认权重配置
	fullTextWeight := 0.5
	semanticWeight := 0.5

	if weight, ok := options["full_text_weight"].(float64); ok {
		fullTextWeight = weight
	}
	if weight, ok := options["semantic_weight"].(float64); ok {
		semanticWeight = weight
	}

	// 确保权重总和为1
	totalWeight := fullTextWeight + semanticWeight
	if totalWeight > 0 {
		fullTextWeight = fullTextWeight / totalWeight
		semanticWeight = semanticWeight / totalWeight
	}

	return &HybridRetriever{
		fullTextRetriever: fullTextRetriever,
		semanticRetriever: semanticRetriever,
		datasetIDs:        datasetIDs,
		fullTextWeight:    fullTextWeight,
		semanticWeight:    semanticWeight,
	}
}

func (r *HybridRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 并行执行全文检索和语义检索
	fullTextChan := make(chan retrievalResult, 1)
	semanticChan := make(chan retrievalResult, 1)

	// 启动全文检索
	go func() {
		docs, err := r.fullTextRetriever.Retrieve(ctx, query)
		fullTextChan <- retrievalResult{docs: docs, err: err}
	}()

	// 启动语义检索
	go func() {
		docs, err := r.semanticRetriever.Retrieve(ctx, query)
		semanticChan <- retrievalResult{docs: docs, err: err}
	}()

	// 等待两个检索结果
	fullTextResult := <-fullTextChan
	semanticResult := <-semanticChan

	// 检查错误
	if fullTextResult.err != nil {
		return nil, fmt.Errorf("full text retrieval failed: %w", fullTextResult.err)
	}
	if semanticResult.err != nil {
		return nil, fmt.Errorf("semantic retrieval failed: %w", semanticResult.err)
	}

	// 2. 合并和重新排序结果
	mergedDocs := r.mergeAndRerankResults(fullTextResult.docs, semanticResult.docs)

	return mergedDocs, nil
}

// retrievalResult 检索结果结构
type retrievalResult struct {
	docs []*schema.Document
	err  error
}

// mergeAndRerankResults 合并和重新排序结果
func (r *HybridRetriever) mergeAndRerankResults(fullTextDocs []*schema.Document, semanticDocs []*schema.Document) []*schema.Document {
	// 创建文档映射，以segment_id为键
	docMap := make(map[string]*hybridDocument)

	// 处理全文检索结果
	for i, doc := range fullTextDocs {
		segmentID := r.getSegmentID(doc)
		if segmentID == "" {
			continue
		}

		// 验证文档是否属于指定的数据集
		if !r.isFromValidDataset(doc) {
			continue
		}

		hybridDoc := &hybridDocument{
			document:      doc,
			fullTextRank:  i + 1,
			semanticRank:  -1,
			fullTextScore: 1.0 / float64(i+1), // 基于排名的分数
			semanticScore: 0.0,
			combinedScore: 0.0,
		}
		docMap[segmentID] = hybridDoc
	}

	// 处理语义检索结果
	for i, doc := range semanticDocs {
		segmentID := r.getSegmentID(doc)
		if segmentID == "" {
			continue
		}

		// 验证文档是否属于指定的数据集 - 修复：支持多个数据集
		if !r.isFromValidDataset(doc) {
			continue
		}

		if hybridDoc, exists := docMap[segmentID]; exists {
			// 文档在两个结果中都存在
			hybridDoc.semanticRank = i + 1
			if score, ok := doc.MetaData["score"].(float64); ok {
				hybridDoc.semanticScore = score
			} else if score, ok := doc.MetaData["score"].(float32); ok {
				hybridDoc.semanticScore = float64(score)
			} else {
				hybridDoc.semanticScore = 1.0 / float64(i+1)
			}
		} else {
			// 文档只在语义检索结果中存在
			semanticScore := 0.0
			if score, ok := doc.MetaData["score"].(float64); ok {
				semanticScore = score
			} else if score, ok := doc.MetaData["score"].(float32); ok {
				semanticScore = float64(score)
			} else {
				semanticScore = 1.0 / float64(i+1)
			}

			hybridDoc := &hybridDocument{
				document:      doc,
				fullTextRank:  -1,
				semanticRank:  i + 1,
				fullTextScore: 0.0,
				semanticScore: semanticScore,
				combinedScore: 0.0,
			}
			docMap[segmentID] = hybridDoc
		}
	}

	// 计算组合分数
	for _, hybridDoc := range docMap {
		hybridDoc.combinedScore = r.fullTextWeight*hybridDoc.fullTextScore + r.semanticWeight*hybridDoc.semanticScore
		// 更新文档元数据中的分数
		hybridDoc.document.MetaData["score"] = hybridDoc.combinedScore
		hybridDoc.document.MetaData["full_text_score"] = hybridDoc.fullTextScore
		hybridDoc.document.MetaData["semantic_score"] = hybridDoc.semanticScore
	}

	// 转换为切片并按组合分数排序
	hybridDocs := make([]*hybridDocument, 0, len(docMap))
	for _, hybridDoc := range docMap {
		hybridDocs = append(hybridDocs, hybridDoc)
	}

	sort.Slice(hybridDocs, func(i, j int) bool {
		return hybridDocs[i].combinedScore > hybridDocs[j].combinedScore
	})

	// 提取最终文档列表
	result := make([]*schema.Document, 0, len(hybridDocs))
	for _, hybridDoc := range hybridDocs {
		result = append(result, hybridDoc.document)
	}

	return result
}

// hybridDocument 混合文档结构
type hybridDocument struct {
	document      *schema.Document
	fullTextRank  int
	semanticRank  int
	fullTextScore float64
	semanticScore float64
	combinedScore float64
}

// getSegmentID 从文档元数据中获取segment_id
func (r *HybridRetriever) getSegmentID(doc *schema.Document) string {
	if segmentID, ok := doc.MetaData["segment_id"].(string); ok {
		return segmentID
	}
	return ""
}

// isFromValidDataset 检查文档是否来自有效的数据集
func (r *HybridRetriever) isFromValidDataset(doc *schema.Document) bool {
	datasetIDStr, ok := doc.MetaData["dataset_id"].(string)
	if !ok {
		return false
	}

	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		return false
	}

	// 检查是否在允许的数据集列表中
	for _, validID := range r.datasetIDs {
		if validID == datasetID {
			return true
		}
	}

	return false
}
