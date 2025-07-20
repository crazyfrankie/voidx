package wenxin

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Chat represents the Wenxin language model
// It extends the base OpenAI LLM with Wenxin-specific configurations
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the Wenxin Chat model
func NewChat(options ...openai.Option) (*Chat, error) {
	model, err := openai.New(options...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}