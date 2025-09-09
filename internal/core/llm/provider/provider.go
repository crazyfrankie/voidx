package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/llm/entities"
)

// Provider represents a language model service provider
type Provider struct {
	Name            string                                       `json:"name"`
	Position        int                                          `json:"position"`
	ProviderEntity  entities.ProviderEntity                      `json:"provider_entity"`
	ModelEntityMap  map[string]entities.ModelEntity              `json:"model_entity_map"`
	ModelFactoryMap map[entities.ModelType]entities.ModelFactory `json:"model_factory_map"`
}

// NewProvider creates a new provider instance
func NewProvider(name string, position int, providerEntity entities.ProviderEntity) (*Provider, error) {
	provider := &Provider{
		Name:            name,
		Position:        position,
		ProviderEntity:  providerEntity,
		ModelEntityMap:  make(map[string]entities.ModelEntity),
		ModelFactoryMap: make(map[entities.ModelType]entities.ModelFactory),
	}

	// Initialize model factories for supported model types
	for _, modelType := range providerEntity.SupportedModelTypes {
		factory, err := getModelFactory(name, modelType)
		if err != nil {
			return nil, fmt.Errorf("failed to get model factory for %s/%s: %w", name, modelType, err)
		}
		provider.ModelFactoryMap[modelType] = factory
	}

	// Load model entities from YAML files
	if err := provider.loadModelEntities(); err != nil {
		return nil, fmt.Errorf("failed to load model entities: %w", err)
	}

	return provider, nil
}

// loadModelEntities loads model entities from YAML configuration files
func (p *Provider) loadModelEntities() error {
	// Get the current working directory and construct the provider path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	providerPath := filepath.Join(currentDir, "internal", "core", "llm", "models", p.Name)

	// Read positions.yaml to get the list of model names
	positionsFile := filepath.Join(providerPath, "positions.yaml")
	positionsData, err := os.ReadFile(positionsFile)
	if err != nil {
		return fmt.Errorf("failed to read positions.yaml: %w", err)
	}

	var modelNames []string
	if err := yaml.Unmarshal(positionsData, &modelNames); err != nil {
		return fmt.Errorf("failed to unmarshal positions.yaml: %w", err)
	}

	// Load each model's configuration
	for _, modelName := range modelNames {
		modelFile := filepath.Join(providerPath, fmt.Sprintf("%s.yaml", modelName))
		modelData, err := os.ReadFile(modelFile)
		if err != nil {
			return fmt.Errorf("failed to read model file %s: %w", modelFile, err)
		}

		var modelConfig map[string]any
		if err := yaml.Unmarshal(modelData, &modelConfig); err != nil {
			return fmt.Errorf("failed to unmarshal model file %s: %w", modelFile, err)
		}

		// Process parameters and apply templates
		parameters, err := p.processParameters(modelConfig["parameters"])
		if err != nil {
			return fmt.Errorf("failed to process parameters for model %s: %w", modelName, err)
		}
		modelConfig["parameters"] = parameters

		// Convert to ModelEntity
		modelEntity, err := p.configToModelEntity(modelConfig)
		if err != nil {
			return fmt.Errorf("failed to convert config to model entity for %s: %w", modelName, err)
		}

		p.ModelEntityMap[modelName] = *modelEntity
	}

	return nil
}

// processParameters processes model parameters and applies templates
func (p *Provider) processParameters(parametersInterface any) ([]entities.ModelParameter, error) {
	parametersSlice, ok := parametersInterface.([]any)
	if !ok {
		return nil, fmt.Errorf("parameters must be a slice")
	}

	var parameters []entities.ModelParameter
	for _, paramInterface := range parametersSlice {
		paramMap, ok := paramInterface.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("parameter must be a map")
		}

		// Check if parameter uses a template
		if useTemplate, exists := paramMap["use_template"]; exists {
			templateName := entities.DefaultModelParameterName(useTemplate.(string))
			if template, exists := entities.DefaultModelParameterTemplate[templateName]; exists {
				// Apply template and override with specific values
				param := template
				delete(paramMap, "use_template")

				// Override template values with specific configuration
				if err := p.overrideParameter(&param, paramMap); err != nil {
					return nil, fmt.Errorf("failed to override parameter: %w", err)
				}
				parameters = append(parameters, param)
			} else {
				return nil, fmt.Errorf("unknown template: %s", templateName)
			}
		} else {
			// Convert directly without template
			param, err := p.mapToParameter(paramMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert parameter: %w", err)
			}
			parameters = append(parameters, *param)
		}
	}

	return parameters, nil
}

// overrideParameter overrides template parameter with specific values
func (p *Provider) overrideParameter(param *entities.ModelParameter, overrides map[string]any) error {
	if name, exists := overrides["name"]; exists {
		param.Name = name.(string)
	}
	if label, exists := overrides["label"]; exists {
		param.Label = label.(string)
	}
	if help, exists := overrides["help"]; exists {
		param.Help = help.(string)
	}
	if required, exists := overrides["required"]; exists {
		param.Required = required.(bool)
	}
	if defaultVal, exists := overrides["default"]; exists {
		param.Default = defaultVal
	}
	if m, exists := overrides["min"]; exists {
		if minFloat, ok := m.(float64); ok {
			param.Min = &minFloat
		}
	}
	if m, exists := overrides["max"]; exists {
		if maxFloat, ok := m.(float64); ok {
			param.Max = &maxFloat
		}
	}
	if precision, exists := overrides["precision"]; exists {
		if precisionInt, ok := precision.(int); ok {
			param.Precision = precisionInt
		}
	}
	return nil
}

