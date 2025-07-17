package deepseek

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Chat represents the DeepSeek language model
// It extends the base OpenAI LLM with DeepSeek-specific configurations
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the DeepSeek Chat model
func NewChat(options ...openai.Option) (*Chat, error) {
	// Add DeepSeek-specific options
	deepSeekOptions := append(options,
		openai.WithBaseURL("https://api.deepseek.com/v1"),
	)

	model, err := openai.New(deepSeekOptions...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}
