package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// Workflow 工作流LangChain工具类
type Workflow struct {
	workflowConfig *entities.WorkflowConfig
}

// NewWorkflow 构造函数，完成工作流函数的初始化
func NewWorkflow(workflowConfig *entities.WorkflowConfig) (*Workflow, error) {
	// 验证工作流配置
	if err := workflowConfig.ValidateWorkflowConfig(); err != nil {
		return nil, fmt.Errorf("workflow config validation failed: %w", err)
	}

	return &Workflow{
		workflowConfig: workflowConfig,
	}, nil
}

// Name 实现Tool接口 - 返回工作流名称
func (w *Workflow) Name() string {
	return w.workflowConfig.Name
}

// Description 实现Tool接口 - 返回工作流描述
func (w *Workflow) Description() string {
	return w.workflowConfig.Description
}

// Call 执行工作流
func (w *Workflow) Call(ctx context.Context, input string) (string, error) {
	// 解析输入参数
	var inputMap map[string]any
	if input != "" {
		if err := json.Unmarshal([]byte(input), &inputMap); err != nil {
			return "", fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		inputMap = make(map[string]any)
	}

	// 创建工作流状态
	state := &entities.WorkflowState{}
	if err := state.SetInputsFromMap(inputMap); err != nil {
		return "", fmt.Errorf("failed to set inputs: %w", err)
	}

	// 执行工作流
	result, err := w.executeWorkflow(ctx, state)
	if err != nil {
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}

	// 获取输出结果
	outputs, err := result.GetOutputsAsMap()
	if err != nil {
		return "", fmt.Errorf("failed to get outputs: %w", err)
	}

	// 将输出结果序列化为字符串
	outputBytes, err := json.Marshal(outputs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outputs: %w", err)
	}

	return string(outputBytes), nil
}

// executeWorkflow 执行工作流的核心逻辑
func (w *Workflow) executeWorkflow(ctx context.Context, state *entities.WorkflowState) (*entities.WorkflowState, error) {
	// 这里实现工作流的执行逻辑
	// 由于Go版本中没有具体的节点实现，这里提供一个基础框架

	// 1. 找到开始节点
	var startNode *entities.BaseNodeData
	for _, node := range w.workflowConfig.Nodes {
		if node.NodeType == entities.NodeTypeStart {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return nil, fmt.Errorf("start node not found")
	}

	// 2. 构建节点执行顺序（这里简化处理，实际需要根据边的关系来确定执行顺序）
	executionOrder := w.buildExecutionOrder()

	// 3. 按顺序执行节点
	for _, nodeID := range executionOrder {
		node := w.findNodeByID(nodeID)
		if node == nil {
			continue
		}

		// 执行节点（这里需要根据具体的节点类型来实现）
		nodeResult, err := w.executeNode(ctx, node, state)
		if err != nil {
			return nil, fmt.Errorf("failed to execute node %s: %w", node.Title, err)
		}

		// 将节点结果添加到状态中
		state.NodeResults = append(state.NodeResults, nodeResult)
	}

	return state, nil
}

// buildExecutionOrder 构建节点执行顺序
func (w *Workflow) buildExecutionOrder() []string {
	// 这里应该实现拓扑排序来确定节点的执行顺序
	// 简化处理，直接返回节点ID列表
	var order []string
	for _, node := range w.workflowConfig.Nodes {
		order = append(order, node.ID.String())
	}
	return order
}

// findNodeByID 根据ID查找节点
func (w *Workflow) findNodeByID(nodeID string) *entities.BaseNodeData {
	for _, node := range w.workflowConfig.Nodes {
		if node.ID.String() == nodeID {
			return node
		}
	}
	return nil
}

// executeNode 执行单个节点
func (w *Workflow) executeNode(ctx context.Context, node *entities.BaseNodeData, state *entities.WorkflowState) (*entities.NodeResult, error) {
	// 创建节点结果
	result := entities.NewNodeResult(node)

	// 获取当前状态的输入数据
	inputs, err := state.GetInputsAsMap()
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to get inputs: %v", err)
		return result, err
	}

	result.Inputs = inputs

	// 根据节点类型执行不同的逻辑
	switch node.NodeType {
	case entities.NodeTypeStart:
		// 开始节点：直接传递输入到输出
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded

		// 更新状态的输出
		if err := state.SetOutputsFromMap(result.Outputs); err != nil {
			result.Status = entities.NodeStatusFailed
			result.Error = fmt.Sprintf("failed to set outputs: %v", err)
			return result, err
		}

	case entities.NodeTypeEnd:
		// 结束节点：输出最终结果
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded

	default:
		// 其他节点类型的处理需要根据具体实现来完成
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded
	}

	return result, nil
}

// Stream 工作流流式输出每个节点对应的结果
func (w *Workflow) Stream(ctx context.Context, input string) (<-chan *entities.NodeResult, error) {
	// 解析输入参数
	var inputMap map[string]any
	if input != "" {
		if err := json.Unmarshal([]byte(input), &inputMap); err != nil {
			return nil, fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		inputMap = make(map[string]any)
	}

	// 创建工作流状态
	state := &entities.WorkflowState{}
	if err := state.SetInputsFromMap(inputMap); err != nil {
		return nil, fmt.Errorf("failed to set inputs: %w", err)
	}

	// 创建结果通道
	resultChan := make(chan *entities.NodeResult, len(w.workflowConfig.Nodes))

	// 启动goroutine执行工作流
	go func() {
		defer close(resultChan)

		executionOrder := w.buildExecutionOrder()
		for _, nodeID := range executionOrder {
			node := w.findNodeByID(nodeID)
			if node == nil {
				continue
			}

			nodeResult, err := w.executeNode(ctx, node, state)
			if err != nil {
				nodeResult.Status = entities.NodeStatusFailed
				nodeResult.Error = err.Error()
			}

			// 发送节点结果到通道
			select {
			case resultChan <- nodeResult:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}
