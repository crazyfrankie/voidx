package resp

import (
	"github.com/google/uuid"
)

// DatasetResp 知识库响应
type DatasetResp struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DocumentCount   int       `json:"document_count"`
	HitCount        int       `json:"hit_count"`
	RelatedAppCount int       `json:"related_app_count"`
	CharacterCount  int       `json:"character_count"`
	Ctime           int64     `json:"ctime"`
	Utime           int64     `json:"utime"`
}

// DocumentResp 文档响应
type DocumentResp struct {
	ID             uuid.UUID `json:"id"`
	DatasetID      uuid.UUID `json:"dataset_id"`
	Position       int       `json:"position"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	SegmentCount   int       `json:"segment_count"`
	CharacterCount int       `json:"character_count"`
	HitCount       int       `json:"hit_count"`
	Enabled        bool      `json:"enabled"`
	DisabledAt     int64     `json:"disabled_at"`
	Ctime          int64     `json:"ctime"`
	Utime          int64     `json:"utime"`
}

type DocumentStatusResp struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	Size                  int64     `json:"size"`
	Extension             string    `json:"extension"`
	MimeType              string    `json:"mime_type"`
	Position              int       `json:"position"`
	SegmentCount          int       `json:"segment_count"`
	CompletedSegmentCount int       `json:"completed_segment_count"`
	Status                string    `json:"status"`
	ProcessingStartedAt   int64     `json:"processing_started_at"`
	ParsingCompletedAt    int64     `json:"parsing_completed_at"`
	SplittingCompletedAt  int64     `json:"splitting_completed_at"`
	IndexingCompletedAt   int64     `json:"indexing_completed_at"`
	CompletedAt           int64     `json:"completed_at"`
	StoppedAt             int64     `json:"stopped_at"`
	Ctime                 int64     `json:"ctime"`
}

// SegmentResp 片段响应
type SegmentResp struct {
	ID          uuid.UUID  `json:"id"`
	DatasetID   uuid.UUID  `json:"dataset_id"`
	DocumentID  uuid.UUID  `json:"document_id"`
	Position    int        `json:"position"`
	Content     string     `json:"content"`
	WordCount   int        `json:"word_count"`
	TokensCount int        `json:"tokens_count"`
	Keywords    []string   `json:"keywords"`
	HitCount    int        `json:"hit_count"`
	Enabled     bool       `json:"enabled"`
	DisabledAt  *int64     `json:"disabled_at"`
	DisabledBy  *uuid.UUID `json:"disabled_by"`
	Status      string     `json:"status"`
	Ctime       int64      `json:"ctime"`
	Utime       int64      `json:"utime"`
}

// DatasetQueryResp 知识库查询记录响应
type DatasetQueryResp struct {
	ID          uuid.UUID  `json:"id"`
	DatasetID   uuid.UUID  `json:"dataset_id"`
	Content     string     `json:"content"`
	Source      string     `json:"source"`
	SourceAppID *uuid.UUID `json:"source_app_id"`
	Ctime       int64      `json:"ctime"`
}

// HitResult 检索结果
type HitResult struct {
	Document       DocumentResp `json:"document"`
	SegmentID      uuid.UUID    `json:"segment_id"`
	DocumentID     uuid.UUID    `json:"document_id"`
	DatasetID      uuid.UUID    `json:"dataset_id"`
	Content        string       `json:"content"`
	Score          float64      `json:"score"`
	Keywords       []string     `json:"keywords"`
	Position       int          `json:"position"`
	CharacterCount int          `json:"character_count"`
	HitCount       int          `json:"hit_count"`
	TokenCount     int          `json:"token_count"`
	DisabledAt     int64        `json:"disabled_at"`
	Enabled        bool         `json:"enabled"`
	Status         string       `json:"status"`
	Ctime          int64        `json:"ctime"`
	Utime          int64        `json:"utime"`
}
