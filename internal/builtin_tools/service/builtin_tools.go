package service

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/categories"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/entities"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type BuiltinToolsService struct {
	toolManager     *categories.BuiltinCategoryManager
	providerManager *providers.BuiltinProviderManager
}

func NewBuiltinToolsService(toolManager *categories.BuiltinCategoryManager,
	providerManager *providers.BuiltinProviderManager) *BuiltinToolsService {
	return &BuiltinToolsService{
		toolManager:     toolManager,
		providerManager: providerManager,
	}
}

func (s *BuiltinToolsService) GetBuiltinTools(ctx context.Context) []map[string]any {
	pds := s.providerManager.GetProviders()

	var builtinTools []map[string]any
	for _, provider := range pds {
		providerEntity := provider.ProviderEntity

		providerInfo := map[string]any{
			"name":        providerEntity.Name,
			"label":       providerEntity.Label,
			"description": providerEntity.Description,
			"background":  providerEntity.Background,
			"category":    providerEntity.Category,
			"ctime":       providerEntity.Ctime,
		}
		for _, tool := range provider.GetToolEntities() {
			toolInfo := map[string]interface{}{
				"name":        tool.Name,
				"label":       tool.Label,
				"description": tool.Description,
				"inputs":      s.getToolInputs(tool),
			}
			var tools []map[string]any
			tools = append(tools, toolInfo)
			providerInfo["tools"] = tools
		}

		builtinTools = append(builtinTools, providerInfo)
	}

	return builtinTools
}

func (s *BuiltinToolsService) GetProviderTool(ctx context.Context, providerName string, toolName string) (map[string]any, error) {
	provider := s.providerManager.GetProvider(providerName)
	if provider == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(fmt.Sprintf("该提供商 %s 不存在", providerName))
	}

	toolEntity := provider.GetToolEntity(toolName)
	if toolEntity == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(fmt.Sprintf("该工具 %s 不存在", toolName))
	}

	providerEntity := provider.ProviderEntity
	builtinTool := map[string]any{
		"provider": map[string]any{
			"name":        providerEntity.Name,
			"label":       providerEntity.Label,
			"description": providerEntity.Description,
			"background":  providerEntity.Background,
			"category":    providerEntity.Category,
		},
		"name":        toolEntity.Name,
		"label":       toolEntity.Label,
		"description": toolEntity.Description,
		"params":      toolEntity.Params,
		"inputs":      s.getToolInputs(toolEntity),
		"ctime":       providerEntity.Ctime,
	}

	return builtinTool, nil
}

func (s *BuiltinToolsService) GetProviderIcon(ctx context.Context, providerName string) ([]byte, string, error) {
	provider := s.providerManager.GetProvider(providerName)
	if provider == nil {
		return nil, "", errno.ErrNotFound.AppendBizMessage(fmt.Sprintf("该工具提供者 %s 不存在", providerName))
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	providerPath := filepath.Join(
		rootPath,
		"internal", "core", "tools", "builtin_tools", "providers", providerName,
	)

	iconPath := filepath.Join(providerPath, "_asset", provider.ProviderEntity.Icon)

	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return nil, "", errno.ErrNotFound.AppendBizMessage("该工具提供者_asset下未提供图标")
	}

	mimetype := mime.TypeByExtension(filepath.Ext(iconPath))
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}

	byteData, err := os.ReadFile(iconPath)
	if err != nil {
		return nil, "", err
	}

	return byteData, mimetype, nil
}

func (s *BuiltinToolsService) GetCategories(ctx context.Context) []map[string]any {
	categoryMap := s.toolManager.GetCategoryMap()

	res := make([]map[string]any, 0, len(categoryMap))
	for _, category := range categoryMap {
		res = append(res, map[string]any{
			"name":     category.Entity.Name,
			"category": category.Entity.Category,
			"icon":     category.Icon,
		})
	}

	return res
}

func (s *BuiltinToolsService) getToolInputs(tool *entities.ToolEntity) []map[string]any {
	var inputs []map[string]any

	if tool == nil || tool.Params == nil {
		return inputs
	}

	for _, param := range tool.Params {
		input := map[string]any{
			"name":        param.Name,
			"description": param.Label,
			"required":    param.Required,
			"type":        param.Type,
		}

		// 可选字段
		if param.Default != nil {
			input["default"] = param.Default
		}
		if param.Min != nil {
			input["min"] = *param.Min
		}
		if param.Max != nil {
			input["max"] = *param.Max
		}
		if len(param.Options) > 0 {
			input["options"] = param.Options
		}

		inputs = append(inputs, input)
	}

	return inputs
}
