package template_transform

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// TemplateTransformNodeData 模板转换节点数据
type TemplateTransformNodeData struct {
	*entities.BaseNodeData
	Template string                     `json:"template"` // 需要拼接转换的字符串模板
	Inputs   []*entities.VariableEntity `json:"inputs"`   // 输入列表信息
	Outputs  []*entities.VariableEntity `json:"outputs"`  // 输出列表信息
}

// NewTemplateTransformNodeData 创建新的模板转换节点数据
func NewTemplateTransformNodeData() *TemplateTransformNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeTemplateTransform

	// 默认输出变量
	outputs := []*entities.VariableEntity{
		{
			Name: "output",
			Type: entities.VariableTypeString,
			Value: entities.VariableValue{
				Type: entities.VariableValueTypeGenerated,
			},
		},
	}

	return &TemplateTransformNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      outputs,
	}
}

// GetInputs 实现NodeDataInterface接口
func (t *TemplateTransformNodeData) GetInputs() []*entities.VariableEntity {
	return t.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (t *TemplateTransformNodeData) GetOutputs() []*entities.VariableEntity {
	return t.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (t *TemplateTransformNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return t.BaseNodeData
}
