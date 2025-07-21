package embeddings

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAI is the embedder using the OpenAI hub api.
type OpenAI struct {
	client *openai.LLM
	Model  string
	Task   string

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = &OpenAI{}

func NewOpenAI(opts ...Option) (*OpenAI, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (e *OpenAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		curBatchEmbeddings, err := e.client.CreateEmbedding(ctx, batch)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curBatchEmbeddings...)
	}

	return emb, nil
}

func (e *OpenAI) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
