package code

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// CodeLanguage 代码语言枚举
type CodeLanguage string

const (
	CodeLanguageGo     CodeLanguage = "go"
	CodeLanguagePython CodeLanguage = "python"
	CodeLanguageJS     CodeLanguage = "javascript"
)

// CodeNodeData 代码节点数据
type CodeNodeData struct {
	*entities.BaseNodeData
	Inputs   []*entities.VariableEntity `json:"inputs"`   // 输入变量列表
	Outputs  []*entities.VariableEntity `json:"outputs"`  // 输出变量列表
	Code     string                     `json:"code"`     // 代码内容
	Language CodeLanguage               `json:"language"` // 代码语言
}

// NewCodeNodeData 创建新的代码节点数据
func NewCodeNodeData() *CodeNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeCode

	return &CodeNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
		Language:     CodeLanguageGo,
	}
}

// GetInputs 实现NodeDataInterface接口
func (c *CodeNodeData) GetInputs() []*entities.VariableEntity {
	return c.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (c *CodeNodeData) GetOutputs() []*entities.VariableEntity {
	return c.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (c *CodeNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return c.BaseNodeData
}
