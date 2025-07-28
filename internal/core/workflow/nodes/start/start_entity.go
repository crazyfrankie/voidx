package start

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// StartNodeData 开始节点数据
type StartNodeData struct {
	*entities.BaseNodeData
	Inputs []*entities.VariableEntity `json:"inputs"` // 输入变量列表
}

// NewStartNodeData 创建新的开始节点数据
func NewStartNodeData() *StartNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeStart

	return &StartNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
	}
}

// GetInputs 实现NodeDataInterface接口
func (s *StartNodeData) GetInputs() []*entities.VariableEntity {
	return s.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (s *StartNodeData) GetOutputs() []*entities.VariableEntity {
	return make([]*entities.VariableEntity, 0) // 开始节点没有输出
}

// GetBaseNodeData 实现NodeDataInterface接口
func (s *StartNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return s.BaseNodeData
}
