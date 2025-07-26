package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/analysis/repository/cache"
	"github.com/crazyfrankie/voidx/internal/analysis/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
)

type AnalysisRepo struct {
	dao   *dao.AnalysisDao
	cache *cache.AnalysisCache
}

func NewAnalysisRepo(d *dao.AnalysisDao, c *cache.AnalysisCache) *AnalysisRepo {
	return &AnalysisRepo{
		dao:   d,
		cache: c,
	}
}

// GetAppAnalysisFromCache 从缓存获取应用分析数据
func (r *AnalysisRepo) GetAppAnalysisFromCache(ctx context.Context, appID uuid.UUID) (*resp.AppAnalysisResp, error) {
	cacheKey := r.buildCacheKey(appID)
	return r.cache.GetAppAnalysis(ctx, cacheKey)
}

// SetAppAnalysisToCache 将应用分析数据存储到缓存
func (r *AnalysisRepo) SetAppAnalysisToCache(ctx context.Context, appID uuid.UUID, analysis *resp.AppAnalysisResp) error {
	cacheKey := r.buildCacheKey(appID)
	return r.cache.SetAppAnalysis(ctx, cacheKey, analysis)
}

// GetMessagesByTimeRange 根据时间范围获取消息数据
func (r *AnalysisRepo) GetMessagesByTimeRange(ctx context.Context, appID uuid.UUID, startAt, endAt time.Time) ([]entity.Message, error) {
	return r.dao.GetMessagesByTimeRange(ctx, appID, startAt, endAt)
}

// buildCacheKey 构建缓存键
func (r *AnalysisRepo) buildCacheKey(appID uuid.UUID) string {
	now := time.Now()
	return fmt.Sprintf("%s:%s", now.Format("2006_01_02"), appID.String())
}
