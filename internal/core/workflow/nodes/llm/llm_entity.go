package llm

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// LLMNodeData LLM节点数据
type LLMNodeData struct {
	*entities.BaseNodeData
	Inputs      []*entities.VariableEntity `json:"inputs"`      // 输入变量列表
	Outputs     []*entities.VariableEntity `json:"outputs"`     // 输出变量列表
	Model       string                     `json:"model"`       // 使用的模型名称
	Prompt      string                     `json:"prompt"`      // 提示词模板
	MaxTokens   int                        `json:"max_tokens"`  // 最大token数
	Temperature float64                    `json:"temperature"` // 温度参数
}

// NewLLMNodeData 创建新的LLM节点数据
func NewLLMNodeData() *LLMNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeLLM

	return &LLMNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
		Model:        "gpt-3.5-turbo",
		MaxTokens:    1000,
		Temperature:  0.7,
	}
}

// GetInputs 实现NodeDataInterface接口
func (l *LLMNodeData) GetInputs() []*entities.VariableEntity {
	return l.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (l *LLMNodeData) GetOutputs() []*entities.VariableEntity {
	return l.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (l *LLMNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return l.BaseNodeData
}
