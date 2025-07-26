package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/entities"
)

// BuiltinProviderManager manages all built-in service providers
type BuiltinProviderManager struct {
	providerMap map[string]*entities.Provider
	mu          sync.RWMutex
}

// NewBuiltinProviderManager creates a new BuiltinProviderManager instance
func NewBuiltinProviderManager() (*BuiltinProviderManager, error) {
	manager := &BuiltinProviderManager{
		providerMap: make(map[string]*entities.Provider),
	}

	if err := manager.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return manager, nil
}

// GetProvider returns a provider by name
func (m *BuiltinProviderManager) GetProvider(providerName string) *entities.Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providerMap[providerName]
}

// GetProviders returns all providers
func (m *BuiltinProviderManager) GetProviders() []*entities.Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]*entities.Provider, 0, len(m.providerMap))
	for _, provider := range m.providerMap {
		providers = append(providers, provider)
	}
	return providers
}

// GetProviderEntities returns all provider entities
func (m *BuiltinProviderManager) GetProviderEntities() []*entities.ProviderEntity {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entities := make([]*entities.ProviderEntity, 0, len(m.providerMap))
	for _, provider := range m.providerMap {
		entities = append(entities, &provider.ProviderEntity)
	}
	return entities
}

// GetTool returns a specific tool by provider name and tool name
func (m *BuiltinProviderManager) GetTool(providerName, toolName string) interface{} {
	provider := m.GetProvider(providerName)
	if provider == nil {
		return nil
	}
	return provider.GetTool(toolName)
}

// initializeProviders loads all providers from providers.yaml
func (m *BuiltinProviderManager) initializeProviders() error {
	// Get current directory and construct providers.yaml path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	providersPath := filepath.Join(currentDir, "internal", "core", "tools", "builtin_tools", "providers")
	providersYamlPath := filepath.Join(providersPath, "providers.yaml")

	// Read providers.yaml
	data, err := os.ReadFile(providersYamlPath)
	if err != nil {
		return fmt.Errorf("failed to read providers.yaml: %w", err)
	}

	// Parse providers.yaml
	var providerConfigs []entities.ProviderEntity
	if err := yaml.Unmarshal(data, &providerConfigs); err != nil {
		return fmt.Errorf("failed to parse providers.yaml: %w", err)
	}

	// Initialize each provider
	for idx, providerConfig := range providerConfigs {
		provider, err := entities.NewProvider(providerConfig.Name, idx+1, providerConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", providerConfig.Name, err)
		}

		m.providerMap[providerConfig.Name] = provider
	}

	return nil
}
