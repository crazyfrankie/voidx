package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/cache"
	"github.com/crazyfrankie/voidx/internal/retriever/repository/dao"
)

// KeywordRepository 关键词表存储库
type KeywordRepository struct {
	dao   *dao.KeywordDao
	cache *cache.KeyWordCache
}

// NewKeywordRepository 创建一个新的关键词表存储库
func NewKeywordRepository(d *dao.KeywordDao, c *cache.KeyWordCache) *KeywordRepository {
	return &KeywordRepository{
		dao:   d,
		cache: c,
	}
}

// GetByDatasetID 根据数据集ID获取关键词表
func (r *KeywordRepository) GetByDatasetID(ctx context.Context, datasetID uuid.UUID) (*entity.Keyword, error) {
	return r.dao.GetByDatasetID(ctx, datasetID)
}

// Create 创建关键词表
func (r *KeywordRepository) Create(ctx context.Context, keywordTable *entity.Keyword) error {
	return r.dao.Create(ctx, keywordTable)
}

// Update 更新关键词表
func (r *KeywordRepository) Update(ctx context.Context, keywordTableID uuid.UUID, updates map[string]any) error {
	return r.dao.Update(ctx, keywordTableID, updates)
}

func (r *KeywordRepository) GetKeywordBySegments(ctx context.Context, segmentIDs []uuid.UUID) ([]entity.Segment, error) {
	return r.dao.GetKeywordBySegments(ctx, segmentIDs)
}

func (r *KeywordRepository) AcquireLock(ctx context.Context, key string) string {
	return r.cache.AcquireLock(ctx, key)
}

func (r *KeywordRepository) ReleaseLock(ctx context.Context, key, value string) (any, error) {
	return r.cache.ReleaseLock(ctx, key, value)
}
