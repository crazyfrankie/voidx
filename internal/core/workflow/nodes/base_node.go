package nodes

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// BaseNode 工作流节点基类接口
type BaseNode interface {
	// Invoke 执行节点逻辑
	Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error)

	// GetNodeData 获取节点数据
	GetNodeData() *entities.BaseNodeData
}

// BaseNodeImpl 基础节点实现
type BaseNodeImpl struct {
	NodeData *entities.BaseNodeData
}

// GetNodeData 获取节点数据
func (b *BaseNodeImpl) GetNodeData() *entities.BaseNodeData {
	return b.NodeData
}

// NewBaseNodeImpl 创建基础节点实现
func NewBaseNodeImpl(nodeData *entities.BaseNodeData) *BaseNodeImpl {
	return &BaseNodeImpl{
		NodeData: nodeData,
	}
}
