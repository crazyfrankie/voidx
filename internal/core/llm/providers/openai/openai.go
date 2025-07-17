package openai

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Chat represents the OpenAI language model
type Chat struct {
	llms.Model
}

// NewChat creates a new instance of the OpenAI Chat model
func NewChat(options ...openai.Option) (*Chat, error) {
	model, err := openai.New(options...)
	if err != nil {
		return nil, err
	}

	return &Chat{Model: model}, nil
}

// Completion represents the OpenAI completion model
type Completion struct {
	llms.Model
}

// NewCompletion creates a new instance of the OpenAI Completion model
func NewCompletion(options ...openai.Option) (*Completion, error) {
	model, err := openai.New(options...)
	if err != nil {
		return nil, err
	}

	return &Completion{Model: model}, nil
}