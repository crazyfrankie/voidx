package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
)

// LanguageModelManager manages all language model providers and their models
type LanguageModelManager struct {
	providerMap map[string]*Provider
	mu          sync.RWMutex
}

// NewLanguageModelManager creates a new language model manager instance
func NewLanguageModelManager() (*LanguageModelManager, error) {
	manager := &LanguageModelManager{
		providerMap: make(map[string]*Provider),
	}

	if err := manager.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize language model manager: %w", err)
	}

	return manager, nil
}

// initialize loads all providers from the providers.yaml configuration
func (lmm *LanguageModelManager) initialize() error {
	// Get the current working directory and construct the providers path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	providersPath := filepath.Join(currentDir, "internal", "core", "llm", "providers")
	providersYamlPath := filepath.Join(providersPath, "providers.yaml")

	// Read providers.yaml
	providersData, err := os.ReadFile(providersYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read providers.yaml: %w", err)
	}

	var providersConfig []entity.ProviderEntity
	if err := yaml.Unmarshal(providersData, &providersConfig); err != nil {
		return fmt.Errorf("failed to unmarshal providers.yaml: %w", err)
	}

	// Initialize each provider
	for i, providerEntity := range providersConfig {
		provider, err := NewProvider(providerEntity.Name, i+1, providerEntity)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", providerEntity.Name, err)
		}

		lmm.providerMap[providerEntity.Name] = provider
	}

	return nil
}

// GetProvider returns a provider by name
func (lmm *LanguageModelManager) GetProvider(providerName string) (*Provider, error) {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	provider, exists := lmm.providerMap[providerName]
	if !exists {
		return nil, entity.NotFoundError("该模型服务提供商不存在，请核实后重试")
	}

	return provider, nil
}

// GetProviders returns all available providers
func (lmm *LanguageModelManager) GetProviders() []*Provider {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	providers := make([]*Provider, 0, len(lmm.providerMap))
	for _, provider := range lmm.providerMap {
		providers = append(providers, provider)
	}

	return providers
}

// GetModelFactoryByProviderAndType returns a model factory by provider name and model type
func (lmm *LanguageModelManager) GetModelFactoryByProviderAndType(providerName string, modelType entity.ModelType) (entity.ModelFactory, error) {
	provider, err := lmm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetModelFactory(modelType)
}

// GetModelFactoryByProviderAndModel returns a model factory by provider name and model name
func (lmm *LanguageModelManager) GetModelFactoryByProviderAndModel(providerName string, modelName string) (entity.ModelFactory, error) {
	// Get the provider
	provider, err := lmm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Get the model entity to determine its type
	modelEntity, err := provider.GetModelEntity(modelName)
	if err != nil {
		return nil, err
	}

	// Get the model factory for the model type
	return provider.GetModelFactory(modelEntity.ModelType)
}

// CreateModel creates a language model instance
func (lmm *LanguageModelManager) CreateModel(providerName string, modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
	provider, err := lmm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.CreateModel(modelName, config)
}

// GetModelEntity returns a model entity by provider and model name
func (lmm *LanguageModelManager) GetModelEntity(providerName string, modelName string) (*entity.ModelEntity, error) {
	provider, err := lmm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetModelEntity(modelName)
}

// GetAllModels returns all available models from all providers
func (lmm *LanguageModelManager) GetAllModels() map[string][]entity.ModelEntity {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	allModels := make(map[string][]entity.ModelEntity)
	for providerName, provider := range lmm.providerMap {
		allModels[providerName] = provider.GetModelEntities()
	}

	return allModels
}

// GetModelsByProvider returns all models for a specific provider
func (lmm *LanguageModelManager) GetModelsByProvider(providerName string) ([]entity.ModelEntity, error) {
	provider, err := lmm.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetModelEntities(), nil
}

// GetModelsByType returns all models of a specific type from all providers
func (lmm *LanguageModelManager) GetModelsByType(modelType entity.ModelType) map[string][]entity.ModelEntity {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	modelsByType := make(map[string][]entity.ModelEntity)
	for providerName, provider := range lmm.providerMap {
		var modelsOfType []entity.ModelEntity
		for _, model := range provider.GetModelEntities() {
			if model.ModelType == modelType {
				modelsOfType = append(modelsOfType, model)
			}
		}
		if len(modelsOfType) > 0 {
			modelsByType[providerName] = modelsOfType
		}
	}

	return modelsByType
}

// GetModelsByFeature returns all models that support a specific feature
func (lmm *LanguageModelManager) GetModelsByFeature(feature entity.ModelFeature) map[string][]entity.ModelEntity {
	lmm.mu.RLock()
	defer lmm.mu.RUnlock()

	modelsByFeature := make(map[string][]entity.ModelEntity)
	for providerName, provider := range lmm.providerMap {
		var modelsWithFeature []entity.ModelEntity
		for _, model := range provider.GetModelEntities() {
			for _, modelFeature := range model.Features {
				if modelFeature == feature {
					modelsWithFeature = append(modelsWithFeature, model)
					break
				}
			}
		}
		if len(modelsWithFeature) > 0 {
			modelsByFeature[providerName] = modelsWithFeature
		}
	}

	return modelsByFeature
}

// ValidateModelConfig validates model configuration parameters
func (lmm *LanguageModelManager) ValidateModelConfig(providerName string, modelName string, config map[string]interface{}) error {
	modelEntity, err := lmm.GetModelEntity(providerName, modelName)
	if err != nil {
		return err
	}

	// Validate each parameter in the configuration
	for configKey, configValue := range config {
		var paramFound bool
		for _, param := range modelEntity.Parameters {
			if param.Name == configKey {
				paramFound = true
				if err := lmm.validateParameter(param, configValue); err != nil {
					return fmt.Errorf("invalid parameter %s: %w", configKey, err)
				}
				break
			}
		}
		if !paramFound {
			return fmt.Errorf("unknown parameter: %s", configKey)
		}
	}

	// Check required parameters
	for _, param := range modelEntity.Parameters {
		if param.Required {
			if _, exists := config[param.Name]; !exists {
				return fmt.Errorf("required parameter missing: %s", param.Name)
			}
		}
	}

	return nil
}

// validateParameter validates a single parameter value
func (lmm *LanguageModelManager) validateParameter(param entity.ModelParameter, value interface{}) error {
	switch param.Type {
	case entity.ParameterTypeFloat:
		if floatVal, ok := value.(float64); ok {
			if param.Min != nil && floatVal < *param.Min {
				return fmt.Errorf("value %f is below minimum %f", floatVal, *param.Min)
			}
			if param.Max != nil && floatVal > *param.Max {
				return fmt.Errorf("value %f is above maximum %f", floatVal, *param.Max)
			}
		} else {
			return fmt.Errorf("expected float, got %T", value)
		}
	case entity.ParameterTypeInt:
		if intVal, ok := value.(int); ok {
			floatVal := float64(intVal)
			if param.Min != nil && floatVal < *param.Min {
				return fmt.Errorf("value %d is below minimum %f", intVal, *param.Min)
			}
			if param.Max != nil && floatVal > *param.Max {
				return fmt.Errorf("value %d is above maximum %f", intVal, *param.Max)
			}
		} else {
			return fmt.Errorf("expected int, got %T", value)
		}
	case entity.ParameterTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case entity.ParameterTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	}

	return nil
}

// Reload reloads all provider configurations
func (lmm *LanguageModelManager) Reload() error {
	lmm.mu.Lock()
	defer lmm.mu.Unlock()

	// Clear existing providers
	lmm.providerMap = make(map[string]*Provider)

	// Reinitialize
	return lmm.initialize()
}
