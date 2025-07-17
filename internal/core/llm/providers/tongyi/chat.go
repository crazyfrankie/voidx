package tongyi

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Chat represents the Tongyi language model
// It extends the base OpenAI LLM with Tongyi-specific configurations
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the Tongyi Chat model
func NewChat(options ...openai.Option) (*Chat, error) {
	model, err := openai.New(options...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}