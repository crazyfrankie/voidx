package node

import (
	"context"
	"errors"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// StartNodeData 开始节点数据
type StartNodeData struct {
	*entity.BaseNodeData
	InitialState map[string]any `json:"initial_state"` // 初始状态
}

// StartNode 开始节点
type StartNode struct {
	BaseNode
	Data *StartNodeData
}

// NewStartNode 创建开始节点
func NewStartNode(data *StartNodeData) *StartNode {
	return &StartNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行开始节点
func (n *StartNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 开始节点将初始状态合并到当前状态
	result := make(map[string]any)

	// 复制当前状态
	for k, v := range state {
		result[k] = v
	}

	// 合并初始状态
	for k, v := range n.Data.InitialState {
		result[k] = v
	}

	return result, nil
}

// Validate 验证开始节点配置
func (n *StartNode) Validate() error {
	if n.Data.Type != entity.NodeTypeStart {
		return errors.New("invalid node type for start node")
	}

	if len(n.Data.Inputs) > 0 {
		return errors.New("start node cannot have inputs")
	}

	if len(n.Data.Outputs) == 0 {
		return errors.New("start node must have at least one output")
	}

	return nil
}
