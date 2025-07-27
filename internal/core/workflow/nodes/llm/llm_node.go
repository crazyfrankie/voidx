package llm

import (
	"fmt"
	"github.com/bytedance/sonic"
	"strings"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
)

// LLMNode LLM节点
type LLMNode struct {
	*nodes.BaseNodeImpl
	nodeData *LLMNodeData
}

// NewLLMNode 创建新的LLM节点
func NewLLMNode(nodeData *LLMNodeData) *LLMNode {
	return &LLMNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke LLM节点执行函数
func (l *LLMNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 处理输入数据
	inputs := make(map[string]any)
	originInputs := make(map[string]any)
	if err := sonic.UnmarshalString(state.Inputs, &originInputs); err != nil {
		return nil, err
	}
	for _, input := range l.nodeData.Inputs {
		if val, exists := originInputs[input.Name]; exists {
			inputs[input.Name] = val
		} else if input.Required {
			return nil, fmt.Errorf("LLM节点缺少必需的输入参数: %s", input.Name)
		}
	}

	// 构建提示词
	prompt := l.buildPrompt(inputs)

	// 模拟LLM调用（实际实现中应该调用真实的LLM API）
	response := l.simulateLLMCall(prompt)

	// 处理输出数据
	outputs := make(map[string]any)
	for _, output := range l.nodeData.Outputs {
		switch output.Name {
		case "response":
			outputs[output.Name] = response
		case "prompt_used":
			outputs[output.Name] = prompt
		case "model":
			outputs[output.Name] = l.nodeData.Model
		default:
			outputs[output.Name] = ""
		}
	}

	// 构建节点结果
	nodeResult := entities.NewNodeResult(l.nodeData.BaseNodeData)
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

// buildPrompt 构建提示词
func (l *LLMNode) buildPrompt(inputs map[string]any) string {
	prompt := l.nodeData.Prompt

	// 简单的模板替换
	for key, value := range inputs {
		placeholder := fmt.Sprintf("{{%s}}", key)
		prompt = strings.ReplaceAll(prompt, placeholder, fmt.Sprintf("%v", value))
	}

	return prompt
}

// simulateLLMCall 模拟LLM调用
func (l *LLMNode) simulateLLMCall(prompt string) string {
	// 这里是模拟实现，实际应该调用真实的LLM API
	return fmt.Sprintf("基于提示词 '%s' 生成的模拟响应，使用模型: %s",
		prompt, l.nodeData.Model)
}
