package entity

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// ModelType represents the type of language model
type ModelType string

const (
	ModelTypeChat       ModelType = "chat"
	ModelTypeCompletion ModelType = "completion"
)

// ModelFeature represents capabilities supported by a model
type ModelFeature string

const (
	FeatureToolCall     ModelFeature = "tool_call"
	FeatureAgentThought ModelFeature = "agent_thought"
	FeatureImageInput   ModelFeature = "image_input"
)

// ModelParameterType represents the type of model parameters
type ModelParameterType string

const (
	ParameterTypeFloat   ModelParameterType = "float"
	ParameterTypeInt     ModelParameterType = "int"
	ParameterTypeString  ModelParameterType = "string"
	ParameterTypeBoolean ModelParameterType = "boolean"
)

// DefaultModelParameterName represents standard parameter names
type DefaultModelParameterName string

const (
	ParameterTemperature      DefaultModelParameterName = "temperature"
	ParameterTopP             DefaultModelParameterName = "top_p"
	ParameterPresencePenalty  DefaultModelParameterName = "presence_penalty"
	ParameterFrequencyPenalty DefaultModelParameterName = "frequency_penalty"
	ParameterMaxTokens        DefaultModelParameterName = "max_tokens"
)

// ModelParameterOption represents a configuration option for model parameters
type ModelParameterOption struct {
	Label string      `json:"label" yaml:"label"`
	Value interface{} `json:"value" yaml:"value"`
}

// ModelParameter represents a parameter configuration for a model
type ModelParameter struct {
	Name      string                 `json:"name" yaml:"name"`
	Label     string                 `json:"label" yaml:"label"`
	Type      ModelParameterType     `json:"type" yaml:"type"`
	Help      string                 `json:"help" yaml:"help"`
	Required  bool                   `json:"required" yaml:"required"`
	Default   interface{}            `json:"default" yaml:"default"`
	Min       *float64               `json:"min,omitempty" yaml:"min,omitempty"`
	Max       *float64               `json:"max,omitempty" yaml:"max,omitempty"`
	Precision int                    `json:"precision" yaml:"precision"`
	Options   []ModelParameterOption `json:"options" yaml:"options"`
}

// ModelEntity represents a language model configuration
type ModelEntity struct {
	ModelName       string                 `json:"model_name" yaml:"model"`
	Label           string                 `json:"label" yaml:"label"`
	ModelType       ModelType              `json:"model_type" yaml:"model_type"`
	Features        []ModelFeature         `json:"features" yaml:"features"`
	ContextWindow   int                    `json:"context_window" yaml:"context_window"`
	MaxOutputTokens int                    `json:"max_output_tokens" yaml:"max_output_tokens"`
	Attributes      map[string]interface{} `json:"attributes" yaml:"attributes"`
	Parameters      []ModelParameter       `json:"parameters" yaml:"parameters"`
	Metadata        map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ProviderEntity represents a language model provider configuration
type ProviderEntity struct {
	Name                string      `json:"name" yaml:"name"`
	Label               string      `json:"label" yaml:"label"`
	Description         string      `json:"description" yaml:"description"`
	Icon                string      `json:"icon" yaml:"icon"`
	Background          string      `json:"background" yaml:"background"`
	SupportedModelTypes []ModelType `json:"supported_model_types" yaml:"supported_model_types"`
}

// BaseLanguageModel is the interface that all language models must implement
type BaseLanguageModel interface {
	llms.Model
	GetFeatures() []ModelFeature
	GetMetadata() map[string]interface{}
	GetPricing() (float64, float64, float64) // input_price, output_price, unit
	ConvertToHumanMessage(query string, imageURLs []string) llms.MessageContent
}

// PricingInfo represents pricing information for a model
type PricingInfo struct {
	Input  float64 `json:"input" yaml:"input"`
	Output float64 `json:"output" yaml:"output"`
	Unit   float64 `json:"unit" yaml:"unit"`
}

// BaseLLM provides common functionality for all language models
type BaseLLM struct {
	llms.Model
	features []ModelFeature
	metadata map[string]interface{}
}

// NewBaseLLM creates a new base language model
func NewBaseLLM(model llms.Model, features []ModelFeature, metadata map[string]interface{}) *BaseLLM {
	return &BaseLLM{
		Model:    model,
		features: features,
		metadata: metadata,
	}
}

// GetFeatures returns the features supported by the model
func (b *BaseLLM) GetFeatures() []ModelFeature {
	return b.features
}

// GetMetadata returns the metadata of the model
func (b *BaseLLM) GetMetadata() map[string]interface{} {
	return b.metadata
}

// GetPricing returns the pricing information for the model
func (b *BaseLLM) GetPricing() (float64, float64, float64) {
	pricing, ok := b.metadata["pricing"].(map[string]interface{})
	if !ok {
		return 0.0, 0.0, 0.0
	}

	input, _ := pricing["input"].(float64)
	output, _ := pricing["output"].(float64)
	unit, _ := pricing["unit"].(float64)

	return input, output, unit
}

// ConvertToHumanMessage converts query and image URLs to human messages
func (b *BaseLLM) ConvertToHumanMessage(query string, imageURLs []string) llms.MessageContent {
	// Check if images are provided and the model supports image input
	if len(imageURLs) == 0 || !b.hasFeature(FeatureImageInput) {
		return llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(query)},
		}
	}

	// Create multimodal message with text and images
	var parts []llms.ContentPart
	parts = append(parts, llms.TextPart(query))

	for _, imageURL := range imageURLs {
		parts = append(parts, llms.ImageURLPart(imageURL))
	}

	return llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: parts,
	}
}

// hasFeature checks if the model has a specific feature
func (b *BaseLLM) hasFeature(feature ModelFeature) bool {
	for _, f := range b.features {
		if f == feature {
			return true
		}
	}
	return false
}

// ModelFactory is a function type for creating language models
type ModelFactory func(modelName string, config map[string]interface{}) (BaseLanguageModel, error)

// LanguageModelError represents errors in the language model system
type LanguageModelError struct {
	Message string
	Code    string
}

func (e *LanguageModelError) Error() string {
	return fmt.Sprintf("language model error [%s]: %s", e.Code, e.Message)
}

// NotFoundError creates a not found error
func NotFoundError(message string) *LanguageModelError {
	return &LanguageModelError{
		Message: message,
		Code:    "NOT_FOUND",
	}
}

// InvalidConfigError creates an invalid configuration error
func InvalidConfigError(message string) *LanguageModelError {
	return &LanguageModelError{
		Message: message,
		Code:    "INVALID_CONFIG",
	}
}
