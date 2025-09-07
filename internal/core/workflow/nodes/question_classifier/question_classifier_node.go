package question_classifier

import (
	"fmt"
	"strings"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// QuestionClassifierNode 问题分类器节点
type QuestionClassifierNode struct {
	*nodes.BaseNodeImpl
	nodeData *QuestionClassifierNodeData
}

// NewQuestionClassifierNode 创建新的问题分类器节点
func NewQuestionClassifierNode(nodeData *QuestionClassifierNodeData) *QuestionClassifierNode {
	return &QuestionClassifierNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke 覆盖重写invoke实现问题分类器节点，执行问题分类后返回节点的名称，如果LLM判断错误默认返回第一个节点名称
func (q *QuestionClassifierNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	// 1. 提取节点输入变量字典映射
	inputsDict := q.extractVariablesFromState(state)

	// 2. 构建分类信息
	classInfos := make([]map[string]any, 0, len(q.nodeData.Classes))
	for _, classConfig := range q.nodeData.Classes {
		classInfos = append(classInfos, map[string]any{
			"query": classConfig.Query,
			"class": fmt.Sprintf("qc_source_handle_%s", classConfig.SourceHandleID),
		})
	}

	// 3. 序列化分类信息
	classInfosJSON, err := sonic.Marshal(classInfos)
	if err != nil {
		return nil, fmt.Errorf("序列化分类信息失败: %v", err)
	}

	// 4. 构建提示词
	systemPrompt := fmt.Sprintf(QuestionClassifierSystemPrompt, string(classInfosJSON))
	query := "用户没有输入任何内容"
	if queryValue, exists := inputsDict["query"]; exists {
		query = fmt.Sprintf("%v", queryValue)
	}

	// 5. 调用LLM进行分类（这里需要实现LLM调用逻辑）
	nodeFlag, err := q.callLLMForClassification(systemPrompt, query)
	if err != nil {
		return nil, fmt.Errorf("LLM分类调用失败: %v", err)
	}

	// 6. 获取所有分类信息
	allClasses := make([]string, 0, len(q.nodeData.Classes))
	for _, item := range q.nodeData.Classes {
		allClasses = append(allClasses, fmt.Sprintf("qc_source_handle_%s", item.SourceHandleID))
	}

	// 7. 检测获取的分类标识是否在规定列表内，并提取节点标识
	if len(allClasses) == 0 {
		nodeFlag = "END"
	} else if !contains(allClasses, nodeFlag) {
		nodeFlag = allClasses[0]
	}

	// 8. 对于分类器节点，我们需要返回特殊的状态来指示下一个节点
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: state.NodeResults,
	}

	return newState, nil
}

// callLLMForClassification 调用LLM进行分类（模拟实现）
func (q *QuestionClassifierNode) callLLMForClassification(systemPrompt, query string) (string, error) {
	// 这里应该实现真正的LLM调用逻辑
	// 目前提供一个简单的模拟实现

	// 简单的关键词匹配逻辑
	queryLower := strings.ToLower(query)

	for _, classConfig := range q.nodeData.Classes {
		classQueryLower := strings.ToLower(classConfig.Query)
		if strings.Contains(queryLower, classQueryLower) || strings.Contains(classQueryLower, queryLower) {
			return fmt.Sprintf("qc_source_handle_%s", classConfig.SourceHandleID), nil
		}
	}

	// 如果没有匹配，返回第一个分类
	if len(q.nodeData.Classes) > 0 {
		return fmt.Sprintf("qc_source_handle_%s", q.nodeData.Classes[0].SourceHandleID), nil
	}

	return "END", nil
}

// extractVariablesFromState 从状态中提取变量
func (q *QuestionClassifierNode) extractVariablesFromState(state *entities.WorkflowState) map[string]any {
	result := make(map[string]any)

	for _, input := range q.nodeData.Inputs {
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

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
