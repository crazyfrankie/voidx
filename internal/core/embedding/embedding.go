package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/pkg/langchainx/embeddings"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

type EmbeddingService struct {
	cmd      redis.Cmdable
	embedder *embeddings.OpenAI
}

func NewEmbeddingService(cmd redis.Cmdable, embedder *embeddings.OpenAI) *EmbeddingService {
	return &EmbeddingService{cmd: cmd, embedder: embedder}
}

func (s *EmbeddingService) Embeddings(ctx context.Context, query string) ([]float32, error) {
	return s.embedder.EmbedQuery(ctx, query)
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
