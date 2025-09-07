package code

import (
	"fmt"
	"strings"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// CodeNode 代码节点
type CodeNode struct {
	*nodes.BaseNodeImpl
	nodeData *CodeNodeData
}

// NewCodeNode 创建新的代码节点
func NewCodeNode(nodeData *CodeNodeData) *CodeNode {
	return &CodeNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke 代码节点执行函数
func (c *CodeNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 处理输入数据
	inputs := make(map[string]any)
	originInputs := make(map[string]any)
	if err := sonic.UnmarshalString(state.Inputs, &originInputs); err != nil {
		return nil, err
	}
	for _, input := range c.nodeData.Inputs {
		if val, exists := originInputs[input.Name]; exists {
			inputs[input.Name] = val
		} else if input.Required {
			return nil, fmt.Errorf("代码节点缺少必需的输入参数: %s", input.Name)
		}
	}

	// 执行代码
	result, err := c.executeCode(inputs)
	if err != nil {
		nodeResult := entities.NewNodeResult(c.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = err.Error()
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	// 处理输出数据
	outputs := make(map[string]any)
	for _, output := range c.nodeData.Outputs {
		if val, exists := result[output.Name]; exists {
			outputs[output.Name] = val
		} else {
			outputs[output.Name] = entities.VariableTypeDefaultValueMap[output.Type]
		}
	}

	// 构建节点结果
	nodeResult := entities.NewNodeResult(c.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = inputs
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	// 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}

// executeCode 执行代码
func (c *CodeNode) executeCode(inputs map[string]any) (map[string]any, error) {
	// 这里是简化的代码执行实现
	// 实际实现中应该根据语言类型调用相应的解释器或编译器

	switch c.nodeData.Language {
	case CodeLanguageGo:
		return c.executeGoCode(inputs)
	case CodeLanguagePython:
		return c.executePythonCode(inputs)
	case CodeLanguageJS:
		return c.executeJSCode(inputs)
	default:
		return nil, fmt.Errorf("不支持的代码语言: %s", c.nodeData.Language)
	}
}

// executeGoCode 执行Go代码（模拟实现）
func (c *CodeNode) executeGoCode(inputs map[string]any) (map[string]any, error) {
	// 模拟Go代码执行
	result := make(map[string]any)

	// 简单的字符串处理示例
	if input, exists := inputs["input"]; exists {
		inputStr := fmt.Sprintf("%v", input)
		result["output"] = strings.ToUpper(inputStr)
		result["length"] = len(inputStr)
	}

	result["language"] = "go"
	result["code_executed"] = c.nodeData.Code

	return result, nil
}

// executePythonCode 执行Python代码（模拟实现）
func (c *CodeNode) executePythonCode(inputs map[string]any) (map[string]any, error) {
	// 模拟Python代码执行
	result := make(map[string]any)

	if input, exists := inputs["input"]; exists {
		inputStr := fmt.Sprintf("%v", input)
		result["output"] = strings.ToLower(inputStr)
		result["reversed"] = reverseString(inputStr)
	}

	result["language"] = "python"
	result["code_executed"] = c.nodeData.Code

	return result, nil
}

// executeJSCode 执行JavaScript代码（模拟实现）
func (c *CodeNode) executeJSCode(inputs map[string]any) (map[string]any, error) {
	// 模拟JavaScript代码执行
	result := make(map[string]any)

	if input, exists := inputs["input"]; exists {
		inputStr := fmt.Sprintf("%v", input)
		result["output"] = fmt.Sprintf("JS: %s", inputStr)
		result["char_count"] = len(inputStr)
	}

	result["language"] = "javascript"
	result["code_executed"] = c.nodeData.Code

	return result, nil
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
