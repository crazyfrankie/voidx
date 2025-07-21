package ioc

import (
	embeddings2 "github.com/crazyfrankie/voidx/pkg/langchainx/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

func InitEmbedding() *embeddings2.OpenAI {
	client, err := openai.New(
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		openai.WithEmbeddingModel("text-embedding-v4"),
	)
	if err != nil {
		panic(err)
	}

	embedding, err := embeddings2.NewOpenAI(embeddings2.WithClient(*client))
	if err != nil {
		panic(err)
	}

	return embedding
}
