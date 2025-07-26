package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// CodeNodeData 代码节点数据
type CodeNodeData struct {
	*entity.BaseNodeData
	Language    string            `json:"language"`    // 编程语言：javascript, python, go
	Code        string            `json:"code"`        // 代码内容
	InputKeys   map[string]string `json:"input_keys"`  // 输入键映射
	OutputKey   string            `json:"output_key"`  // 输出键
	Timeout     int               `json:"timeout"`     // 超时时间（秒）
	Environment map[string]any    `json:"environment"` // 环境变量
}

// CodeNode 代码节点
type CodeNode struct {
	BaseNode
	Data *CodeNodeData
	// 实际项目中需要注入代码执行服务
	// codeExecutor *service.CodeExecutor
}

// NewCodeNode 创建代码节点
func NewCodeNode(data *CodeNodeData) *CodeNode {
	return &CodeNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行代码节点
func (n *CodeNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 准备代码执行的输入参数
	codeInputs := make(map[string]any)
	for paramName, stateKey := range n.Data.InputKeys {
		if value, ok := state[stateKey]; ok {
			codeInputs[paramName] = value
		} else {
			return nil, fmt.Errorf("input key %s not found in state", stateKey)
		}
	}

	// 在实际项目中，这里应该调用代码执行服务
	// 可以使用以下方案之一：
	// 1. 嵌入式JavaScript引擎（如otto或goja）
	// 2. 通过API调用外部代码执行服务
	// 3. 使用Docker容器执行代码

	// 这里只是模拟代码执行结果
	var codeResult map[string]any

	switch n.Data.Language {
	case "javascript":
		codeResult = n.executeJavaScript(codeInputs)
	case "python":
		codeResult = n.executePython(codeInputs)
	case "go":
		codeResult = n.executeGo(codeInputs)
	default:
		return nil, fmt.Errorf("unsupported language: %s", n.Data.Language)
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "code_result"
	}
	result[outputKey] = codeResult

	return result, nil
}

// executeJavaScript 模拟执行JavaScript代码
func (n *CodeNode) executeJavaScript(inputs map[string]any) map[string]any {
	// 在实际项目中，这里应该使用JavaScript引擎执行代码
	return map[string]any{
		"success":  true,
		"result":   fmt.Sprintf("JavaScript code executed with inputs: %v", inputs),
		"output":   "console.log('Hello from JavaScript');",
		"language": "javascript",
	}
}

// executePython 模拟执行Python代码
func (n *CodeNode) executePython(inputs map[string]any) map[string]any {
	// 在实际项目中，这里应该通过API或容器执行Python代码
	return map[string]any{
		"success":  true,
		"result":   fmt.Sprintf("Python code executed with inputs: %v", inputs),
		"output":   "print('Hello from Python')",
		"language": "python",
	}
}

// executeGo 模拟执行Go代码
func (n *CodeNode) executeGo(inputs map[string]any) map[string]any {
	// 在实际项目中，这里应该通过编译和执行Go代码
	return map[string]any{
		"success":  true,
		"result":   fmt.Sprintf("Go code executed with inputs: %v", inputs),
		"output":   `fmt.Println("Hello from Go")`,
		"language": "go",
	}
}

// Validate 验证代码节点配置
func (n *CodeNode) Validate() error {
	if n.Data.Type != entity.NodeTypeCode {
		return errors.New("invalid node type for code node")
	}

	if n.Data.Code == "" {
		return errors.New("code is required for code node")
	}

	validLanguages := []string{"javascript", "python", "go"}
	if n.Data.Language == "" {
		n.Data.Language = "javascript" // 默认JavaScript
	} else {
		valid := false
		for _, lang := range validLanguages {
			if n.Data.Language == lang {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid language: %s, must be one of %v", n.Data.Language, validLanguages)
		}
	}

	if n.Data.InputKeys == nil {
		n.Data.InputKeys = make(map[string]string)
	}

	if n.Data.Environment == nil {
		n.Data.Environment = make(map[string]any)
	}

	if n.Data.Timeout <= 0 {
		n.Data.Timeout = 30 // 默认30秒超时
	}

	return nil
}
