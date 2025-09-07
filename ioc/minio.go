package ioc

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/infra/contract/storage"
	storageimpl "github.com/crazyfrankie/voidx/infra/impl/storage"
)

func InitMinIO() storage.Storage {
	client, err := minio.New(conf.GetConf().MinIO.Endpoint[0], &minio.Options{
		Creds: credentials.NewStaticV4(conf.GetConf().MinIO.AccessKey, conf.GetConf().MinIO.SecretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	return storageimpl.New(client, conf.GetConf().MinIO.BucketName)
}
