package moonshot

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Chat represents the Moonshot language model
// It extends the base OpenAI LLM with Moonshot-specific configurations
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the Moonshot Chat model
func NewChat(options ...openai.Option) (*Chat, error) {
	// Add Moonshot-specific options
	moonshotOptions := append(options,
		openai.WithBaseURL("https://api.moonshot.cn/v1"),
	)

	model, err := openai.New(moonshotOptions...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}
