package req

import "github.com/google/uuid"

// CreateDatasetReq 创建知识库请求
type CreateDatasetReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	Permission  string `json:"permission" binding:"required,oneof=only_me all_team_members"`
	IndexStruct string `json:"index_struct" binding:"required,oneof=high_quality economy"`
}

// UpdateDatasetReq 更新知识库请求
type UpdateDatasetReq struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Permission  string `json:"permission" binding:"omitempty,oneof=only_me all_team_members"`
}

// GetDatasetsWithPageReq 获取知识库分页列表请求
type GetDatasetsWithPageReq struct {
	CurrentPage int    `form:"current_page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
	SearchWord  string `form:"search_word"`
}

// HitReq 检索测试请求
type HitReq struct {
	Query            string  `json:"query" binding:"required,max=2000"`
	RetrievalModel   string  `json:"retrieval_model" binding:"required,oneof=semantic full_text hybrid"`
	K                int     `json:"k" binding:"required,min=1,max=20"`
	Score            float32 `json:"score" binding:"required,min=0,max=1"`
	RerankingModel   string  `json:"reranking_model"`
	RerankingEnabled bool    `json:"reranking_enabled"`
}

// CreateDocumentsReq 创建文档请求
type CreateDocumentsReq struct {
	UploadFileIDs []uuid.UUID    `json:"upload_file_ids"`
	ProcessType   string         `json:"process_type"`
	Rule          map[string]any `json:"rule"`
}

// UpdateDocumentReq 更新文档请求
type UpdateDocumentReq struct {
	Name           string         `json:"name" binding:"omitempty,max=255"`
	Text           string         `json:"text"`
	ProcessingRule map[string]any `json:"processing_rule"`
	Enabled        *bool          `json:"enabled"`
}

// GetDocumentsWithPageReq 获取文档分页列表请求
type GetDocumentsWithPageReq struct {
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100"`
	SearchWord string `form:"search_word"`
	Status     string `form:"status"`
}

// CreateSegmentReq 创建片段请求
type CreateSegmentReq struct {
	Content  string   `json:"content" binding:"required"`
	Keywords []string `json:"keywords"`
}

// UpdateSegmentReq 更新片段请求
type UpdateSegmentReq struct {
	Content       string   `json:"content" binding:"omitempty"`
	Keywords      []string `json:"keywords"`
	EnabledStatus *bool    `json:"enabled_status"`
}

// GetSegmentsWithPageReq 获取片段分页列表请求
type GetSegmentsWithPageReq struct {
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100"`
	SearchWord string `form:"search_word"`
	Status     string `form:"status"`
}
