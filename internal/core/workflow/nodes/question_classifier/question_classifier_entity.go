package question_classifier

import (
	"fmt"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// 问题分类器系统预设prompt
const QuestionClassifierSystemPrompt = `# 角色
你是一个文本分类引擎，负责对输入的文本进行分类，并返回相应的分类名称，如果没有匹配的分类则返回第一个分类，预设分类会以json列表的名称提供，请注意正确识别。

## 技能
### 技能1：文本分类
- 接收用户输入的文本内容。
- 使用自然语言处理技术分析文本特征。
- 根据预设的分类信息，将文本准确划分至相应类别，并返回分类名称。
- 分类名称格式为xxx_uuid，例如: qc_source_handle_1e3ac414-52f9-48f5-94fd-fbf4d3fe2df7，请注意识别。

## 预设分类信息
预设分类信息如下:
%s

## 限制
- 仅处理文本分类相关任务。
- 输出仅包含分类名称，不提供额外解释或信息。
- 确保分类结果的准确性，避免错误分类。
- 使用预设的分类标准进行判断，不进行主观解释。 
- 如果预设的分类没有符合条件，请直接返回第一个分类。`

// ClassConfig 问题分类器配置，存储分类query、连接的节点类型/id
type ClassConfig struct {
	Query          string `json:"query"`            // 问题分类对应的query描述
	NodeID         string `json:"node_id"`          // 该分类连接的节点id
	NodeType       string `json:"node_type"`        // 该分类连接的节点类型
	SourceHandleID string `json:"source_handle_id"` // 起点句柄id
}

// QuestionClassifierNodeData 问题分类器/意图识别节点数据
type QuestionClassifierNodeData struct {
	*entities.BaseNodeData
	Inputs  []*entities.VariableEntity `json:"inputs"`  // 输入变量信息
	Outputs []*entities.VariableEntity `json:"outputs"` // 输出变量信息
	Classes []*ClassConfig             `json:"classes"` // 分类配置列表
}

// NewQuestionClassifierNodeData 创建新的问题分类器节点数据
func NewQuestionClassifierNodeData() *QuestionClassifierNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeQuestionClassifier

	return &QuestionClassifierNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0), // 输出为空，因为这是一个路由节点
		Classes:      make([]*ClassConfig, 0),
	}
}

// ValidateInputs 校验输入变量信息
func (q *QuestionClassifierNodeData) ValidateInputs() error {
	// 1. 判断是否只有一个输入变量，如果有多个则抛出错误
	if len(q.Inputs) != 1 {
		return fmt.Errorf("问题分类节点输入变量信息出错")
	}

	// 2. 判断输入变量类型及字段名称是否出错
	queryInput := q.Inputs[0]
	if queryInput.Name != "query" || !queryInput.Required {
		return fmt.Errorf("问题分类节点输入变量名字/类型/必填属性出错")
	}

	return nil
}

// GetInputs 实现NodeDataInterface接口
func (q *QuestionClassifierNodeData) GetInputs() []*entities.VariableEntity {
	return q.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (q *QuestionClassifierNodeData) GetOutputs() []*entities.VariableEntity {
	return q.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (q *QuestionClassifierNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return q.BaseNodeData
}
