package retrievers

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"

	model "github.com/crazyfrankie/voidx/internal/models/entity"
)

// ScoredSegment represents a document segment with its relevance score
type ScoredSegment struct {
	Segment *model.DocumentSegment
	Score   float64
}

// SemanticRetriever implements semantic search using vector embeddings with eino
type SemanticRetriever struct {
	db             *gorm.DB
	embedder       embedding.Embedder
	datasetIDs     uuid.UUID
	scoreThreshold float64
}

// NewSemanticRetriever creates a new semantic retriever
func NewSemanticRetriever(db *gorm.DB, embedder embedding.Embedder) *SemanticRetriever {
	return &SemanticRetriever{
		db:             db,
		embedder:       embedder,
		scoreThreshold: 0.0,
	}
}

// WithDatasetID sets the dataset ID for retrieval
func (r *SemanticRetriever) WithDatasetID(datasetID uuid.UUID) *SemanticRetriever {
	r.datasetIDs = datasetID
	return r
}

// WithScoreThreshold sets the minimum similarity score threshold
func (r *SemanticRetriever) WithScoreThreshold(threshold float64) *SemanticRetriever {
	r.scoreThreshold = threshold
	return r
}

// Retrieve implements the eino retriever interface
func (r *SemanticRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// Generate embedding for the query
	queryEmbeddings, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(queryEmbeddings) == 0 || len(queryEmbeddings[0]) == 0 {
		return nil, fmt.Errorf("empty query embedding")
	}

	queryVector := queryEmbeddings[0]

	// Query document segments from database
	var segments []model.DocumentSegment
	queryBuilder := r.db.Where("dataset_id IN ? AND is_deleted = ?", r.datasetIDs, false)

	// Add enabled filter if available
	queryBuilder = queryBuilder.Where("enabled = ?", true)

	result := queryBuilder.Find(&segments)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query document segments: %w", result.Error)
	}

	// Calculate similarity scores
	var scoredSegments []ScoredSegment
	for i := range segments {
		segment := &segments[i]

		// Skip segments without embeddings
		if len(segment.Embedding) == 0 {
			continue
		}

		// Calculate cosine similarity
		score := r.cosineSimilarity(queryVector, segment.Embedding)

		// Apply score threshold
		if score >= r.scoreThreshold {
			scoredSegments = append(scoredSegments, ScoredSegment{
				Segment: segment,
				Score:   score,
			})
		}
	}

	// Sort by score in descending order
	sort.Slice(scoredSegments, func(i, j int) bool {
		return scoredSegments[i].Score > scoredSegments[j].Score
	})

	// Convert to eino documents
	var documents []*schema.Document
	for _, scored := range scoredSegments {
		segment := scored.Segment

		// Create document metadata
		metadata := map[string]interface{}{
			"segment_id":  segment.ID.String(),
			"document_id": segment.DocumentID.String(),
			"dataset_id":  segment.DatasetID.String(),
			"position":    segment.Position,
			"word_count":  segment.WordCount,
			"token_count": segment.TokenCount,
			"score":       scored.Score,
			"hit_count":   segment.HitCount,
			"hash":        segment.Hash,
			"created_at":  segment.Ctime,
			"updated_at":  segment.Utime,
		}

		// Add document information if available
		if segment.Document != nil {
			metadata["document_name"] = segment.Document.Name
		}

		doc := &schema.Document{
			Content:  segment.Content,
			MetaData: metadata,
		}

		documents = append(documents, doc)
	}

	// Update hit counts asynchronously
	go r.updateHitCounts(ctx, scoredSegments)

	return documents, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func (r *SemanticRetriever) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}

// updateHitCounts updates hit counts for retrieved segments
func (r *SemanticRetriever) updateHitCounts(ctx context.Context, scoredSegments []ScoredSegment) {
	for _, scored := range scoredSegments {
		// Increment hit count
		r.db.Model(&model.DocumentSegment{}).
			Where("id = ?", scored.Segment.ID).
			Update("hit_count", gorm.Expr("hit_count + ?", 1))
	}
}
