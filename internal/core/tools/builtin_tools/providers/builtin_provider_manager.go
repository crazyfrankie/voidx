package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/cloudwego/eino/components/tool"

	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/dalle"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/duckduckgo"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/gaode"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/google"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/pptx"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/time"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/wikipedia"
)

// ToolParamType represents the type of a tool parameter
type ToolParamType string

const (
	ToolParamTypeString  ToolParamType = "string"
	ToolParamTypeNumber  ToolParamType = "number"
	ToolParamTypeBoolean ToolParamType = "boolean"
	ToolParamTypeSelect  ToolParamType = "select"
)

// ToolParam represents a tool parameter configuration
type ToolParam struct {
	Name     string           `json:"name"`
	Label    string           `json:"label"`
	Type     ToolParamType    `json:"type"`
	Required bool             `json:"required"`
	Default  any              `json:"default,omitempty"`
	Min      *float64         `json:"min,omitempty"`
	Max      *float64         `json:"max,omitempty"`
	Options  []map[string]any `json:"options,omitempty"`
}

// ToolEntity represents a tool configuration
type ToolEntity struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Params      []ToolParam `json:"params"`
}

// ProviderEntity represents a service provider configuration
type ProviderEntity struct {
	Name        string `yaml:"name" json:"name"`
	Label       string `yaml:"label" json:"label"`
	Description string `yaml:"description" json:"description"`
	Icon        string `yaml:"icon" json:"icon"`
	Background  string `yaml:"background" json:"background"`
	Category    string `yaml:"category" json:"category"`
	Ctime       int64  `yaml:"ctime" json:"ctime"`
}

// Provider represents a service provider with its tools
type Provider struct {
	Name           string                 `json:"name"`
	Position       int                    `json:"position"`
	ProviderEntity ProviderEntity         `json:"provider_entity"`
	ToolEntityMap  map[string]*ToolEntity `json:"tool_entity_map"`
	toolMap        map[string]tool.InvokableTool
}

// GetTool returns a tool implementation by name
func (p *Provider) GetTool(toolName string) any {
	return p.toolMap[toolName]
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

// BuiltinProviderManager manages all built-in service providers
type BuiltinProviderManager struct {
	providerMap map[string]*Provider
	toolMap     map[string]tool.InvokableTool
	mu          sync.RWMutex
}

// NewBuiltinProviderManager creates a new BuiltinProviderManager instance
func NewBuiltinProviderManager() (*BuiltinProviderManager, error) {
	manager := &BuiltinProviderManager{
		providerMap: make(map[string]*Provider),
		toolMap:     make(map[string]tool.InvokableTool),
	}

	if err := manager.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return manager, nil
}

// GetProvider returns a provider by name
func (m *BuiltinProviderManager) GetProvider(providerName string) (*Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providerMap[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}
	return provider, nil
}

// GetProviders returns all providers
func (m *BuiltinProviderManager) GetProviders() []*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]*Provider, 0, len(m.providerMap))
	for _, provider := range m.providerMap {
		providers = append(providers, provider)
	}
	return providers
}

// initialize loads all providers and tools
func (m *BuiltinProviderManager) initialize() error {
	// 加载提供商配置
	if err := m.loadProviderConfigs(); err != nil {
		return fmt.Errorf("failed to load provider configs: %w", err)
	}

	// 注册工具
	if err := m.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// 加载工具配置
	if err := m.loadToolConfigs(); err != nil {
		return fmt.Errorf("failed to load tool configs: %w", err)
	}

	return nil
}

// loadProviderConfigs 加载提供商配置
func (m *BuiltinProviderManager) loadProviderConfigs() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, "internal", "core", "tools", "builtin_tools", "providers", "providers.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read providers config: %w", err)
	}

	var configs []ProviderEntity
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal providers config: %w", err)
	}

	for i, config := range configs {
		provider := &Provider{
			Name:           config.Name,
			Position:       i + 1,
			ProviderEntity: config,
			ToolEntityMap:  make(map[string]*ToolEntity),
			toolMap:        make(map[string]tool.InvokableTool),
		}
		m.providerMap[config.Name] = provider
	}

	return nil
}

