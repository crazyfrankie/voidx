package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// ToolNodeData 工具节点数据
type ToolNodeData struct {
	*entity.BaseNodeData
	ToolID     uuid.UUID         `json:"tool_id"`    // 工具ID
	InputKeys  map[string]string `json:"input_keys"` // 输入键映射
	OutputKey  string            `json:"output_key"` // 输出键
	Parameters map[string]any    `json:"parameters"` // 工具参数
}

// ToolNode 工具节点
type ToolNode struct {
	BaseNode
	Data *ToolNodeData
	// TODO 实际项目中需要注入工具服务
	// toolService *service.ToolService
}

// NewToolNode 创建工具节点
func NewToolNode(data *ToolNodeData) *ToolNode {
	return &ToolNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行工具节点
func (n *ToolNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 准备工具输入参数
	toolInputs := make(map[string]any)
	for paramName, stateKey := range n.Data.InputKeys {
		if value, ok := state[stateKey]; ok {
			toolInputs[paramName] = value
		} else {
			return nil, fmt.Errorf("input key %s not found in state", stateKey)
		}
	}

	// 合并静态参数
	for k, v := range n.Data.Parameters {
		toolInputs[k] = v
	}

	// TODO 在实际项目中，这里应该调用工具服务
	// 这里只是模拟工具执行结果
	toolResult := map[string]any{
		"success": true,
		"result":  fmt.Sprintf("Tool %s executed with inputs: %v", n.Data.ToolID, toolInputs),
		"inputs":  toolInputs,
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "tool_result"
	}
	result[outputKey] = toolResult

	return result, nil
}

// Validate 验证工具节点配置
func (n *ToolNode) Validate() error {
	if n.Data.Type != entity.NodeTypeTool {
		return errors.New("invalid node type for tool node")
	}

	if n.Data.ToolID == uuid.Nil {
		return errors.New("tool_id is required for tool node")
	}

	if n.Data.InputKeys == nil {
		n.Data.InputKeys = make(map[string]string)
	}

	if n.Data.Parameters == nil {
		n.Data.Parameters = make(map[string]any)
	}

	return nil
}
