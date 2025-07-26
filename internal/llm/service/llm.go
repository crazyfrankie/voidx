package service

import (
	"context"
	"fmt"

	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
)

type LLMService struct {
	llmCore *llm.LanguageModelManager
}

func NewLLMService(llmCore *llm.LanguageModelManager) *LLMService {
	return &LLMService{
		llmCore: llmCore,
	}
}

// GetProviders 获取所有模型提供商
func (s *LLMService) GetProviders(ctx context.Context) ([]*resp.ProviderResp, error) {
	providers := s.llmCore.GetProviders()

	providerResps := make([]*resp.ProviderResp, 0, len(providers))
	for _, provider := range providers {
		providerResps = append(providerResps, &resp.ProviderResp{
			Name:        provider.Name,
			Label:       provider.ProviderEntity.Label,
			Description: provider.ProviderEntity.Description,
			Icon:        provider.ProviderEntity.Icon,
			Background:  provider.ProviderEntity.Background,
			ModelTypes:  provider.ProviderEntity.SupportedModelTypes,
			Position:    provider.Position,
		})
	}

	return providerResps, nil
}

// GetModels 获取指定提供商的模型列表
func (s *LLMService) GetModels(ctx context.Context, providerName string) ([]*resp.ModelResp, error) {
	provider, err := s.llmCore.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", providerName, err)
	}

	models := provider.GetModelEntities()
	modelResps := make([]*resp.ModelResp, 0, len(models))

	for _, model := range models {
		modelResps = append(modelResps, &resp.ModelResp{
			ModelName:       model.ModelName,
			Label:           model.Label,
			ModelType:       string(model.ModelType),
			Features:        convertFeatures(model.Features),
			ContextWindow:   model.ContextWindow,
			MaxOutputTokens: model.MaxOutputTokens,
			Parameters:      convertParameters(model.Parameters),
			Metadata:        model.Metadata,
		})
	}

	return modelResps, nil
}

// GetModelsByType 根据模型类型获取模型列表
func (s *LLMService) GetModelsByType(ctx context.Context, modelType string) (map[string][]*resp.ModelResp, error) {
	models := s.llmCore.GetModelsByType(entity.ModelType(modelType))

	result := make(map[string][]*resp.ModelResp)
	for providerName, modelList := range models {
		modelResps := make([]*resp.ModelResp, 0, len(modelList))
		for _, model := range modelList {
			modelResps = append(modelResps, &resp.ModelResp{
				ModelName:       model.ModelName,
				Label:           model.Label,
				ModelType:       string(model.ModelType),
				Features:        convertFeatures(model.Features),
				ContextWindow:   model.ContextWindow,
				MaxOutputTokens: model.MaxOutputTokens,
				Parameters:      convertParameters(model.Parameters),
				Metadata:        model.Metadata,
			})
		}
		result[providerName] = modelResps
	}

	return result, nil
}

// GetModelEntity 获取模型实体信息
func (s *LLMService) GetModelEntity(ctx context.Context, provider, modelName string) (*resp.ModelEntityResp, error) {
	en, err := s.llmCore.GetModelEntity(provider, modelName)
	if err != nil {
		return nil, err
	}

	return &resp.ModelEntityResp{
		ModelName:       en.ModelName,
		Label:           en.Label,
		ModelType:       string(en.ModelType),
		Features:        convertFeatures(en.Features),
		ContextWindow:   en.ContextWindow,
		MaxOutputTokens: en.MaxOutputTokens,
		Attributes:      en.Attributes,
		Parameters:      convertParameters(en.Parameters),
		Metadata:        en.Metadata,
	}, nil
}

