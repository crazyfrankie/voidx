package template_transform

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/bytedance/sonic"
	
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
)

// TemplateTransformNode 模板转换节点，将多个变量信息合并成一个
type TemplateTransformNode struct {
	*nodes.BaseNodeImpl
	nodeData *TemplateTransformNodeData
}

// NewTemplateTransformNode 创建新的模板转换节点
func NewTemplateTransformNode(nodeData *TemplateTransformNodeData) *TemplateTransformNode {
	return &TemplateTransformNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke 模板转换节点执行函数，将传递的多个变量合并成字符串后返回
func (t *TemplateTransformNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 1. 提取节点中的输入数据
	inputsDict := t.extractVariablesFromState(state)

	// 2. 使用Go template格式模板信息
	tmpl, err := template.New("transform").Parse(t.nodeData.Template)
	if err != nil {
		nodeResult := entities.NewNodeResult(t.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = fmt.Sprintf("模板解析失败: %v", err)
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	// 3. 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, inputsDict); err != nil {
		nodeResult := entities.NewNodeResult(t.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = fmt.Sprintf("模板渲染失败: %v", err)
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	templateValue := buf.String()

	// 4. 提取并构建输出数据结构
	outputs := map[string]any{
		"output": templateValue,
	}

	// 5. 构建节点结果
	nodeResult := entities.NewNodeResult(t.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = inputsDict
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	// 6. 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}

// extractVariablesFromState 从状态中提取变量
func (t *TemplateTransformNode) extractVariablesFromState(state *entities.WorkflowState) map[string]any {
	result := make(map[string]any)

	for _, input := range t.nodeData.Inputs {
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
