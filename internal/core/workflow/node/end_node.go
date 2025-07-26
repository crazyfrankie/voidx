package node

import (
	"context"
	"errors"
	entity2 "github.com/crazyfrankie/voidx/internal/models/entity"
)

// EndNodeData 结束节点数据
type EndNodeData struct {
	*entity2.BaseNodeData
	OutputKeys []string `json:"output_keys"` // 输出键列表
}

// EndNode 结束节点
type EndNode struct {
	BaseNode
	Data *EndNodeData
}

// NewEndNode 创建结束节点
func NewEndNode(data *EndNodeData) *EndNode {
	return &EndNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行结束节点
func (n *EndNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 结束节点只返回指定的输出键对应的值
	result := make(map[string]any)

	// 如果没有指定输出键，则返回所有状态
	if len(n.Data.OutputKeys) == 0 {
		for k, v := range state {
			result[k] = v
		}
		return result, nil
	}

	// 否则只返回指定的输出键
	for _, key := range n.Data.OutputKeys {
		if value, ok := state[key]; ok {
			result[key] = value
		}
	}

	return result, nil
}

// GetNextNode 结束节点没有下一个节点
func (n *EndNode) GetNextNode(state map[string]any) string {
	return ""
}

// Validate 验证结束节点配置
func (n *EndNode) Validate() error {
	if n.Data.Type != entity2.NodeTypeEnd {
		return errors.New("invalid node type for end node")
	}

	if len(n.Data.Outputs) > 0 {
		return errors.New("end node cannot have outputs")
	}

	if len(n.Data.Inputs) == 0 {
		return errors.New("end node must have at least one input")
	}

	return nil
}
