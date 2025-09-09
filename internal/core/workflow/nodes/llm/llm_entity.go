package llm

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// LLMNodeData represents the data structure for LLM workflow nodes
type LLMNodeData struct {
	*entities.BaseNodeData
	Prompt              string                     `json:"prompt"`
	LanguageModelConfig *LLMConfig                 `json:"model_config"`
	Inputs              []*entities.VariableEntity `json:"inputs"`
	Outputs             []*entities.VariableEntity `json:"outputs"`
}

type LLMConfig struct {
	Model       string  `json:"model"`       // 使用的模型名称
	Prompt      string  `json:"prompt"`      // 提示词模板
	MaxTokens   int     `json:"max_tokens"`  // 最大token数
	Temperature float64 `json:"temperature"` // 温度参数
}

// NewLLMNodeData creates a new LLM node data instance
func NewLLMNodeData() *LLMNodeData {
	return &LLMNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeLLM},
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (l *LLMNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return l.BaseNodeData
}
