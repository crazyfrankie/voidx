package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/pkoukk/tiktoken-go"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/pkg/sonic"
)

type EmbeddingService struct {
	cmd      redis.Cmdable
	embedder embedding.Embedder
}

func NewEmbeddingService(cmd redis.Cmdable, embedder embedding.Embedder) *EmbeddingService {
	return &EmbeddingService{cmd: cmd, embedder: embedder}
}

func (s *EmbeddingService) Embeddings(ctx context.Context, query string) ([]float32, error) {
	// 使用eino的Embedder接口
	embeddings, err := s.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, nil
	}

	// 转换float64到float32
	result := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		result[i] = float32(v)
	}

	return result, nil
}

func (s *EmbeddingService) StoreEmbedded(ctx context.Context, query string, embedded []float32) error {
	hashKey := cacheKey(query)

	data, err := sonic.Marshal(embedded)
	if err != nil {
		return err
	}

	return s.cmd.Set(ctx, hashKey, data, time.Hour*6).Err()
}

func (s *EmbeddingService) GetEmbedded(ctx context.Context, query string) ([]float32, error) {
	hashKey := cacheKey(query)
	val, err := s.cmd.Get(ctx, hashKey).Result()
	if err != nil {
		return nil, err
	}

	var res []float32
	if err := sonic.Unmarshal([]byte(val), &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *EmbeddingService) EmbedText(ctx context.Context, text string) ([]float32, error) {
	return s.Embeddings(ctx, text)
}

func (s *EmbeddingService) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	result, err := s.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to embed texts: %w", err)
	}

	embeddings := make([][]float32, len(result))
	for i, emb := range result {
		embeddings[i] = make([]float32, len(emb))
		for j, v := range emb {
			embeddings[i][j] = float32(v)
		}
	}

	return embeddings, nil
}

func (s *EmbeddingService) CalculateTokenCount(query string) int {
	encoding, err := tiktoken.EncodingForModel("gpt-3.5")
	if err != nil {
		return -1
	}

	return len(encoding.EncodeOrdinary(query))
}

func cacheKey(query string) string {
	hash := sha256.Sum256([]byte(query))
	return "embedding:" + hex.EncodeToString(hash[:])
}
