package cache

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type KeyWordCache struct {
	cmd     redis.Cmdable
	lockTTL time.Duration
}

func NewKeyWordCache(cmd redis.Cmdable) *KeyWordCache {
	return &KeyWordCache{
		cmd:     cmd,
		lockTTL: time.Second * 30,
	}
}

func (c *KeyWordCache) AcquireLock(ctx context.Context, key string) string {
	// 生成唯一的锁值
	lockValue := uuid.New().String()

	// 尝试获取锁
	success, err := c.cmd.SetNX(ctx, key, lockValue, c.lockTTL).Result()
	if err != nil || !success {
		return ""
	}

	return lockValue
}

// ReleaseLock 释放锁
func (c *KeyWordCache) ReleaseLock(ctx context.Context, key, value string) (any, error) {
	// 使用Lua脚本确保只释放自己的锁
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`
	return c.cmd.Eval(ctx, script, []string{key}, value).Result()
}
