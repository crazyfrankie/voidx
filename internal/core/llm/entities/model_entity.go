package entities

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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
	FeatureFunctionCall ModelFeature = "function_calling"
	FeatureToolCall     ModelFeature = "tool_calling"
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
	Label string `json:"label" yaml:"label"`
	Value any    `json:"value" yaml:"value"`
}

// ModelParameter represents a parameter configuration for a model
type ModelParameter struct {
	Name      string                 `json:"name" yaml:"name"`
	Label     string                 `json:"label" yaml:"label"`
	Type      ModelParameterType     `json:"type" yaml:"type"`
	Help      string                 `json:"help" yaml:"help"`
	Required  bool                   `json:"required" yaml:"required"`
	Default   any                    `json:"default" yaml:"default"`
	Min       *float64               `json:"min,omitempty" yaml:"min,omitempty"`
	Max       *float64               `json:"max,omitempty" yaml:"max,omitempty"`
	Precision int                    `json:"precision" yaml:"precision"`
	Options   []ModelParameterOption `json:"options" yaml:"options"`
}

// ModelEntity represents a language model configuration
type ModelEntity struct {
	ModelName       string           `json:"model_name" yaml:"model"`
	Label           string           `json:"label" yaml:"label"`
	ModelType       ModelType        `json:"model_type" yaml:"model_type"`
	Features        []ModelFeature   `json:"features" yaml:"features"`
	ContextWindow   int              `json:"context_window" yaml:"context_window"`
	MaxOutputTokens int              `json:"max_output_tokens" yaml:"max_output_tokens"`
	Attributes      map[string]any   `json:"attributes" yaml:"attributes"`
	Parameters      []ModelParameter `json:"parameters" yaml:"parameters"`
	Metadata        map[string]any   `json:"metadata" yaml:"metadata"`
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

// BaseLanguageModel is the interface that wraps eino's ChatModel with additional features
type BaseLanguageModel interface {
	model.BaseChatModel
	GetFeatures() []ModelFeature
	GetMetadata() map[string]any
	GetPricing() (float64, float64, float64) // input_price, output_price, unit
	ConvertToHumanMessage(query string, imageURLs []string) *schema.Message
}

// LLMModel wraps an eino ChatModel to implement BaseLanguageModel
type LLMModel struct {
	model.BaseChatModel
	features []ModelFeature
	metadata map[string]any
}

// NewLLMModel creates a new wrapper for eino ChatModel
func NewLLMModel(chatModel model.BaseChatModel, features []ModelFeature, metadata map[string]any) *LLMModel {
	return &LLMModel{
		BaseChatModel: chatModel,
		features:      features,
		metadata:      metadata,
	}
}

// GetFeatures returns the features supported by the model
func (w *LLMModel) GetFeatures() []ModelFeature {
	return w.features
}

// GetMetadata returns the metadata of the model
func (w *LLMModel) GetMetadata() map[string]any {
	return w.metadata
}

// GetPricing returns the pricing information for the model
func (w *LLMModel) GetPricing() (float64, float64, float64) {
	pricing, ok := w.metadata["pricing"].(map[string]any)
	if !ok {
		return 0.0, 0.0, 0.0
	}

	input, _ := pricing["input"].(float64)
	output, _ := pricing["output"].(float64)
	unit, _ := pricing["unit"].(float64)

	return input, output, unit
}

// ConvertToHumanMessage converts query and image URLs to human messages
func (w *LLMModel) ConvertToHumanMessage(query string, imageURLs []string) *schema.Message {
	// Check if images are provided and the model supports image input
	if len(imageURLs) == 0 || !w.hasFeature(FeatureImageInput) {
		return schema.UserMessage(query)
	}

	// Create multimodal message with text and images
	var parts []schema.ChatMessagePart
	parts = append(parts, schema.ChatMessagePart{
		Type: schema.ChatMessagePartTypeText,
		Text: query,
	})

	for _, imageURL := range imageURLs {
		parts = append(parts, schema.ChatMessagePart{
			Type: schema.ChatMessagePartTypeImageURL,
			ImageURL: &schema.ChatMessageImageURL{
				URL: imageURL,
			},
		})
	}

	return &schema.Message{
		Role:         schema.User,
		MultiContent: parts,
	}
}

// hasFeature checks if the model has a specific feature
func (w *LLMModel) hasFeature(feature ModelFeature) bool {
	for _, f := range w.features {
		if f == feature {
			return true
		}
	}
	return false
}

// ModelFactory is a function type for creating language models
type ModelFactory func(ctx context.Context, modelName string, config map[string]any) (BaseLanguageModel, error)

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

// Provider represents a simplified provider for service layer
type Provider struct {
	Name           string         `json:"name"`
	Position       int            `json:"position"`
	ProviderEntity ProviderEntity `json:"provider_entity"`
	ModelEntities  []*ModelEntity `json:"model_entities"`
}
