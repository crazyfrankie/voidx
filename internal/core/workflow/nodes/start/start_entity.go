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