// registerTools 注册所有工具
func (m *BuiltinProviderManager) registerTools() error {
	var err error

	// 注册时间工具
	m.toolMap["current_time"], err = time.NewCurrentTimeTool()
	if err != nil {
		return fmt.Errorf("failed to create current_time tool: %w", err)
	}

	// 注册搜索工具
	m.toolMap["duckduckgo_search"], err = duckduckgo.NewDuckDuckGoSearchTool()
	if err != nil {
		return fmt.Errorf("failed to create duckduckgo_search tool: %w", err)
	}

	m.toolMap["google_serper"], err = google.NewGoogleSerperTool()
	if err != nil {
		return fmt.Errorf("failed to create google_serper tool: %w", err)
	}

	m.toolMap["wikipedia_search"], err = wikipedia.NewWikipediaSearchTool()
	if err != nil {
		return fmt.Errorf("failed to create wikipedia_search tool: %w", err)
	}

	// 注册图像工具
	m.toolMap["dalle3"], err = dalle.NewDalle3Tool()
	if err != nil {
		return fmt.Errorf("failed to create dalle3 tool: %w", err)
	}

	// 注册天气工具
	m.toolMap["gaode_weather"], err = gaode.NewGaodeWeatherTool()
	if err != nil {
		return fmt.Errorf("failed to create gaode_weather tool: %w", err)
	}

	// 注册PPT工具
	m.toolMap["markdown_to_pptx"], err = pptx.NewMarkdownToPPTXTool()
	if err != nil {
		return fmt.Errorf("failed to create markdown_to_pptx tool: %w", err)
	}

	return nil
}

// loadToolConfigs 加载工具配置
func (m *BuiltinProviderManager) loadToolConfigs() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// 为每个提供商加载工具配置
	toolProviderMap := map[string][]string{
		"time":       {"current_time"},
		"duckduckgo": {"duckduckgo_search"},
		"google":     {"google_serper"},
		"wikipedia":  {"wikipedia_search"},
		"dalle":      {"dalle3"},
		"gaode":      {"gaode_weather"},
		"pptx":       {"markdown_to_pptx"},
	}

	for providerName, toolNames := range toolProviderMap {
		provider, exists := m.providerMap[providerName]
		if !exists {
			continue
		}

		providerPath := filepath.Join(currentDir, "internal", "core", "tools", "builtin_tools", "providers", providerName)

		for _, toolName := range toolNames {
			// 读取工具配置
			toolConfigPath := filepath.Join(providerPath, toolName+".yaml")
			if _, err := os.Stat(toolConfigPath); os.IsNotExist(err) {
				// 如果配置文件不存在，创建默认配置
				toolEntity := &ToolEntity{
					Name:        toolName,
					Label:       toolName,
					Description: fmt.Sprintf("%s tool", toolName),
					Params:      []ToolParam{},
				}
				provider.ToolEntityMap[toolName] = toolEntity
			} else {
				// 读取配置文件
				data, err := os.ReadFile(toolConfigPath)
				if err != nil {
					return fmt.Errorf("failed to read tool config %s: %w", toolName, err)
				}

				var toolEntity ToolEntity
				if err := yaml.Unmarshal(data, &toolEntity); err != nil {
					return fmt.Errorf("failed to unmarshal tool config %s: %w", toolName, err)
				}

				provider.ToolEntityMap[toolName] = &toolEntity
			}

			// 关联工具实现
			if toolImpl, exists := m.toolMap[toolName]; exists {
				provider.toolMap[toolName] = toolImpl
			}
		}
	}

	return nil
}

// GetTool 获取工具
func (m *BuiltinProviderManager) GetTool(name string) (tool.InvokableTool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.toolMap[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return t, nil
}

// ListTools 列出所有工具
func (m *BuiltinProviderManager) ListTools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	for name := range m.toolMap {
		result = append(result, name)
	}
	return result
}

// GetAllTools 获取所有工具的映射
func (m *BuiltinProviderManager) GetAllTools() map[string]tool.InvokableTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]tool.InvokableTool)
	for name, t := range m.toolMap {
		result[name] = t
	}
	return result
}
