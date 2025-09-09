package provider

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"

	"github.com/crazyfrankie/voidx/internal/core/llm/entities"
)

// GetOpenAIModelFactory returns the OpenAI model factory
func GetOpenAIModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("OpenAI only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// Extract configuration parameters
		apiKey, _ := config["api_key"].(string)
		if apiKey == "" {
			return nil, entities.InvalidConfigError("api_key is required for OpenAI models")
		}

		// Build OpenAI configuration
		openaiConfig := &openai.ChatModelConfig{
			APIKey: apiKey,
			Model:  modelName,
		}

		// Apply optional parameters
		if temperature, ok := config["temperature"].(float64); ok {
			temp := float32(temperature)
			openaiConfig.Temperature = &temp
		}
		if topP, ok := config["top_p"].(float64); ok {
			tp := float32(topP)
			openaiConfig.TopP = &tp
		}
		if maxTokens, ok := config["max_tokens"].(int); ok {
			openaiConfig.MaxTokens = &maxTokens
		}
		if presencePenalty, ok := config["presence_penalty"].(float64); ok {
			pp := float32(presencePenalty)
			openaiConfig.PresencePenalty = &pp
		}
		if frequencyPenalty, ok := config["frequency_penalty"].(float64); ok {
			fp := float32(frequencyPenalty)
			openaiConfig.FrequencyPenalty = &fp
		}

		// Create eino OpenAI chat model
		chatModel, err := openai.NewChatModel(ctx, openaiConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
		}

		// Get features and metadata from config
		features := extractFeatures(config)
		metadata := extractMetadata(config)

		// Wrap with our interface
		return entities.NewLLMModel(chatModel, features, metadata), nil
	}, nil
}

// GetDeepSeekModelFactory returns the DeepSeek model factory
func GetDeepSeekModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("DeepSeek only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// Extract configuration parameters
		apiKey, _ := config["api_key"].(string)
		if apiKey == "" {
			return nil, entities.InvalidConfigError("api_key is required for DeepSeek models")
		}

		// Build DeepSeek configuration
		deepseekConfig := &deepseek.ChatModelConfig{
			APIKey: apiKey,
			Model:  modelName,
		}

		// Apply optional parameters
		if temperature, ok := config["temperature"].(float64); ok {
			temp := float32(temperature)
			deepseekConfig.Temperature = temp
		}
		if topP, ok := config["top_p"].(float64); ok {
			tp := float32(topP)
			deepseekConfig.TopP = tp
		}
		if maxTokens, ok := config["max_tokens"].(int); ok {
			deepseekConfig.MaxTokens = maxTokens
		}

		// Create eino DeepSeek chat model
		chatModel, err := deepseek.NewChatModel(ctx, deepseekConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create DeepSeek chat model: %w", err)
		}

		// Get features and metadata from config
		features := extractFeatures(config)
		metadata := extractMetadata(config)

		// Wrap with our interface
		return entities.NewLLMModel(chatModel, features, metadata), nil
	}, nil
}

// GetTongyiModelFactory returns the Tongyi (Qwen) model factory
func GetTongyiModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("Tongyi only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// Extract configuration parameters
		apiKey, _ := config["api_key"].(string)
		if apiKey == "" {
			return nil, entities.InvalidConfigError("api_key is required for Tongyi models")
		}

		// Build Qwen configuration
		qwenConfig := &qwen.ChatModelConfig{
			APIKey: apiKey,
			Model:  modelName,
		}

		// Apply optional parameters
		if temperature, ok := config["temperature"].(float64); ok {
			temp := float32(temperature)
			qwenConfig.Temperature = &temp
		}
		if topP, ok := config["top_p"].(float64); ok {
			tp := float32(topP)
			qwenConfig.TopP = &tp
		}
		if maxTokens, ok := config["max_tokens"].(int); ok {
			qwenConfig.MaxTokens = &maxTokens
		}

		// Create eino Qwen chat model
		chatModel, err := qwen.NewChatModel(ctx, qwenConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Qwen chat model: %w", err)
		}

		// Get features and metadata from config
		features := extractFeatures(config)
		metadata := extractMetadata(config)

		// Wrap with our interface
		return entities.NewLLMModel(chatModel, features, metadata), nil
	}, nil
}

