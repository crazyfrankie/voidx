package ioc

import (
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/pkg/jwt"
)

func InitJWT(cmd redis.Cmdable) *jwt.TokenService {
	return jwt.NewTokenService(cmd, conf.GetConf().JWT.SignAlgo, conf.GetConf().JWT.SecretKey)
}
