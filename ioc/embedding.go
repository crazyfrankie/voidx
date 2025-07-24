package ioc

import (
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/crazyfrankie/voidx/pkg/langchainx/embeddings"
)

func InitEmbedding() *embeddings.OpenAI {
	client, err := openai.New(
		openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		openai.WithEmbeddingModel("text-embedding-v4"),
	)
	if err != nil {
		panic(err)
	}

	embedding, err := embeddings.NewOpenAI(embeddings.WithClient(*client))
	if err != nil {
		panic(err)
	}

	return embedding
}
