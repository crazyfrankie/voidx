package ollama

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Chat represents the Ollama language model
// It extends the base Ollama LLM with custom configurations
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the Ollama Chat model
func NewChat(options ...ollama.Option) (*Chat, error) {
	// Add default base URL if not provided
	defaultOptions := append([]ollama.Option{
		ollama.WithServerURL("http://60.247.21.102:9432"),
	}, options...)

	model, err := ollama.New(defaultOptions...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}