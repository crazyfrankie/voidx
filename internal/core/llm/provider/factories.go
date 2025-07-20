package provider

import (
	"fmt"

	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	deepseekpkg "github.com/crazyfrankie/voidx/internal/core/llm/models/deepseek"
	moonshotpkg "github.com/crazyfrankie/voidx/internal/core/llm/models/moonshot"
	ollamapkg "github.com/crazyfrankie/voidx/internal/core/llm/models/ollama"
	openaipkg "github.com/crazyfrankie/voidx/internal/core/llm/models/openai"
	tongyipkg "github.com/crazyfrankie/voidx/internal/core/llm/models/tongyi"
	wenxinpkg "github.com/crazyfrankie/voidx/internal/core/llm/models/wenxin"
)

// GetOpenAIModelFactory returns the model factory for OpenAI
func GetOpenAIModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			// Apply configuration options
			if baseURL, exists := config["base_url"]; exists {
				if url, ok := baseURL.(string); ok {
					options = append(options, openai.WithBaseURL(url))
				}
			}

			chat, err := openaipkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	case entity.ModelTypeCompletion:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			completion, err := openaipkg.NewCompletion(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(completion.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}

// GetMoonshotModelFactory returns the model factory for Moonshot
func GetMoonshotModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			// Apply configuration options
			chat, err := moonshotpkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}

// GetDeepSeekModelFactory returns the model factory for DeepSeek
func GetDeepSeekModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			// Apply configuration options
			chat, err := deepseekpkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}

// GetTongyiModelFactory returns the model factory for Tongyi
func GetTongyiModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			chat, err := tongyipkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}

// GetOllamaModelFactory returns the model factory for Ollama
func GetOllamaModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []ollama.Option{
				ollama.WithModel(modelName),
			}

			// Apply configuration options
			//if temperature, exists := config["temperature"]; exists {
			//	if temp, ok := temperature.(float64); ok {
			//		options = append(options, ollama.WithTemperature(temp))
			//	}
			//}

			chat, err := ollamapkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}

// GetWenxinModelFactory returns the model factory for Wenxin
func GetWenxinModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	switch modelType {
	case entity.ModelTypeChat:
		return func(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
			options := []openai.Option{
				openai.WithModel(modelName),
			}

			chat, err := wenxinpkg.NewChat(options...)
			if err != nil {
				return nil, err
			}

			return entity.NewBaseLLM(chat.Model, []entity.ModelFeature{}, make(map[string]interface{})), nil
		}, nil
	default:
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported model type: %s", modelType))
	}
}
