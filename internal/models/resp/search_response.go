package resp

import "github.com/google/uuid"

// SearchResult 检索结果项
type SearchResult struct {
	SegmentID      uuid.UUID `json:"segment_id"`
	DocumentID     uuid.UUID `json:"document_id"`
	DocumentName   string    `json:"document_name"`
	DatasetID      uuid.UUID `json:"dataset_id"`
	Score          float64   `json:"score"`
	Position       int       `json:"position"`
	Content        string    `json:"content"`
	Keywords       []string  `json:"keywords"`
	CharacterCount int       `json:"character_count"`
	TokenCount     int       `json:"token_count"`
	HitCount       int       `json:"hit_count"`
	Enabled        bool      `json:"enabled"`
	Status         string    `json:"status"`
	Ctime          int64     `json:"ctime"`
	Utime          int64     `json:"utime"`
}
