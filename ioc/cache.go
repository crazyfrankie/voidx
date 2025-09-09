package ioc

import (
	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/infra/contract/cache"
	cacheimpl "github.com/crazyfrankie/voidx/infra/impl/cache/redis"
)

func InitCache() cache.Cmdable {
	return cacheimpl.NewWithAddrAndPassword(conf.GetConf().Redis.Addr, "")
}
