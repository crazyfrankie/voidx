package code

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

const DefaultCode = `# 在这里编写你的Python代码
# 可以使用以下变量:
# - 所有输入变量都可以直接使用
# - 使用 result 变量来设置输出结果

result = "Hello, World!"
`

// CodeNodeData represents the data structure for code execution workflow nodes
type CodeNodeData struct {
	*entities.BaseNodeData
	Code    string                     `json:"code"`
	Inputs  []*entities.VariableEntity `json:"inputs"`
	Outputs []*entities.VariableEntity `json:"outputs"`
}

// NewCodeNodeData creates a new code node data instance
func NewCodeNodeData() *CodeNodeData {
	return &CodeNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeCode},
		Code:         DefaultCode,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (c *CodeNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return c.BaseNodeData
}
