package embeddings

import (
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
	_defaultModel         = "BAAI/bge-small-en-v1.5"
	_defaultTask          = "feature-extraction"
)

// Option is a function type that can be used to modify the client.
type Option func(p *OpenAI)

// WithModel is an option for providing the model name to use.
func WithModel(model string) Option {
	return func(p *OpenAI) {
		p.Model = model
	}
}

// WithTask is an option for providing the task to call the model with.
func WithTask(task string) Option {
	return func(p *OpenAI) {
		p.Task = task
	}
}

// WithClient is an option for providing the LLM client.
func WithClient(client openai.LLM) Option {
	return func(p *OpenAI) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *OpenAI) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *OpenAI) {
		p.BatchSize = batchSize
	}
}

func applyOptions(opts ...Option) (*OpenAI, error) {
	o := &OpenAI{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
		Model:         _defaultModel,
		Task:          _defaultTask,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := openai.New()
		if err != nil {
			return nil, err
		}
		o.client = client
	}

	return o, nil
}
