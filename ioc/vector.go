package ioc

import (
	"context"
	"os"
	"time"

	"github.com/crazyfrankie/voidx/infra/contract/document/vecstore"
	"github.com/crazyfrankie/voidx/infra/contract/embedding"
	"github.com/crazyfrankie/voidx/infra/impl/document/vecstore/milvus"
	"github.com/crazyfrankie/voidx/pkg/lang/ptr"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func InitVectorStore(emb embedding.Embedder) vecstore.Manager {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	milvusAddr := os.Getenv("MILVUS_ADDR")
	user := os.Getenv("MILVUS_USER")
	password := os.Getenv("MILVUS_PASSWORD")
	mc, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  milvusAddr,
		Username: user,
		Password: password,
	})
	if err != nil {
		panic(err)
	}

	mgr, err := milvus.NewManager(&milvus.ManagerConfig{
		Client:       mc,
		Embedding:    emb,
		EnableHybrid: ptr.Of(true),
	})
	if err != nil {
		panic(err)
	}

	return mgr
}
