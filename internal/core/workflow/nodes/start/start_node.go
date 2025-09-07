package start

import (
	"fmt"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// StartNode 开始节点
type StartNode struct {
	*nodes.BaseNodeImpl
	nodeData *StartNodeData
}

// NewStartNode 创建新的开始节点
func NewStartNode(nodeData *StartNodeData) *StartNode {
	return &StartNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke 开始节点执行函数，该函数会提取状态中的输入信息并生成节点结果
func (s *StartNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 提取节点数据中的输入数据
	inputs := s.nodeData.Inputs

	// 循环遍历输入数据，并提取需要的数据，同时检测必填的数据是否传递
	outputs := make(map[string]any)

	originInputs := make(map[string]any)
	if err := sonic.UnmarshalString(state.Inputs, &originInputs); err != nil {
		return nil, err
	}
	for _, input := range inputs {
		inputValue, exists := originInputs[input.Name]
		// 检测字段是否必填，如果是则检测是否赋值
		if !exists || inputValue == nil {
			if input.Required {
				return nil, fmt.Errorf("工作流参数生成出错，%s为必填参数", input.Name)
			} else {
				inputValue = entities.VariableTypeDefaultValueMap[input.Type]
			}
		}

		// 提取出输出数据
		outputs[input.Name] = inputValue
	}

	// 构建节点结果
	nodeResult := entities.NewNodeResult(s.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = originInputs
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	// 构建状态数据并返回
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}
