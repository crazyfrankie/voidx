package iteration

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/logs"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// WorkflowInterface 工作流接口
type WorkflowInterface interface {
	Invoke(inputs map[string]any) (map[string]any, error)
	GetArgs() map[string]any
}

// IterationNode 迭代节点
type IterationNode struct {
	*nodes.BaseNodeImpl
	nodeData *IterationNodeData
	workflow WorkflowInterface
}

// NewIterationNode 创建新的迭代节点
func NewIterationNode(nodeData *IterationNodeData) *IterationNode {
	node := &IterationNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}

	// 初始化工作流
	if err := node.initializeWorkflow(); err != nil {
		logs.Errorf("Failed to build iteration node sub-workflow: %v", err)
		node.workflow = nil
	}

	return node
}

// initializeWorkflow 初始化工作流
func (i *IterationNode) initializeWorkflow() error {
	// 1. 判断是否传递的工作流id
	if len(i.nodeData.WorkflowIDs) != 1 {
		i.workflow = nil
		return nil
	}

	// 2. 根据工作流ID获取工作流（这里需要实现数据库查询逻辑）
	workflowID := i.nodeData.WorkflowIDs[0]
	workflow, err := i.getWorkflowByID(workflowID)
	if err != nil {
		return fmt.Errorf("获取工作流失败: %v", err)
	}

	i.workflow = workflow
	return nil
}

// getWorkflowByID 根据ID获取工作流（模拟实现）
func (i *IterationNode) getWorkflowByID(workflowID uuid.UUID) (WorkflowInterface, error) {
	// 这里应该实现真正的数据库查询逻辑
	// 目前提供一个模拟实现
	return &MockWorkflow{
		id:   workflowID,
		args: map[string]any{"input": "string"},
	}, nil
}

// Invoke 迭代节点调用函数，循环遍历将工作流的结果进行输出
func (i *IterationNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 1. 提取节点输入变量字典映射
	inputsDict := i.extractVariablesFromState(state)
	inputs, exists := inputsDict["inputs"]
	if !exists {
		inputs = []any{}
	}

	// 2. 异常检测，涵盖工作流不存在、工作流输入参数不唯一、数据为非列表、长度为0等
	inputsList, ok := inputs.([]any)
	if i.workflow == nil || !ok || len(inputsList) == 0 {
		nodeResult := entities.NewNodeResult(i.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Inputs = inputsDict
		nodeResult.Outputs = map[string]any{"outputs": []any{}}
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, nil
	}

	// 3. 获取工作流的输入字段结构
	args := i.workflow.GetArgs()
	if len(args) != 1 {
		nodeResult := entities.NewNodeResult(i.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Inputs = inputsDict
		nodeResult.Outputs = map[string]any{"outputs": []any{}}
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, nil
	}

	// 获取参数键
	var paramKey string
	for key := range args {
		paramKey = key
		break
	}

	// 4. 工作流+数据均存在，则循环遍历输入数据调用迭代工作流获取结果
	outputs := make([]any, 0, len(inputsList))
	for _, item := range inputsList {
		// 5. 构建输入字典信息
		data := map[string]any{paramKey: item}

		// 6. 调用工作流获取结果
		iterationResult, err := i.workflow.Invoke(data)
		if err != nil {
			logs.Errorf("Failed to invoke iteration workflow: %v", err)
			continue
		}

		// 转换成JSON字符串
		jsonBytes, err := sonic.Marshal(iterationResult)
		if err != nil {
			logs.Errorf("Failed to serialize iteration results: %v", err)
			continue
		}

		outputs = append(outputs, string(jsonBytes))
	}

	// 7. 构建节点结果
	nodeResult := entities.NewNodeResult(i.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = inputsDict
	nodeResult.Outputs = map[string]any{"outputs": outputs}
	nodeResult.Latency = time.Since(startAt)

	// 8. 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}

// extractVariablesFromState 从状态中提取变量
func (i *IterationNode) extractVariablesFromState(state *entities.WorkflowState) map[string]any {
	result := make(map[string]any)

	for _, input := range i.nodeData.Inputs {
		inputs := make(map[string]any)
		if err := sonic.UnmarshalString(state.Inputs, &inputs); err != nil {
			return nil
		}
		if val, exists := inputs[input.Name]; exists {
			result[input.Name] = val
		}
	}

	return result
}

// MockWorkflow 模拟工作流实现
type MockWorkflow struct {
	id   uuid.UUID
	args map[string]any
}

func (m *MockWorkflow) Invoke(inputs map[string]any) (map[string]any, error) {
	// 模拟工作流执行
	result := make(map[string]any)
	for key, value := range inputs {
		result[fmt.Sprintf("processed_%s", key)] = fmt.Sprintf("处理后的值: %v", value)
	}
	result["timestamp"] = time.Now().Unix()
	return result, nil
}

func (m *MockWorkflow) GetArgs() map[string]any {
	return m.args
}
