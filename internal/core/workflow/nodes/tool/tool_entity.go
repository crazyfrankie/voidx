package tool

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// ToolType 工具类型
type ToolType string

const (
	ToolTypeBuiltin ToolType = "builtin_tool"
	ToolTypeAPI     ToolType = "api_tool"
)

// ToolNodeData 工具节点数据
type ToolNodeData struct {
	*entities.BaseNodeData
	ToolType   ToolType                   `json:"type"`        // 工具类型
	ProviderID string                     `json:"provider_id"` // 工具提供者id
	ToolID     string                     `json:"tool_id"`     // 工具id
	Params     map[string]any             `json:"params"`      // 内置工具设置参数
	Inputs     []*entities.VariableEntity `json:"inputs"`      // 输入变量列表
	Outputs    []*entities.VariableEntity `json:"outputs"`     // 输出字段列表信息
}

// NewToolNodeData 创建新的工具节点数据
func NewToolNodeData() *ToolNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeTool

	// 默认输出变量
	outputs := []*entities.VariableEntity{
		{
			Name: "text",
			Type: entities.VariableTypeString,
			Value: entities.VariableValue{
				Type: entities.VariableValueTypeGenerated,
			},
		},
	}

	return &ToolNodeData{
		BaseNodeData: baseData,
		Params:       make(map[string]any),
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      outputs,
	}
}

// GetInputs 实现NodeDataInterface接口
func (t *ToolNodeData) GetInputs() []*entities.VariableEntity {
	return t.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (t *ToolNodeData) GetOutputs() []*entities.VariableEntity {
	return t.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (t *ToolNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return t.BaseNodeData
}
