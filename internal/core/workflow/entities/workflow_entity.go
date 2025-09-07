package entities

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/sonic"
)

var WorkflowConfigNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

const WorkflowConfigDescriptionMaxLength = 1024

// WorkflowConfig 工作流配置信息
type WorkflowConfig struct {
	AccountID   uuid.UUID       `json:"account_id"`  // 用户的唯一标识数据
	Name        string          `json:"name"`        // 工作流名称，必须是英文
	Description string          `json:"description"` // 工作流描述信息，用于告知LLM什么时候需要调用工作流
	Nodes       []*BaseNodeData `json:"nodes"`       // 工作流对应的节点列表信息
	Edges       []*BaseEdgeData `json:"edges"`       // 工作流对应的边列表信息
}

func NewWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		Nodes: make([]*BaseNodeData, 0),
		Edges: make([]*BaseEdgeData, 0),
	}
}

// WorkflowState 工作流图程序状态
type WorkflowState struct {
	Inputs      string        `json:"inputs"`       // 工作流的最初始输入，也就是工具输入 (使用string存储)
	Outputs     string        `json:"outputs"`      // 工作流的最终输出结果，也就是工具输出 (使用string存储)
	NodeResults []*NodeResult `json:"node_results"` // 各节点的运行结果
}

// GetInputsAsMap 将inputs字符串解析为map[string]any
func (ws *WorkflowState) GetInputsAsMap() (map[string]any, error) {
	if ws.Inputs == "" {
		return make(map[string]any), nil
	}
	var result map[string]any
	if err := sonic.Unmarshal([]byte(ws.Inputs), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}
	return result, nil
}

// SetInputsFromMap 将map[string]any转换为inputs字符串
func (ws *WorkflowState) SetInputsFromMap(inputs map[string]any) error {
	if inputs == nil {
		ws.Inputs = ""
		return nil
	}
	data, err := sonic.Marshal(inputs)
	if err != nil {
		return fmt.Errorf("failed to marshal inputs: %w", err)
	}
	ws.Inputs = string(data)
	return nil
}

// GetOutputsAsMap 将outputs字符串解析为map[string]any
func (ws *WorkflowState) GetOutputsAsMap() (map[string]any, error) {
	if ws.Outputs == "" {
		return make(map[string]any), nil
	}
	var result map[string]any
	if err := sonic.Unmarshal([]byte(ws.Outputs), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}
	return result, nil
}

// SetOutputsFromMap 将map[string]any转换为outputs字符串
func (ws *WorkflowState) SetOutputsFromMap(outputs map[string]any) error {
	if outputs == nil {
		ws.Outputs = ""
		return nil
	}
	data, err := sonic.Marshal(outputs)
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}
	ws.Outputs = string(data)
	return nil
}

// ProcessDict 工作流状态字典归纳函数
func ProcessDict(left, right map[string]any) map[string]any {
	if left == nil {
		left = make(map[string]any)
	}
	if right == nil {
		right = make(map[string]any)
	}

	result := make(map[string]any)
	for k, v := range left {
		result[k] = v
	}
	for k, v := range right {
		result[k] = v
	}
	return result
}

// ProcessNodeResults 工作流状态节点结果列表归纳函数
func ProcessNodeResults(left, right []*NodeResult) []*NodeResult {
	if left == nil {
		left = []*NodeResult{}
	}
	if right == nil {
		right = []*NodeResult{}
	}
	return append(left, right...)
}
