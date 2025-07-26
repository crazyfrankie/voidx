package node

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// ClassifierNodeData 分类节点数据
type ClassifierNodeData struct {
	*entity.BaseNodeData
	InputKey     string               `json:"input_key"`     // 输入键
	OutputKey    string               `json:"output_key"`    // 输出键
	Classes      []ClassifierClass    `json:"classes"`       // 分类定义
	DefaultClass string               `json:"default_class"` // 默认分类
	Method       string               `json:"method"`        // 分类方法：keyword, llm, rule
	LLMConfig    *ClassifierLLMConfig `json:"llm_config"`    // LLM配置（当method为llm时）
	Rules        []ClassifierRule     `json:"rules"`         // 规则配置（当method为rule时）
}

// ClassifierClass 分类定义
type ClassifierClass struct {
	ID          string   `json:"id"`          // 分类ID
	Name        string   `json:"name"`        // 分类名称
	Description string   `json:"description"` // 分类描述
	Keywords    []string `json:"keywords"`    // 关键词列表
	NextNode    string   `json:"next_node"`   // 下一个节点ID
}

// ClassifierLLMConfig LLM分类配置
type ClassifierLLMConfig struct {
	Provider    string  `json:"provider"`    // 模型提供者
	Model       string  `json:"model"`       // 模型名称
	Prompt      string  `json:"prompt"`      // 分类提示词
	Temperature float64 `json:"temperature"` // 温度
}

// ClassifierRule 分类规则
type ClassifierRule struct {
	Condition string `json:"condition"` // 条件表达式
	ClassID   string `json:"class_id"`  // 分类ID
}

// ClassifierNode 分类节点
type ClassifierNode struct {
	BaseNode
	Data *ClassifierNodeData
	// 实际项目中需要注入LLM服务
	// llmService *service.LLMService
}

