package categories

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/entities"
)

// CategoryInfo holds category entity and its icon data
type CategoryInfo struct {
	Entity *entities.CategoryEntity
	Icon   string
}

// BuiltinCategoryManager manages built-in tool categories
type BuiltinCategoryManager struct {
	categoryMap map[string]*CategoryInfo
	mutex       sync.RWMutex
}

// NewBuiltinCategoryManager creates a new BuiltinCategoryManager instance
func NewBuiltinCategoryManager() (*BuiltinCategoryManager, error) {
	manager := &BuiltinCategoryManager{
		categoryMap: make(map[string]*CategoryInfo),
	}

	if err := manager.initialize(); err != nil {
		return nil, err
	}

	return manager, nil
}

// GetCategoryMap returns the category mapping
func (m *BuiltinCategoryManager) GetCategoryMap() map[string]*CategoryInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the map to prevent concurrent modification
	copyCate := make(map[string]*CategoryInfo, len(m.categoryMap))
	for k, v := range m.categoryMap {
		copyCate[k] = v
	}

	return copyCate
}

// initialize loads category configurations and icons
func (m *BuiltinCategoryManager) initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.categoryMap) > 0 {
		return nil
	}

	// Get the current directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	categoryPath := filepath.Join(currentDir, "internal", "core", "tools", "builtin_tools", "categories")
	categoryYAMLPath := filepath.Join(categoryPath, "categories.yaml")

	// Read and parse categories.yaml
	categoryData, err := os.ReadFile(categoryYAMLPath)
	if err != nil {
		return fmt.Errorf("failed to read categories.yaml: %w", err)
	}

	var categories []map[string]any
	if err := yaml.Unmarshal(categoryData, &categories); err != nil {
		return fmt.Errorf("failed to parse categories.yaml: %w", err)
	}

	// Process each category
	for _, category := range categories {
		// Create category entity
		categoryData, err := yaml.Marshal(category)
		if err != nil {
			return fmt.Errorf("failed to marshal category data: %w", err)
		}

		var categoryEntity entities.CategoryEntity
		if err := yaml.Unmarshal(categoryData, &categoryEntity); err != nil {
			return fmt.Errorf("failed to unmarshal category entity: %w", err)
		}

		// Validate category entity
		if err := categoryEntity.Validate(); err != nil {
			return fmt.Errorf("invalid category entity: %w", err)
		}

		// Check and read icon file
		iconPath := filepath.Join(categoryPath, "icons", categoryEntity.Icon)
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			return fmt.Errorf("icon not found for category %s: %w", categoryEntity.Category, err)
		}

		iconData, err := os.ReadFile(iconPath)
		if err != nil {
			return fmt.Errorf("failed to read icon file for category %s: %w", categoryEntity.Category, err)
		}

		// Store category information
		m.categoryMap[categoryEntity.Category] = &CategoryInfo{
			Entity: &categoryEntity,
			Icon:   string(iconData),
		}
	}

	return nil
}