// GetOllamaModelFactory returns the Ollama model factory
func GetOllamaModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("Ollama only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// Build Ollama configuration
		ollamaConfig := &ollama.ChatModelConfig{
			Model: modelName,
		}

		// Apply optional parameters
		if baseURL, ok := config["base_url"].(string); ok {
			ollamaConfig.BaseURL = baseURL
		}
		if temperature, ok := config["temperature"].(float64); ok {
			temp := float32(temperature)
			ollamaConfig.Options.Temperature = temp
		}
		if topP, ok := config["top_p"].(float64); ok {
			tp := float32(topP)
			ollamaConfig.Options.TopP = tp
		}

		// Create eino Ollama chat model
		chatModel, err := ollama.NewChatModel(ctx, ollamaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama chat model: %w", err)
		}

		// Get features and metadata from config
		features := extractFeatures(config)
		metadata := extractMetadata(config)

		// Wrap with our interface
		return entities.NewLLMModel(chatModel, features, metadata), nil
	}, nil
}

// GetMoonshotModelFactory returns the Moonshot model factory
func GetMoonshotModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("Moonshot only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// Moonshot uses OpenAI-compatible API, so we use OpenAI client with custom base URL
		apiKey, _ := config["api_key"].(string)
		if apiKey == "" {
			return nil, entities.InvalidConfigError("api_key is required for Moonshot models")
		}

		// Build OpenAI configuration with Moonshot base URL
		openaiConfig := &openai.ChatModelConfig{
			APIKey:  apiKey,
			Model:   modelName,
			BaseURL: "https://api.moonshot.cn/v1",
		}

		// Apply optional parameters
		if temperature, ok := config["temperature"].(float64); ok {
			temp := float32(temperature)
			openaiConfig.Temperature = &temp
		}
		if topP, ok := config["top_p"].(float64); ok {
			tp := float32(topP)
			openaiConfig.TopP = &tp
		}
		if maxTokens, ok := config["max_tokens"].(int); ok {
			openaiConfig.MaxTokens = &maxTokens
		}

		// Create eino OpenAI chat model (compatible with Moonshot)
		chatModel, err := openai.NewChatModel(ctx, openaiConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Moonshot chat model: %w", err)
		}

		// Get features and metadata from config
		features := extractFeatures(config)
		metadata := extractMetadata(config)

		// Wrap with our interface
		return entities.NewLLMModel(chatModel, features, metadata), nil
	}, nil
}

// GetWenxinModelFactory returns the Wenxin model factory
func GetWenxinModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	if modelType != entities.ModelTypeChat {
		return nil, entities.InvalidConfigError("Wenxin only supports chat models")
	}

	return func(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
		// For now, return an error as Wenxin is not implemented in eino-ext yet
		return nil, entities.InvalidConfigError("Wenxin provider is not yet implemented with eino")
	}, nil
}

// extractFeatures extracts model features from configuration
func extractFeatures(config map[string]any) []entities.ModelFeature {
	var features []entities.ModelFeature

	if featuresInterface, exists := config["features"]; exists {
		if featuresSlice, ok := featuresInterface.([]any); ok {
			for _, feature := range featuresSlice {
				if featureStr, ok := feature.(string); ok {
					features = append(features, entities.ModelFeature(featureStr))
				}
			}
		}
	}

	return features
}

// extractMetadata extracts model metadata from configuration
func extractMetadata(config map[string]any) map[string]any {
	metadata := make(map[string]any)

	if metadataInterface, exists := config["metadata"]; exists {
		if metadataMap, ok := metadataInterface.(map[string]any); ok {
			metadata = metadataMap
		}
	}

	return metadata
}
