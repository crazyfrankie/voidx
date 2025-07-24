package ioc

import (
	"context"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/tmc/langchaingo/vectorstores/milvus"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/pkg/langchainx/embeddings"
)

func InitVectorStore(embedder *embeddings.OpenAI) *milvus.Store {
	idx, err := entity.NewIndexAUTOINDEX(entity.L2)
	if err != nil {
		panic(err)
	}
	opts := []milvus.Option{
		milvus.WithIndex(idx),
		milvus.WithEmbedder(embedder),
		milvus.WithCollectionName(conf.GetConf().Milvus.CollectionName),
	}

	store, err := milvus.New(context.Background(), client.Config{
		Address: conf.GetConf().Milvus.Addr,
		DBName:  conf.GetConf().Milvus.DBName,
	}, opts...)
	if err != nil {
		panic(err)
	}

	return &store
}