// NewClassifierNode 创建分类节点
func NewClassifierNode(data *ClassifierNodeData) *ClassifierNode {
	return &ClassifierNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行分类节点
func (n *ClassifierNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 获取输入文本
	var inputText string
	if n.Data.InputKey != "" {
		if inputValue, ok := state[n.Data.InputKey]; ok {
			if strValue, ok := inputValue.(string); ok {
				inputText = strValue
			} else {
				return nil, fmt.Errorf("input value for key %s is not a string", n.Data.InputKey)
			}
		} else {
			return nil, fmt.Errorf("input key %s not found in state", n.Data.InputKey)
		}
	}

	// 执行分类
	var classifiedResult *ClassificationResult
	var err error

	switch n.Data.Method {
	case "keyword":
		classifiedResult = n.classifyByKeyword(inputText)
	case "llm":
		classifiedResult, err = n.classifyByLLM(ctx, inputText)
		if err != nil {
			return nil, fmt.Errorf("LLM classification failed: %w", err)
		}
	case "rule":
		classifiedResult = n.classifyByRule(inputText, state)
	default:
		classifiedResult = n.classifyByKeyword(inputText) // 默认使用关键词分类
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "classification_result"
	}
	result[outputKey] = classifiedResult

	return result, nil
}

// ClassificationResult 分类结果
type ClassificationResult struct {
	ClassID    string  `json:"class_id"`   // 分类ID
	ClassName  string  `json:"class_name"` // 分类名称
	Confidence float64 `json:"confidence"` // 置信度
	NextNode   string  `json:"next_node"`  // 下一个节点ID
	Method     string  `json:"method"`     // 分类方法
	InputText  string  `json:"input_text"` // 输入文本
}

// classifyByKeyword 基于关键词的分类
func (n *ClassifierNode) classifyByKeyword(inputText string) *ClassificationResult {
	inputLower := strings.ToLower(inputText)

	// 遍历所有分类，寻找匹配的关键词
	for _, class := range n.Data.Classes {
		for _, keyword := range class.Keywords {
			if strings.Contains(inputLower, strings.ToLower(keyword)) {
				return &ClassificationResult{
					ClassID:    class.ID,
					ClassName:  class.Name,
					Confidence: 0.8, // 关键词匹配的置信度设为0.8
					NextNode:   class.NextNode,
					Method:     "keyword",
					InputText:  inputText,
				}
			}
		}
	}

	// 如果没有匹配，返回默认分类
	return n.getDefaultClassification(inputText, "keyword")
}

// classifyByLLM 基于LLM的分类
func (n *ClassifierNode) classifyByLLM(ctx context.Context, inputText string) (*ClassificationResult, error) {
	if n.Data.LLMConfig == nil {
		return nil, errors.New("LLM config is required for LLM classification")
	}

	// 在实际项目中，这里应该调用LLM服务进行分类
	// 这里只是模拟LLM分类结果

	// 构建分类提示词
	classNames := make([]string, len(n.Data.Classes))
	for i, class := range n.Data.Classes {
		classNames[i] = fmt.Sprintf("%s: %s", class.ID, class.Description)
	}

	prompt := fmt.Sprintf("%s\n\nAvailable classes:\n%s\n\nInput text: %s",
		n.Data.LLMConfig.Prompt,
		strings.Join(classNames, "\n"),
		inputText)

	// 模拟LLM响应
	_ = prompt // 避免未使用变量警告

	// 简单的模拟逻辑：如果输入包含"问题"或"help"，分类为第一个类别
	if len(n.Data.Classes) > 0 && (strings.Contains(strings.ToLower(inputText), "问题") ||
		strings.Contains(strings.ToLower(inputText), "help")) {
		return &ClassificationResult{
			ClassID:    n.Data.Classes[0].ID,
			ClassName:  n.Data.Classes[0].Name,
			Confidence: 0.9,
			NextNode:   n.Data.Classes[0].NextNode,
			Method:     "llm",
			InputText:  inputText,
		}, nil
	}

	return n.getDefaultClassification(inputText, "llm"), nil
}

// classifyByRule 基于规则的分类
func (n *ClassifierNode) classifyByRule(inputText string, state map[string]any) *ClassificationResult {
	// 在实际项目中，这里应该实现规则引擎
	// 这里只是简单的模拟

	for _, rule := range n.Data.Rules {
		// 简单的条件判断（实际项目中应该使用表达式引擎）
		if n.evaluateRule(rule.Condition, inputText, state) {
			// 找到对应的分类
			for _, class := range n.Data.Classes {
				if class.ID == rule.ClassID {
					return &ClassificationResult{
						ClassID:    class.ID,
						ClassName:  class.Name,
						Confidence: 0.95, // 规则匹配的置信度设为0.95
						NextNode:   class.NextNode,
						Method:     "rule",
						InputText:  inputText,
					}
				}
			}
		}
	}

	return n.getDefaultClassification(inputText, "rule")
}

// evaluateRule 简单的规则评估（实际项目中应该使用表达式引擎）
func (n *ClassifierNode) evaluateRule(condition, inputText string, state map[string]any) bool {
	// 这里只是简单的字符串包含判断
	// 实际项目中应该实现完整的表达式引擎
	return strings.Contains(strings.ToLower(inputText), strings.ToLower(condition))
}

// getDefaultClassification 获取默认分类
func (n *ClassifierNode) getDefaultClassification(inputText, method string) *ClassificationResult {
	defaultClassID := n.Data.DefaultClass
	if defaultClassID == "" && len(n.Data.Classes) > 0 {
		defaultClassID = n.Data.Classes[0].ID
	}

	// 找到默认分类
	for _, class := range n.Data.Classes {
		if class.ID == defaultClassID {
			return &ClassificationResult{
				ClassID:    class.ID,
				ClassName:  class.Name,
				Confidence: 0.5, // 默认分类的置信度设为0.5
				NextNode:   class.NextNode,
				Method:     method,
				InputText:  inputText,
			}
		}
	}

	// 如果找不到默认分类，返回第一个分类
	if len(n.Data.Classes) > 0 {
		return &ClassificationResult{
			ClassID:    n.Data.Classes[0].ID,
			ClassName:  n.Data.Classes[0].Name,
			Confidence: 0.5,
			NextNode:   n.Data.Classes[0].NextNode,
			Method:     method,
			InputText:  inputText,
		}
	}

	// 如果没有任何分类定义，返回空结果
	return &ClassificationResult{
		ClassID:    "unknown",
		ClassName:  "Unknown",
		Confidence: 0.0,
		NextNode:   "",
		Method:     method,
		InputText:  inputText,
	}
}

// GetNextNode 根据分类结果获取下一个节点
func (n *ClassifierNode) GetNextNode(state map[string]any) string {
	// 从状态中获取分类结果
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "classification_result"
	}

	if result, ok := state[outputKey]; ok {
		if classResult, ok := result.(*ClassificationResult); ok {
			return classResult.NextNode
		}
	}

	// 如果没有分类结果，返回默认的下一个节点
	return n.BaseNode.GetNextNode(state)
}

// Validate 验证分类节点配置
func (n *ClassifierNode) Validate() error {
	if n.Data.Type != entity.NodeTypeClassifier {
		return errors.New("invalid node type for classifier node")
	}

	if n.Data.InputKey == "" {
		return errors.New("input_key is required for classifier node")
	}

	if len(n.Data.Classes) == 0 {
		return errors.New("at least one class is required for classifier node")
	}

	// 验证分类定义
	for i, class := range n.Data.Classes {
		if class.ID == "" {
			return fmt.Errorf("class[%d].id is required", i)
		}
		if class.Name == "" {
			return fmt.Errorf("class[%d].name is required", i)
		}
	}

	// 验证分类方法
	validMethods := []string{"keyword", "llm", "rule"}
	if n.Data.Method == "" {
		n.Data.Method = "keyword" // 默认关键词分类
	} else {
		valid := false
		for _, method := range validMethods {
			if n.Data.Method == method {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid method: %s, must be one of %v", n.Data.Method, validMethods)
		}
	}

	// 验证LLM配置
	if n.Data.Method == "llm" && n.Data.LLMConfig == nil {
		return errors.New("llm_config is required when method is 'llm'")
	}

	// 验证规则配置
	if n.Data.Method == "rule" && len(n.Data.Rules) == 0 {
		return errors.New("rules are required when method is 'rule'")
	}

	return nil
}
