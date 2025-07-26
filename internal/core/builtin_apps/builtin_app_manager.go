package builtin_apps

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/crazyfrankie/voidx/internal/core/builtin_apps/entities"
)

// BuiltinAppManager manages built-in applications and their categories
type BuiltinAppManager struct {
	builtinAppMap map[string]*entities.BuiltinAppEntity
	categories    []*entities.CategoryEntity
	mutex         sync.RWMutex
}

// NewBuiltinAppManager creates a new instance of BuiltinAppManager
func NewBuiltinAppManager() *BuiltinAppManager {
	manager := &BuiltinAppManager{
		builtinAppMap: make(map[string]*entities.BuiltinAppEntity),
		categories:    make([]*entities.CategoryEntity, 0),
	}

	manager.initCategories()
	manager.initBuiltinAppMap()

	return manager
}

// GetBuiltinApp returns a built-in application by its ID
func (m *BuiltinAppManager) GetBuiltinApp(builtinAppID string) *entities.BuiltinAppEntity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.builtinAppMap[builtinAppID]
}

// GetBuiltinApps returns all built-in applications
func (m *BuiltinAppManager) GetBuiltinApps() []*entities.BuiltinAppEntity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	apps := make([]*entities.BuiltinAppEntity, 0, len(m.builtinAppMap))
	for _, app := range m.builtinAppMap {
		apps = append(apps, app)
	}
	return apps
}

// GetCategories returns all categories
func (m *BuiltinAppManager) GetCategories() []*entities.CategoryEntity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.categories
}

// initBuiltinAppMap initializes the built-in application map from YAML files
func (m *BuiltinAppManager) initBuiltinAppMap() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.builtinAppMap) > 0 {
		return
	}

	// Get the current file's directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	builtinAppsPath := filepath.Join(currentDir, "internal", "core", "builtin_apps", "builtin_apps")

	// Read all YAML files in the builtin_apps directory
	entries, err := os.ReadDir(builtinAppsPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml") {
			filePath := filepath.Join(builtinAppsPath, entry.Name())

			// Read and parse YAML file
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var appConfig map[string]any
			if err := yaml.Unmarshal(data, &appConfig); err != nil {
				continue
			}

			// Convert model_config to language_model_config
			if modelConfig, ok := appConfig["model_config"]; ok {
				appConfig["language_model_config"] = modelConfig
				delete(appConfig, "model_config")
			}

			// Create BuiltinAppEntity
			app := &entities.BuiltinAppEntity{}
			if err := yaml.Unmarshal(data, app); err != nil {
				continue
			}

			m.builtinAppMap[app.ID] = app
		}
	}
}

// initCategories initializes categories from YAML file
func (m *BuiltinAppManager) initCategories() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.categories) > 0 {
		return
	}

	// Get the current file's directory path
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}

	categoriesPath := filepath.Join(currentDir, "internal", "core", "builtin_apps", "categories", "categories.yaml")

	// Read and parse categories YAML file
	data, err := os.ReadFile(categoriesPath)
	if err != nil {
		return
	}

	var categories []map[string]any
	if err := yaml.Unmarshal(data, &categories); err != nil {
		return
	}

	// Create CategoryEntity instances
	for _, category := range categories {
		categoryEntity := &entities.CategoryEntity{}
		categoryData, err := yaml.Marshal(category)
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(categoryData, categoryEntity); err != nil {
			continue
		}

		m.categories = append(m.categories, categoryEntity)
	}
}