// ProcessAndValidateModelConfig 处理并验证模型配置
func (s *LLMService) ProcessAndValidateModelConfig(modelConfig map[string]any) (map[string]any, error) {
	// 检查模型配置格式
	if modelConfig == nil {
		return s.getDefaultModelConfig(), nil
	}

	// 提取配置信息
	providerName, ok := modelConfig["provider"].(string)
	if !ok || providerName == "" {
		return s.getDefaultModelConfig(), nil
	}

	modelName, ok := modelConfig["model"].(string)
	if !ok || modelName == "" {
		return s.getDefaultModelConfig(), nil
	}

	// 验证提供商是否存在
	provider, err := s.llmCore.GetProvider(providerName)
	if err != nil {
		return s.getDefaultModelConfig(), nil
	}

	// 验证模型是否存在
	modelEntity, err := provider.GetModelEntity(modelName)
	if err != nil {
		return s.getDefaultModelConfig(), nil
	}

	// 处理参数
	parameters := make(map[string]any)
	configParams, ok := modelConfig["parameters"].(map[string]any)
	if !ok {
		configParams = make(map[string]any)
	}

	// 为每个参数设置值
	for _, param := range modelEntity.Parameters {
		value, exists := configParams[param.Name]
		if !exists {
			value = param.Default
		}

		// 验证参数值
		if param.Required && value == nil {
			value = param.Default
		}

		// 类型验证
		if value != nil && !s.validateParameterType(value, param.Type) {
			value = param.Default
		}

		// 范围验证
		if param.Min != nil || param.Max != nil {
			if floatVal, ok := value.(float64); ok {
				if param.Min != nil && floatVal < *param.Min {
					value = param.Default
				}
				if param.Max != nil && floatVal > *param.Max {
					value = param.Default
				}
			}
		}

		parameters[param.Name] = value
	}

	return map[string]any{
		"provider":   providerName,
		"model":      modelName,
		"parameters": parameters,
	}, nil
}

// LoadLanguageModel 从模型配置加载语言模型
func (s *LLMService) LoadLanguageModel(modelConfig map[string]any) (entity.BaseLanguageModel, error) {
	// 验证并处理模型配置
	validConfig, err := s.ProcessAndValidateModelConfig(modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to validate model config: %w", err)
	}

	providerName := validConfig["provider"].(string)
	modelName := validConfig["model"].(string)
	parameters := validConfig["parameters"].(map[string]any)

	// 创建模型实例
	return s.llmCore.CreateModel(providerName, modelName, parameters)
}

// validateParameterType 验证参数类型
func (s *LLMService) validateParameterType(value interface{}, paramType entity.ModelParameterType) bool {
	switch paramType {
	case entity.ParameterTypeFloat:
		_, ok := value.(float64)
		return ok
	case entity.ParameterTypeInt:
		_, ok := value.(int)
		if !ok {
			// 也接受 float64 类型的整数
			if f, ok := value.(float64); ok {
				return f == float64(int(f))
			}
		}
		return ok
	case entity.ParameterTypeString:
		_, ok := value.(string)
		return ok
	case entity.ParameterTypeBoolean:
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}

// getDefaultModelConfig 获取默认模型配置
func (s *LLMService) getDefaultModelConfig() map[string]any {
	return map[string]any{
		"provider": "openai",
		"model":    "gpt-4o-mini",
		"parameters": map[string]any{
			"temperature": 0.7,
			"max_tokens":  1000,
		},
	}
}

func convertFeatures(features []entity.ModelFeature) []string {
	result := make([]string, 0, len(features))
	for _, feature := range features {
		result = append(result, string(feature))
	}
	return result
}

func convertParameters(parameters []entity.ModelParameter) []resp.ModelParameterResp {
	result := make([]resp.ModelParameterResp, 0, len(parameters))
	for _, param := range parameters {
		result = append(result, resp.ModelParameterResp{
			Name:      param.Name,
			Label:     param.Label,
			Type:      string(param.Type),
			Help:      param.Help,
			Required:  param.Required,
			Default:   param.Default,
			Min:       param.Min,
			Max:       param.Max,
			Precision: param.Precision,
			Options:   convertParameterOptions(param.Options),
		})
	}
	return result
}

func convertParameterOptions(options []entity.ModelParameterOption) []resp.ModelParameterOptionResp {
	result := make([]resp.ModelParameterOptionResp, 0, len(options))
	for _, option := range options {
		result = append(result, resp.ModelParameterOptionResp{
			Label: option.Label,
			Value: option.Value,
		})
	}
	return result
}
