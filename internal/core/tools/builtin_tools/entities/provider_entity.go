package entities

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProviderEntity represents a service provider configuration
type ProviderEntity struct {
	// Name is the provider's name
	Name string `json:"name"`

	// Label is the provider's display label
	Label string `json:"label"`

	// Description is the provider's description
	Description string `json:"description"`

	// Icon is the provider's icon URL
	Icon string `json:"icon"`

	// Background is the icon's background color
	Background string `json:"background"`

	// Category is the provider's category
	Category string `json:"category"`

	// Ctime is the provider's creation timestamp
	Ctime int64 `json:"ctime"`
}

// ToolFunc represents a tool function signature
type ToolFunc func(args map[string]interface{}) (interface{}, error)

// toolRegistry holds all registered tool functions
var toolRegistry = make(map[string]map[string]ToolFunc)

// RegisterTool registers a tool function for a specific provider
func RegisterTool(providerName, toolName string, toolFunc ToolFunc) {
	if toolRegistry[providerName] == nil {
		toolRegistry[providerName] = make(map[string]ToolFunc)
	}
	toolRegistry[providerName][toolName] = toolFunc
}

// GetRegisteredTool retrieves a registered tool function
func GetRegisteredTool(providerName, toolName string) (ToolFunc, bool) {
	if providerTools, exists := toolRegistry[providerName]; exists {
		if toolFunc, exists := providerTools[toolName]; exists {
			return toolFunc, true
		}
	}
	return nil, false
}

// Provider represents a service provider with its tools
type Provider struct {
	// Name is the provider's name
	Name string `json:"name"`

	// Position is the provider's order
	Position int `json:"position"`

	// ProviderEntity contains the provider's configuration
	ProviderEntity ProviderEntity `json:"provider_entity"`

	// ToolEntityMap maps tool names to their configurations
	ToolEntityMap map[string]*ToolEntity `json:"tool_entity_map"`

	// toolFuncMap maps tool names to their implementations
	toolFuncMap map[string]ToolFunc
}

// NewProvider creates a new Provider instance
func NewProvider(name string, position int, providerEntity ProviderEntity) (*Provider, error) {
	provider := &Provider{
		Name:           name,
		Position:       position,
		ProviderEntity: providerEntity,
		ToolEntityMap:  make(map[string]*ToolEntity),
		toolFuncMap:    make(map[string]ToolFunc),
	}

	if err := provider.initialize(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GetTool returns a tool implementation by name
func (p *Provider) GetTool(toolName string) interface{} {
	return p.toolFuncMap[toolName]
}

// GetToolEntity returns a tool configuration by name
func (p *Provider) GetToolEntity(toolName string) *ToolEntity {
	return p.ToolEntityMap[toolName]
}

// GetToolEntities returns all tool configurations
func (p *Provider) GetToolEntities() []*ToolEntity {
	entities := make([]*ToolEntity, 0, len(p.ToolEntityMap))
	for _, entity := range p.ToolEntityMap {
		entities = append(entities, entity)
	}
	return entities
}

// initialize loads tool configurations and implementations
func (p *Provider) initialize() error {
	// Get the provider's directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	providerPath := filepath.Join(currentDir, "internal", "core", "tools", "builtin_tools", "providers", p.Name)

	// Read positions.yaml
	positionsPath := filepath.Join(providerPath, "positions.yaml")
	positionsData, err := os.ReadFile(positionsPath)
	if err != nil {
		return fmt.Errorf("failed to read positions.yaml: %w", err)
	}

	var positions []string
	if err := yaml.Unmarshal(positionsData, &positions); err != nil {
		return fmt.Errorf("failed to parse positions.yaml: %w", err)
	}

	// Load tool configurations and implementations
	for _, toolName := range positions {
		// Read tool configuration
		toolPath := filepath.Join(providerPath, toolName+".yaml")
		toolData, err := os.ReadFile(toolPath)
		if err != nil {
			return fmt.Errorf("failed to read tool configuration %s: %w", toolName, err)
		}

		var toolEntity ToolEntity
		if err := yaml.Unmarshal(toolData, &toolEntity); err != nil {
			return fmt.Errorf("failed to parse tool configuration %s: %w", toolName, err)
		}

		p.ToolEntityMap[toolName] = &toolEntity

		// Load tool implementation from registry
		if toolFunc, exists := GetRegisteredTool(p.Name, toolName); exists {
			p.toolFuncMap[toolName] = toolFunc
		} else {
			return fmt.Errorf("tool function %s not registered for provider %s", toolName, p.Name)
		}
	}

	return nil
}
