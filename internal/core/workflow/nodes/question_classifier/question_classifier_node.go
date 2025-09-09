package question_classifier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/utils"
)

// Question classifier system prompt
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

// QuestionClassifierNode represents a question classifier workflow node
type QuestionClassifierNode struct {
	nodeData *QuestionClassifierNodeData
	llmModel model.BaseChatModel
}

// ClassInfo represents class information for the prompt
type ClassInfo struct {
	Query string `json:"query"`
	Class string `json:"class"`
}

// NewQuestionClassifierNode creates a new question classifier node instance
func NewQuestionClassifierNode(nodeData *QuestionClassifierNodeData, llmModel model.BaseChatModel) *QuestionClassifierNode {

	return &QuestionClassifierNode{
		nodeData: nodeData,
		llmModel: llmModel,
	}
}

// Execute executes the question classifier node
func (n *QuestionClassifierNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Extract input variables from state
	inputsDict, err := utils.ExtractVariablesFromState(n.nodeData.Inputs, state)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to extract input variables: %v", err)
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	result.Inputs = inputsDict

	// Get the query from inputs
	query, ok := inputsDict["query"]
	if !ok {
		query = "用户没有输入任何内容"
	}

	queryStr, ok := query.(string)
	if !ok {
		queryStr = fmt.Sprintf("%v", query)
	}

	// Build class information for the prompt
	classInfos := make([]ClassInfo, len(n.nodeData.Classes))
	for i, class := range n.nodeData.Classes {
		classInfos[i] = ClassInfo{
			Query: class.Query,
			Class: fmt.Sprintf("qc_source_handle_%s", class.SourceHandleID),
		}
	}

	// Convert class information to JSON
	classInfoJSON, err := json.Marshal(classInfos)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to marshal class information: %v", err)
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Build the system prompt with class information
	systemPrompt := fmt.Sprintf(QuestionClassifierSystemPrompt, string(classInfoJSON))

	// Create messages for the LLM
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: queryStr,
		},
	}

	// Call the LLM model
	response, err := n.llmModel.Generate(ctx, messages, model.WithTemperature(0.0), model.WithMaxTokens(512))
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("LLM generation failed: %v", err)
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Extract the classification result
	nodeFlag := strings.TrimSpace(response.Content)

	// Get all valid class names
	allClasses := n.nodeData.GetClassNames()

	// Validate the classification result
	if len(allClasses) == 0 {
		nodeFlag = "END"
	} else {
		// Check if the result is in the valid classes
		validClass := false
		for _, class := range allClasses {
			if nodeFlag == class {
				validClass = true
				break
			}
		}

		// If not valid, use the first class
		if !validClass {
			nodeFlag = allClasses[0]
		}
	}

	// Set successful result
	result.Status = entities.NodeStatusSucceeded
	result.Outputs = map[string]any{
		"classification": nodeFlag,
	}
	result.EndTime = time.Now().Unix()

	return result, nil
}

// GetNextNodeFlag returns the next node flag based on classification result
// This method is used by the workflow engine to determine the next node to execute
func (n *QuestionClassifierNode) GetNextNodeFlag(ctx context.Context, state *entities.WorkflowState) (string, error) {
	// Execute the node to get the classification result
	result, err := n.Execute(ctx, state)
	if err != nil {
		return "", err
	}

	if result.Status == entities.NodeStatusFailed {
		return "", fmt.Errorf("question classifier node failed: %s", result.Error)
	}

	// Extract the classification result
	classification, ok := result.Outputs["classification"].(string)
	if !ok {
		// Fallback to first class if extraction fails
		allClasses := n.nodeData.GetClassNames()
		if len(allClasses) > 0 {
			return allClasses[0], nil
		}
		return "END", nil
	}

	return classification, nil
}
