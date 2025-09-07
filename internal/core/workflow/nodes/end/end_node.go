package end

import (
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// EndNode 结束节点
type EndNode struct {
	*nodes.BaseNodeImpl
	nodeData *EndNodeData
}

// NewEndNode 创建新的结束节点
func NewEndNode(nodeData *EndNodeData) *EndNode {
	return &EndNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke 结束节点执行函数，该函数会处理输出数据并生成最终结果
func (e *EndNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 处理输出数据
	outputs := make(map[string]any)

	for _, output := range e.nodeData.Outputs {
		// 这里简化处理，实际应该根据变量的value类型来处理引用或字面值
		if output.Value.Type == entities.VariableValueTypeRef {
			// 处理引用类型的输出
			// 这里需要根据引用信息从前置节点获取数据
			// 简化实现，直接从state中获取
			inputs := make(map[string]any)
			if err := sonic.UnmarshalString(state.Inputs, &inputs); err != nil {
				return nil, err
			}
			if val, exists := inputs[output.Name]; exists {
				outputs[output.Name] = val
			}
		} else {
			// 处理字面值类型的输出
			outputs[output.Name] = output.Value.Content
		}
	}

	// 构建节点结果
	nodeResult := entities.NewNodeResult(e.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = make(map[string]any) // 结束节点通常没有特定输入
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	output, err := sonic.MarshalString(outputs)
	if err != nil {
		return nil, err
	}

	// 构建状态数据并返回，设置最终输出
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     output, // 设置工作流的最终输出
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}
