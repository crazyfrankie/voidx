package end

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// EndNodeData 结束节点数据
type EndNodeData struct {
	*entities.BaseNodeData
	Outputs []*entities.VariableEntity `json:"outputs"` // 输出变量列表
}

// NewEndNodeData 创建新的结束节点数据
func NewEndNodeData() *EndNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeEnd

	return &EndNodeData{
		BaseNodeData: baseData,
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetInputs 实现NodeDataInterface接口
func (e *EndNodeData) GetInputs() []*entities.VariableEntity {
	return make([]*entities.VariableEntity, 0) // 结束节点没有输入
}

// GetOutputs 实现NodeDataInterface接口
func (e *EndNodeData) GetOutputs() []*entities.VariableEntity {
	return e.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (e *EndNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return e.BaseNodeData
}