// mapToParameter converts a map to ModelParameter
func (p *Provider) mapToParameter(paramMap map[string]any) (*entities.ModelParameter, error) {
	param := &entities.ModelParameter{
		Options: []entities.ModelParameterOption{},
	}

	if name, exists := paramMap["name"]; exists {
		param.Name = name.(string)
	}
	if label, exists := paramMap["label"]; exists {
		param.Label = label.(string)
	}
	if paramType, exists := paramMap["type"]; exists {
		param.Type = entities.ModelParameterType(paramType.(string))
	}
	if help, exists := paramMap["help"]; exists {
		param.Help = help.(string)
	}
	if required, exists := paramMap["required"]; exists {
		param.Required = required.(bool)
	}
	if defaultVal, exists := paramMap["default"]; exists {
		param.Default = defaultVal
	}
	if m, exists := paramMap["min"]; exists {
		if minFloat, ok := m.(float64); ok {
			param.Min = &minFloat
		}
	}
	if m, exists := paramMap["max"]; exists {
		if maxFloat, ok := m.(float64); ok {
			param.Max = &maxFloat
		}
	}
	if precision, exists := paramMap["precision"]; exists {
		if precisionInt, ok := precision.(int); ok {
			param.Precision = precisionInt
		}
	}

	return param, nil
}

// configToModelEntity converts configuration map to ModelEntity
func (p *Provider) configToModelEntity(config map[string]any) (*entities.ModelEntity, error) {
	entity := &entities.ModelEntity{
		Attributes: make(map[string]any),
		Metadata:   make(map[string]any),
	}

	if model, exists := config["model"]; exists {
		entity.ModelName = model.(string)
	}
	if label, exists := config["label"]; exists {
		entity.Label = label.(string)
	}
	if modelType, exists := config["model_type"]; exists {
		entity.ModelType = entities.ModelType(modelType.(string))
	}
	if contextWindow, exists := config["context_window"]; exists {
		if cw, ok := contextWindow.(int); ok {
			entity.ContextWindow = cw
		}
	}
	if maxOutputTokens, exists := config["max_output_tokens"]; exists {
		if mot, ok := maxOutputTokens.(int); ok {
			entity.MaxOutputTokens = mot
		}
	}
	if attributes, exists := config["attributes"]; exists {
		if attr, ok := attributes.(map[string]any); ok {
			entity.Attributes = attr
		}
	}
	if metadata, exists := config["metadata"]; exists {
		if meta, ok := metadata.(map[string]any); ok {
			entity.Metadata = meta
		}
	}
	if parameters, exists := config["parameters"]; exists {
		if params, ok := parameters.([]entities.ModelParameter); ok {
			entity.Parameters = params
		}
	}
	if features, exists := config["features"]; exists {
		if featuresSlice, ok := features.([]any); ok {
			for _, feature := range featuresSlice {
				if featureStr, ok := feature.(string); ok {
					entity.Features = append(entity.Features, entities.ModelFeature(featureStr))
				}
			}
		}
	}

	return entity, nil
}

// GetModelFactory returns the model factory for the specified model type
func (p *Provider) GetModelFactory(modelType entities.ModelType) (entities.ModelFactory, error) {
	factory, exists := p.ModelFactoryMap[modelType]
	if !exists {
		return nil, entities.NotFoundError("该模型类不存在，请核实后重试")
	}
	return factory, nil
}

// GetModelEntity returns the model entity for the specified model name
func (p *Provider) GetModelEntity(modelName string) (*entities.ModelEntity, error) {
	entity, exists := p.ModelEntityMap[modelName]
	if !exists {
		return nil, entities.NotFoundError("该模型实体不存在，请核实后重试")
	}
	return &entity, nil
}

// GetModelEntities returns all model entities for this provider
func (p *Provider) GetModelEntities() []*entities.ModelEntity {
	entities := make([]*entities.ModelEntity, 0, len(p.ModelEntityMap))
	for _, e := range p.ModelEntityMap {
		entity := e // Create a copy to avoid pointer issues
		entities = append(entities, &entity)
	}
	return entities
}

// CreateModel creates a language model instance
func (p *Provider) CreateModel(ctx context.Context, modelName string, config map[string]any) (entities.BaseLanguageModel, error) {
	entity, err := p.GetModelEntity(modelName)
	if err != nil {
		return nil, err
	}

	factory, err := p.GetModelFactory(entity.ModelType)
	if err != nil {
		return nil, err
	}

	return factory(ctx, modelName, config)
}

// getModelFactory returns the appropriate model factory based on provider and model type
func getModelFactory(providerName string, modelType entities.ModelType) (entities.ModelFactory, error) {
	switch providerName {
	case "openai":
		return GetOpenAIModelFactory(modelType)
	case "moonshot":
		return GetMoonshotModelFactory(modelType)
	case "deepseek":
		return GetDeepSeekModelFactory(modelType)
	case "tongyi":
		return GetTongyiModelFactory(modelType)
	case "ollama":
		return GetOllamaModelFactory(modelType)
	case "wenxin":
		return GetWenxinModelFactory(modelType)
	default:
		return nil, entities.NotFoundError(fmt.Sprintf("unsupported provider: %s", providerName))
	}
}
