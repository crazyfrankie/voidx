package entity

import (
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
)

// Dataset 知识库表
type Dataset struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID   uuid.UUID `gorm:"type:uuid;not null;index:dataset_account_id_name_idx,composite:account_name" json:"account_id"`
	Name        string    `gorm:"size:255;not null;default:'';index:dataset_account_id_name_idx,composite:account_name" json:"name"`
	Icon        string    `gorm:"size:255;not null;default:''" json:"icon"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`
	Utime       int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime       int64     `gorm:"autoCreateTime" json:"ctime"`
}

// Document 文档表模型
type Document struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID            uuid.UUID `gorm:"type:uuid;not null;index:document_account_id_idx" json:"account_id"`
	DatasetID            uuid.UUID `gorm:"type:uuid;not null;index:document_dataset_id_idx" json:"dataset_id"`
	UploadFileID         uuid.UUID `gorm:"type:uuid;not null" json:"upload_file_id"`
	ProcessRuleID        uuid.UUID `gorm:"type:uuid;not null" json:"process_rule_id"`
	Batch                string    `gorm:"size:255;not null;default:'';index:document_batch_idx" json:"batch"`
	Name                 string    `gorm:"size:255;not null;default:''" json:"name"`
	Position             int       `gorm:"not null;default:1" json:"position"`
	CharacterCount       int       `gorm:"not null;default:0" json:"character_count"`
	TokenCount           int       `gorm:"not null;default:0" json:"token_count"`
	ProcessingStartedAt  int64     `gorm:"" json:"processing_started_at"`
	ParsingCompletedAt   int64     `gorm:"" json:"parsing_completed_at"`
	SplittingCompletedAt int64     `gorm:"" json:"splitting_completed_at"`
	IndexingCompletedAt  int64     `gorm:"" json:"indexing_completed_at"`
	CompletedAt          int64     `gorm:"" json:"completed_at"`
	StoppedAt            int64     `gorm:"" json:"stopped_at"`
	Enabled              bool      `gorm:"not null;default:false" json:"enabled"`
	DisabledAt           int64     `gorm:"" json:"disabled_at"`
	Status               string    `gorm:"size:255;not null;default:'waiting'" json:"status"`
	Utime                int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime                int64     `gorm:"autoCreateTime" json:"ctime"`
}

// Segment 片段表模型
type Segment struct {
	ID                  uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID           uuid.UUID            `gorm:"type:uuid;not null;index:segment_account_id_idx" json:"account_id"`
	DatasetID           uuid.UUID            `gorm:"type:uuid;not null;index:segment_dataset_id_idx" json:"dataset_id"`
	DocumentID          uuid.UUID            `gorm:"type:uuid;not null;index:segment_document_id_idx" json:"document_id"`
	NodeID              uuid.UUID            `gorm:"type:uuid;not null" json:"node_id"`
	Position            int                  `gorm:"not null;default:1" json:"position"`
	Content             string               `gorm:"type:text;not null;default:''" json:"content"`
	CharacterCount      int                  `gorm:"not null;default:0" json:"character_count"`
	TokenCount          int                  `gorm:"not null;default:0" json:"token_count"`
	Keywords            []string             `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"keywords"`
	Hash                string               `gorm:"size:255;not null;default:''" json:"hash"`
	HitCount            int                  `gorm:"not null;default:0" json:"hit_count"`
	Enabled             bool                 `gorm:"not null;default:false" json:"enabled"`
	DisabledAt          int64                `gorm:"" json:"disabled_at"`
	ProcessingStartedAt int64                `gorm:"" json:"processing_started_at"`
	IndexingCompletedAt int64                `gorm:"" json:"indexing_completed_at"`
	CompletedAt         int64                `gorm:"" json:"completed_at"`
	StoppedAt           int64                `gorm:"" json:"stopped_at"`
	Error               string               `gorm:"type:text;not null;default:''" json:"error"`
	Status              consts.SegmentStatus `gorm:"size:255;not null;default:'waiting'" json:"status"`
	Utime               int64                `gorm:"autoUpdateTime" json:"utime"`
	Ctime               int64                `gorm:"autoCreateTime" json:"ctime"`
}

// Keyword 关键词表模型
type Keyword struct {
	ID         uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DatasetID  uuid.UUID           `gorm:"type:uuid;not null;index:keyword_table_dataset_id_idx" json:"dataset_id"`
	KeywordMap map[string][]string `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"keyword_map"`
	Utime      int64               `gorm:"autoUpdateTime" json:"utime"`
	Ctime      int64               `gorm:"autoCreateTime" json:"ctime"`
}

// DatasetQuery 知识库查询表模型
type DatasetQuery struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DatasetID   uuid.UUID `gorm:"type:uuid;not null;index:dataset_query_dataset_id_idx" json:"dataset_id"`
	SourceAppID uuid.UUID `gorm:"type:uuid;index:dataset_source_app_id_idx" json:"source_app_id"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;index:dataset_created_by_idx" json:"created_by"`
	Query       string    `gorm:"type:text;not null;default:''" json:"query"`
	Source      string    `gorm:"size:255;not null;default:'HitTesting'" json:"source"`
	Utime       int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime       int64     `gorm:"autoCreateTime" json:"ctime"`
}

// ProcessRule 文档处理规则表模型
type ProcessRule struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID      `gorm:"type:uuid;not null;index:process_rule_account_id_idx" json:"account_id"`
	DatasetID uuid.UUID      `gorm:"type:uuid;not null;index:process_rule_dataset_id_idx" json:"dataset_id"`
	Mode      string         `gorm:"size:255;not null;default:'automic'" json:"mode"`
	Rule      map[string]any `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"rule"`
	Utime     int64          `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64          `gorm:"autoCreateTime" json:"ctime"`
}
