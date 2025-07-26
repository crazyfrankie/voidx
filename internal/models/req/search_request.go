package req

import "github.com/google/uuid"

// SearchRequest 检索请求
type SearchRequest struct {
	// Query 查询文本
	Query string `json:"query" binding:"required"`

	// DatasetIDs 数据集ID列表
	DatasetIDs []uuid.UUID `json:"dataset_ids" binding:"required"`

	// RetrieverType 检索器类型: full_text, semantic, hybrid
	RetrieverType string `json:"retriever_type" binding:"omitempty,oneof=full_text semantic hybrid"`

	// K 返回结果数量
	K int `json:"k" binding:"omitempty,gt=0"`

	// ScoreThreshold 分数阈值，低于此分数的结果将被过滤
	ScoreThreshold float32 `json:"score_threshold" binding:"omitempty,gte=0"`

	// Options 其他检索选项
	Options map[string]any `json:"options" binding:"omitempty"`
}
