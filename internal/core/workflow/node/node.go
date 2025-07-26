package node

import (
	"context"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// Node 节点接口
type Node interface {
	// GetNodeData 获取节点数据
	GetNodeData() *entity.BaseNodeData

	// Invoke 执行节点
	Invoke(ctx context.Context, state map[string]any) (map[string]any, error)

	// GetNextNode 获取下一个节点
	GetNextNode(state map[string]any) string

	// Validate 验证节点配置
	Validate() error
}

// BaseNode 基础节点实现
type BaseNode struct {
	Data *entity.BaseNodeData
}

// GetNodeData 获取节点数据
func (n *BaseNode) GetNodeData() *entity.BaseNodeData {
	return n.Data
}

// GetNextNode 获取下一个节点，默认返回第一个输出节点
func (n *BaseNode) GetNextNode(state map[string]any) string {
	if len(n.Data.Outputs) > 0 {
		return n.Data.Outputs[0]
	}
	return ""
}

// Validate 验证节点配置，子类可以覆盖此方法
func (n *BaseNode) Validate() error {
	return nil
}
