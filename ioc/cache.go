package ioc

import (
	"time"
	
	"github.com/crazyfrankie/voidx/conf"
	"github.com/redis/go-redis/v9"
)

func InitCache() redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr:        conf.GetConf().Redis.Addr,
		Password:    "",
		DialTimeout: time.Minute * 5,
	})

	return client
}
