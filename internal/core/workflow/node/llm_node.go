package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// LLMNodeData LLM节点数据
type LLMNodeData struct {
	*entity.BaseNodeData
	Provider         string         `json:"provider"`          // 模型提供者
	Model            string         `json:"model"`             // 模型名称
	Prompt           string         `json:"prompt"`            // 提示词
	InputKey         string         `json:"input_key"`         // 输入键
	OutputKey        string         `json:"output_key"`        // 输出键
	Parameters       map[string]any `json:"parameters"`        // 模型参数
	MaxTokens        int            `json:"max_tokens"`        // 最大token数
	Temperature      float64        `json:"temperature"`       // 温度
	TopP             float64        `json:"top_p"`             // Top P
	FrequencyPenalty float64        `json:"frequency_penalty"` // 频率惩罚
	PresencePenalty  float64        `json:"presence_penalty"`  // 存在惩罚
}

// LLMNode LLM节点
type LLMNode struct {
	BaseNode
	Data *LLMNodeData
	// TODO 实际项目中需要注入LLM服务
	// llmService *service.LLMService
}

// NewLLMNode 创建LLM节点
func NewLLMNode(data *LLMNodeData) *LLMNode {
	return &LLMNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行LLM节点
func (n *LLMNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 获取输入
	var input string
	if n.Data.InputKey != "" {
		if inputValue, ok := state[n.Data.InputKey]; ok {
			if strValue, ok := inputValue.(string); ok {
				input = strValue
			} else {
				return nil, fmt.Errorf("input value for key %s is not a string", n.Data.InputKey)
			}
		} else {
			return nil, fmt.Errorf("input key %s not found in state", n.Data.InputKey)
		}
	}

	// TODO 在实际项目中，这里应该调用LLM服务
	// 这里只是模拟LLM的响应
	llmResponse := fmt.Sprintf("LLM response for prompt: %s, input: %s", n.Data.Prompt, input)

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "llm_output"
	}
	result[outputKey] = llmResponse

	return result, nil
}

// Validate 验证LLM节点配置
func (n *LLMNode) Validate() error {
	if n.Data.Type != entity.NodeTypeLLM {
		return errors.New("invalid node type for LLM node")
	}

	if n.Data.Provider == "" {
		return errors.New("provider is required for LLM node")
	}

	if n.Data.Model == "" {
		return errors.New("model is required for LLM node")
	}

	if n.Data.Prompt == "" {
		return errors.New("prompt is required for LLM node")
	}

	return nil
}
