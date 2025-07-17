package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/llm/entity"
)

// Provider represents a language model service provider
type Provider struct {
	Name            string                                   `json:"name"`
	Position        int                                      `json:"position"`
	ProviderEntity  entity.ProviderEntity                    `json:"provider_entity"`
	ModelEntityMap  map[string]entity.ModelEntity            `json:"model_entity_map"`
	ModelFactoryMap map[entity.ModelType]entity.ModelFactory `json:"model_factory_map"`
}

// NewProvider creates a new provider instance
func NewProvider(name string, position int, providerEntity entity.ProviderEntity) (*Provider, error) {
	provider := &Provider{
		Name:            name,
		Position:        position,
		ProviderEntity:  providerEntity,
		ModelEntityMap:  make(map[string]entity.ModelEntity),
		ModelFactoryMap: make(map[entity.ModelType]entity.ModelFactory),
	}

	// Initialize model factories for supported model types
	for _, modelType := range providerEntity.SupportedModelTypes {
		factory, err := getModelFactory(name, modelType)
		if err != nil {
			return nil, fmt.Errorf("failed to get model factory for %s/%s: %w", name, modelType, err)
		}
		provider.ModelFactoryMap[modelType] = factory
	}

	// Load model entity from YAML files
	if err := provider.loadModelEntities(); err != nil {
		return nil, fmt.Errorf("failed to load model entity: %w", err)
	}

	return provider, nil
}

// loadModelEntities loads model entity from YAML configuration files
func (p *Provider) loadModelEntities() error {
	// Get the current working directory and construct the provider path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	providerPath := filepath.Join(currentDir, "internal", "core", "llm", "providers", p.Name)

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

		var modelConfig map[string]interface{}
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
func (p *Provider) processParameters(parametersInterface interface{}) ([]entity.ModelParameter, error) {
	parametersSlice, ok := parametersInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parameters must be a slice")
	}

	var parameters []entity.ModelParameter
	for _, paramInterface := range parametersSlice {
		paramMap, ok := paramInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("parameter must be a map")
		}

		// Check if parameter uses a template
		if useTemplate, exists := paramMap["use_template"]; exists {
			templateName := entity.DefaultModelParameterName(useTemplate.(string))
			if template, exists := entity.DefaultModelParameterTemplate[templateName]; exists {
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
func (p *Provider) overrideParameter(param *entity.ModelParameter, overrides map[string]interface{}) error {
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
func (p *Provider) mapToParameter(paramMap map[string]interface{}) (*entity.ModelParameter, error) {
	param := &entity.ModelParameter{
		Options: []entity.ModelParameterOption{},
	}

	if name, exists := paramMap["name"]; exists {
		param.Name = name.(string)
	}
	if label, exists := paramMap["label"]; exists {
		param.Label = label.(string)
	}
	if paramType, exists := paramMap["type"]; exists {
		param.Type = entity.ModelParameterType(paramType.(string))
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
func (p *Provider) configToModelEntity(config map[string]interface{}) (*entity.ModelEntity, error) {
	entities := &entity.ModelEntity{
		Attributes: make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}

	if model, exists := config["model"]; exists {
		entities.ModelName = model.(string)
	}
	if label, exists := config["label"]; exists {
		entities.Label = label.(string)
	}
	if modelType, exists := config["model_type"]; exists {
		entities.ModelType = entity.ModelType(modelType.(string))
	}
	if contextWindow, exists := config["context_window"]; exists {
		if cw, ok := contextWindow.(int); ok {
			entities.ContextWindow = cw
		}
	}
	if maxOutputTokens, exists := config["max_output_tokens"]; exists {
		if mot, ok := maxOutputTokens.(int); ok {
			entities.MaxOutputTokens = mot
		}
	}
	if attributes, exists := config["attributes"]; exists {
		if attr, ok := attributes.(map[string]interface{}); ok {
			entities.Attributes = attr
		}
	}
	if metadata, exists := config["metadata"]; exists {
		if meta, ok := metadata.(map[string]interface{}); ok {
			entities.Metadata = meta
		}
	}
	if parameters, exists := config["parameters"]; exists {
		if params, ok := parameters.([]entity.ModelParameter); ok {
			entities.Parameters = params
		}
	}
	if features, exists := config["features"]; exists {
		if featuresSlice, ok := features.([]interface{}); ok {
			for _, feature := range featuresSlice {
				if featureStr, ok := feature.(string); ok {
					entities.Features = append(entities.Features, entity.ModelFeature(featureStr))
				}
			}
		}
	}

	return entities, nil
}

// GetModelFactory returns the model factory for the specified model type
func (p *Provider) GetModelFactory(modelType entity.ModelType) (entity.ModelFactory, error) {
	factory, exists := p.ModelFactoryMap[modelType]
	if !exists {
		return nil, entity.NotFoundError("该模型类不存在，请核实后重试")
	}
	return factory, nil
}

// GetModelEntity returns the model entity for the specified model name
func (p *Provider) GetModelEntity(modelName string) (*entity.ModelEntity, error) {
	entities, exists := p.ModelEntityMap[modelName]
	if !exists {
		return nil, entity.NotFoundError("该模型实体不存在，请核实后重试")
	}
	return &entities, nil
}

// GetModelEntities returns all model entity for this provider
func (p *Provider) GetModelEntities() []entity.ModelEntity {
	entities := make([]entity.ModelEntity, 0, len(p.ModelEntityMap))
	for _, e := range p.ModelEntityMap {
		entities = append(entities, e)
	}
	return entities
}

// CreateModel creates a language model instance
func (p *Provider) CreateModel(modelName string, config map[string]interface{}) (entity.BaseLanguageModel, error) {
	entities, err := p.GetModelEntity(modelName)
	if err != nil {
		return nil, err
	}

	factory, err := p.GetModelFactory(entities.ModelType)
	if err != nil {
		return nil, err
	}

	return factory(modelName, config)
}

// getModelFactory returns the appropriate model factory based on provider and model type
func getModelFactory(providerName string, modelType entity.ModelType) (entity.ModelFactory, error) {
	switch strings.ToLower(providerName) {
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
		return nil, entity.NotFoundError(fmt.Sprintf("unsupported provider: %s", providerName))
	}
}
