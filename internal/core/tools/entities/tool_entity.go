package entities

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// ToolParamType represents the type of a tool parameter
type ToolParamType string

const (
	// ToolParamTypeString represents a string parameter type
	ToolParamTypeString ToolParamType = "string"

	// ToolParamTypeNumber represents a number parameter type
	ToolParamTypeNumber ToolParamType = "number"

	// ToolParamTypeBoolean represents a boolean parameter type
	ToolParamTypeBoolean ToolParamType = "boolean"

	// ToolParamTypeSelect represents a select parameter type
	ToolParamTypeSelect ToolParamType = "select"
)

// ToolParam represents a tool parameter configuration
type ToolParam struct {
	// Name is the actual parameter name
	Name string `json:"name"`

	// Label is the display label for the parameter
	Label string `json:"label"`

	// Type is the parameter type
	Type ToolParamType `json:"type"`

	// Required indicates if the parameter is mandatory
	Required bool `json:"required"`

	// Default is the default value for the parameter
	Default any `json:"default,omitempty"`

	// Min is the minimum value for number parameters
	Min *float64 `json:"min,omitempty"`

	// Max is the maximum value for number parameters
	Max *float64 `json:"max,omitempty"`

	// Options is the list of options for select parameters
	Options []map[string]any `json:"options,omitempty"`
}

// ToolEntity represents a tool configuration
type ToolEntity struct {
	// Name is the tool's name
	Name string `json:"name"`

	// Label is the tool's display label
	Label string `json:"label"`

	// Description is the tool's description
	Description string `json:"description"`

	// Params is the list of tool parameters
	Params []ToolParam `json:"params"`
}

// APIToolEntity API工具实体信息，记录了创建工具所需的配置信息
type APIToolEntity struct {
	ID          string           `json:"id"`          // API工具提供者对应的id
	Name        string           `json:"name"`        // API工具的名称
	URL         string           `json:"url"`         // API工具发起请求的URL地址
	Method      string           `json:"method"`      // API工具发起请求的方法
	Description string           `json:"description"` // API工具的描述信息
	Headers     []entity.Header  `json:"headers"`     // API工具的请求头信息
	Parameters  []map[string]any `json:"parameters"`  // API工具的参数列表信息
}

// BaseTool wraps eino's tool interface for our system
type BaseTool interface {
	tool.InvokableTool
	GetEntity() *ToolEntity
}

// ProviderEntity represents a tool provider configuration
type ProviderEntity struct {
	Name        string `json:"name" yaml:"name"`
	Label       string `json:"label" yaml:"label"`
	Description string `json:"description" yaml:"description"`
	Icon        string `json:"icon" yaml:"icon"`
	Background  string `json:"background" yaml:"background"`
}

// Provider represents a tool provider
type Provider struct {
	Name           string                 `json:"name"`
	Position       int                    `json:"position"`
	ProviderEntity ProviderEntity         `json:"provider_entity"`
	ToolEntityMap  map[string]*ToolEntity `json:"tool_entity_map"`
}

// GetToolEntity returns the tool entity for the specified tool name
func (p *Provider) GetToolEntity(toolName string) (*ToolEntity, error) {
	entity, exists := p.ToolEntityMap[toolName]
	if !exists {
		return nil, &ToolError{
			Message: "该工具实体不存在，请核实后重试",
			Code:    "NOT_FOUND",
		}
	}
	return entity, nil
}

// GetToolEntities returns all tool entities for this provider
func (p *Provider) GetToolEntities() []*ToolEntity {
	entities := make([]*ToolEntity, 0, len(p.ToolEntityMap))
	for _, e := range p.ToolEntityMap {
		entities = append(entities, e)
	}
	return entities
}

// CategoryEntity represents a tool category configuration
type CategoryEntity struct {
	Category string `json:"category" yaml:"category"`
	Name     string `json:"name" yaml:"name"`
	Icon     string `json:"icon" yaml:"icon"`
}

// Validate validates the category entity
func (c *CategoryEntity) Validate() error {
	if c.Category == "" {
		return &ToolError{
			Message: "category is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if c.Name == "" {
		return &ToolError{
			Message: "name is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if c.Icon == "" {
		return &ToolError{
			Message: "icon is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	return nil
}

// ToolProvider represents a tool provider configuration
type ToolProvider struct {
	Name        string `json:"name" yaml:"name"`
	Label       string `json:"label" yaml:"label"`
	Description string `json:"description" yaml:"description"`
	Icon        string `json:"icon" yaml:"icon"`
	Background  string `json:"background" yaml:"background"`
	Category    string `json:"category" yaml:"category"`
}

// ToolInfo represents basic tool information
type ToolInfo struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// ToolError represents errors in the tool system
type ToolError struct {
	Message string
	Code    string
}

func (e *ToolError) Error() string {
	return e.Message
}
