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
