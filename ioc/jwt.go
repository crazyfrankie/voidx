package ioc

import (
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/infra/contract/token"
	tokenimpl "github.com/crazyfrankie/voidx/infra/impl/token"
)

func InitJWT(cmd redis.Cmdable) token.Token {
	return tokenimpl.NewTokenService(cmd, conf.GetConf().JWT.SignAlgo, conf.GetConf().JWT.SecretKey)
}
