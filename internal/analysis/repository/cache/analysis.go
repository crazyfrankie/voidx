package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/internal/models/resp"
)

type AnalysisCache struct {
	cmd redis.Cmdable
}

func NewAnalysisCache(cmd redis.Cmdable) *AnalysisCache {
	return &AnalysisCache{
		cmd: cmd,
	}
}

// GetAppAnalysis 从缓存获取应用分析数据
func (c *AnalysisCache) GetAppAnalysis(ctx context.Context, cacheKey string) (*resp.AppAnalysisResp, error) {
	data, err := c.cmd.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var analysis resp.AppAnalysisResp
	if err := json.Unmarshal([]byte(data), &analysis); err != nil {
		return nil, err
	}

	return &analysis, nil
}

// SetAppAnalysis 将应用分析数据存储到缓存
func (c *AnalysisCache) SetAppAnalysis(ctx context.Context, cacheKey string, analysis *resp.AppAnalysisResp) error {
	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	return c.cmd.SetEx(ctx, cacheKey, string(data), 24*time.Hour).Err()
}
