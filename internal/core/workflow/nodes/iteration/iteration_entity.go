package iteration

import (
	"fmt"
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"

	"github.com/google/uuid"
)

// IterationNodeData 迭代节点数据
type IterationNodeData struct {
	*entities.BaseNodeData
	WorkflowIDs []uuid.UUID                `json:"workflow_ids"` // 需要迭代的工作流id
	Inputs      []*entities.VariableEntity `json:"inputs"`       // 输入变量列表
	Outputs     []*entities.VariableEntity `json:"outputs"`      // 输出变量列表
}

// NewIterationNodeData 创建新的迭代节点数据
func NewIterationNodeData() *IterationNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeIteration

	// 默认输入变量
	inputs := []*entities.VariableEntity{
		{
			Name:     "inputs",
			Type:     entities.VariableTypeListString,
			Required: true,
			Value: entities.VariableValue{
				Type:    entities.VariableValueTypeLiteral,
				Content: []any{},
			},
		},
	}

	// 默认输出变量
	outputs := []*entities.VariableEntity{
		{
			Name: "outputs",
			Type: entities.VariableTypeListString,
			Value: entities.VariableValue{
				Type: entities.VariableValueTypeGenerated,
			},
		},
	}

	return &IterationNodeData{
		BaseNodeData: baseData,
		WorkflowIDs:  make([]uuid.UUID, 0),
		Inputs:       inputs,
		Outputs:      outputs,
	}
}

// ValidateWorkflowIDs 校验迭代的工作流数量是否小于等于1
func (i *IterationNodeData) ValidateWorkflowIDs() error {
	if len(i.WorkflowIDs) > 1 {
		return fmt.Errorf("迭代节点只能绑定一个工作流")
	}
	return nil
}

// ValidateInputs 校验输入变量是否正确
func (i *IterationNodeData) ValidateInputs() error {
	// 1. 判断是否一个输入变量，如果不是则抛出错误
	if len(i.Inputs) != 1 {
		return fmt.Errorf("迭代节点输入变量信息错误")
	}

	// 2. 判断输入变量类型及字段是否出错
	iterationInputs := i.Inputs[0]
	allowTypes := []entities.VariableType{
		entities.VariableTypeListString,
		entities.VariableTypeListInt,
		entities.VariableTypeListFloat,
		entities.VariableTypeListBoolean,
	}

	if iterationInputs.Name != "inputs" || !iterationInputs.Required {
		return fmt.Errorf("迭代节点输入变量名字/类型/必填属性出错")
	}

	// 检查类型是否在允许的类型列表中
	typeAllowed := false
	for _, allowType := range allowTypes {
		if iterationInputs.Type == allowType {
			typeAllowed = true
			break
		}
	}

	if !typeAllowed {
		return fmt.Errorf("迭代节点输入变量类型不支持")
	}

	return nil
}
