package nodes

import (
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/code"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/dataset_retrieval"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/end"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/http_request"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/iteration"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/llm"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/question_classifier"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/start"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/template_transform"
	toolNode "github.com/crazyfrankie/voidx/internal/core/workflow/nodes/tool"
)

// NodeFactory creates workflow nodes based on node data
type NodeFactory struct {
	llmModel         model.BaseChatModel
	retrieverService *retrievers.RetrieverService
	toolManager      map[string]tool.InvokableTool
}

// NewNodeFactory creates a new node factory instance
func NewNodeFactory(llmModel model.BaseChatModel, retrieverService *retrievers.RetrieverService) *NodeFactory {
	return &NodeFactory{
		llmModel:         llmModel,
		retrieverService: retrieverService,
		toolManager:      make(map[string]tool.InvokableTool),
	}
}

// RegisterTool registers a tool with the factory
func (f *NodeFactory) RegisterTool(name string, tool tool.InvokableTool) {
	f.toolManager[name] = tool
}

// CreateNode creates a workflow node based on the node data
func (f *NodeFactory) CreateNode(nodeData entities.NodeDataInterface, accountID uuid.UUID) (NodeExecutor, error) {
	baseNodeData := nodeData.GetBaseNodeData()

	switch baseNodeData.NodeType {
	case entities.NodeTypeStart:
		if startData, ok := nodeData.(*start.StartNodeData); ok {
			return start.NewStartNode(startData), nil
		}
		return nil, fmt.Errorf("invalid start node data type")

	case entities.NodeTypeEnd:
		if endData, ok := nodeData.(*end.EndNodeData); ok {
			return end.NewEndNode(endData), nil
		}
		return nil, fmt.Errorf("invalid end node data type")

	case entities.NodeTypeLLM:
		if llmData, ok := nodeData.(*llm.LLMNodeData); ok {
			return llm.NewLLMNode(llmData, f.llmModel), nil
		}
		return nil, fmt.Errorf("invalid LLM node data type")

	case entities.NodeTypeTemplateTransform:
		if templateData, ok := nodeData.(*template_transform.TemplateTransformNodeData); ok {
			return template_transform.NewTemplateTransformNode(templateData), nil
		}
		return nil, fmt.Errorf("invalid template transform node data type")

	case entities.NodeTypeDatasetRetrieval:
		if datasetData, ok := nodeData.(*dataset_retrieval.DatasetRetrievalNodeData); ok {
			return dataset_retrieval.NewDatasetRetrievalNode(datasetData, f.retrieverService, accountID), nil
		}
		return nil, fmt.Errorf("invalid dataset retrieval node data type")

	case entities.NodeTypeCode:
		if codeData, ok := nodeData.(*code.CodeNodeData); ok {
			return code.NewCodeNode(codeData), nil
		}
		return nil, fmt.Errorf("invalid code node data type")

	case entities.NodeTypeTool:
		if toolData, ok := nodeData.(*toolNode.ToolNodeData); ok {
			// Get the tool from the tool manager
			tool, exists := f.toolManager[toolData.ToolName]
			if !exists {
				return nil, fmt.Errorf("tool %s not found", toolData.ToolName)
			}
			return toolNode.NewToolNode(toolData, tool), nil
		}
		return nil, fmt.Errorf("invalid tool node data type")

	case entities.NodeTypeHTTPRequest:
		if httpData, ok := nodeData.(*http_request.HTTPRequestNodeData); ok {
			return http_request.NewHTTPRequestNode(httpData), nil
		}
		return nil, fmt.Errorf("invalid HTTP request node data type")

	case entities.NodeTypeIteration:
		if iterationData, ok := nodeData.(*iteration.IterationNodeData); ok {
			return iteration.NewIterationNode(iterationData), nil
		}
		return nil, fmt.Errorf("invalid iteration node data type")

	case entities.NodeTypeQuestionClassifier:
		if qcData, ok := nodeData.(*question_classifier.QuestionClassifierNodeData); ok {
			return question_classifier.NewQuestionClassifierNode(qcData, f.llmModel), nil
		}
		return nil, fmt.Errorf("invalid question classifier node data type")

	default:
		return nil, fmt.Errorf("unsupported node type: %s", baseNodeData.NodeType)
	}
}

// ParseNodeData parses node data from a map based on node type
func (f *NodeFactory) ParseNodeData(nodeMap map[string]interface{}) (entities.NodeDataInterface, error) {
	nodeTypeStr, ok := nodeMap["node_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	nodeType := entities.NodeType(nodeTypeStr)

	switch nodeType {
	case entities.NodeTypeStart:
		nodeData := start.NewStartNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse inputs if present
		if inputs, exists := nodeMap["inputs"]; exists {
			if inputsSlice, ok := inputs.([]interface{}); ok {
				for _, input := range inputsSlice {
					if inputMap, ok := input.(map[string]interface{}); ok {
						variable, err := f.parseVariableEntity(inputMap)
						if err != nil {
							return nil, fmt.Errorf("failed to parse input variable: %w", err)
						}
						nodeData.Inputs = append(nodeData.Inputs, variable)
					}
				}
			}
		}
		return nodeData, nil

	case entities.NodeTypeEnd:
		nodeData := end.NewEndNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse outputs if present
		if outputs, exists := nodeMap["outputs"]; exists {
			if outputsSlice, ok := outputs.([]interface{}); ok {
				for _, output := range outputsSlice {
					if outputMap, ok := output.(map[string]interface{}); ok {
						variable, err := f.parseVariableEntity(outputMap)
						if err != nil {
							return nil, fmt.Errorf("failed to parse output variable: %w", err)
						}
						nodeData.Outputs = append(nodeData.Outputs, variable)
					}
				}
			}
		}
		return nodeData, nil

	case entities.NodeTypeLLM:
		nodeData := llm.NewLLMNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse prompt
		if prompt, exists := nodeMap["prompt"]; exists {
			if promptStr, ok := prompt.(string); ok {
				nodeData.Prompt = promptStr
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeTemplateTransform:
		nodeData := template_transform.NewTemplateTransformNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse template
		if template, exists := nodeMap["template"]; exists {
			if templateStr, ok := template.(string); ok {
				nodeData.Template = templateStr
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeDatasetRetrieval:
		nodeData := dataset_retrieval.NewDatasetRetrievalNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse dataset IDs
		if datasetIDs, exists := nodeMap["dataset_ids"]; exists {
			if idsSlice, ok := datasetIDs.([]interface{}); ok {
				for _, id := range idsSlice {
					if idStr, ok := id.(string); ok {
						if parsedID, err := uuid.Parse(idStr); err == nil {
							nodeData.DatasetIDs = append(nodeData.DatasetIDs, parsedID)
						}
					}
				}
			}
		}
		// Parse retrieval config
		if config, exists := nodeMap["retrieval_config"]; exists {
			if configMap, ok := config.(map[string]interface{}); ok {
				nodeData.RetrievalConfig = configMap
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeCode:
		nodeData := code.NewCodeNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse code
		if codeStr, exists := nodeMap["code"]; exists {
			if code, ok := codeStr.(string); ok {
				nodeData.Code = code
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeTool:
		nodeData := toolNode.NewToolNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse tool name
		if toolName, exists := nodeMap["tool_name"]; exists {
			if name, ok := toolName.(string); ok {
				nodeData.ToolName = name
			}
		}
		// Parse tool config
		if config, exists := nodeMap["tool_config"]; exists {
			if configMap, ok := config.(map[string]interface{}); ok {
				nodeData.ToolConfig = configMap
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeHTTPRequest:
		nodeData := http_request.NewHTTPRequestNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse HTTP specific fields
		if url, exists := nodeMap["url"]; exists {
			if urlStr, ok := url.(string); ok {
				nodeData.URL = urlStr
			}
		}
		if method, exists := nodeMap["method"]; exists {
			if methodStr, ok := method.(string); ok {
				nodeData.Method = http_request.HTTPMethod(methodStr)
			}
		}
		if headers, exists := nodeMap["headers"]; exists {
			if headersMap, ok := headers.(map[string]interface{}); ok {
				nodeData.Headers = make(map[string]string)
				for k, v := range headersMap {
					if vStr, ok := v.(string); ok {
						nodeData.Headers[k] = vStr
					}
				}
			}
		}
		if body, exists := nodeMap["body"]; exists {
			if bodyStr, ok := body.(string); ok {
				nodeData.Body = bodyStr
			}
		}
		if timeout, exists := nodeMap["timeout"]; exists {
			if timeoutInt, ok := timeout.(int); ok {
				nodeData.Timeout = timeoutInt
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		return nodeData, nil

	case entities.NodeTypeIteration:
		nodeData := iteration.NewIterationNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse workflow IDs
		if workflowIDs, exists := nodeMap["workflow_ids"]; exists {
			if idsSlice, ok := workflowIDs.([]interface{}); ok {
				for _, id := range idsSlice {
					if idStr, ok := id.(string); ok {
						if parsedID, err := uuid.Parse(idStr); err == nil {
							nodeData.WorkflowIDs = append(nodeData.WorkflowIDs, parsedID)
						}
					}
				}
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		// Validate the iteration node data
		if err := nodeData.Validate(); err != nil {
			return nil, fmt.Errorf("iteration node validation failed: %w", err)
		}
		return nodeData, nil

	case entities.NodeTypeQuestionClassifier:
		nodeData := question_classifier.NewQuestionClassifierNodeData()
		if err := f.parseBaseNodeData(nodeMap, nodeData.BaseNodeData); err != nil {
			return nil, err
		}
		// Parse classes
		if classes, exists := nodeMap["classes"]; exists {
			if classesSlice, ok := classes.([]interface{}); ok {
				for _, class := range classesSlice {
					if classMap, ok := class.(map[string]interface{}); ok {
						classConfig := &question_classifier.ClassConfig{}

						if query, ok := classMap["query"].(string); ok {
							classConfig.Query = query
						}
						if nodeID, ok := classMap["node_id"].(string); ok {
							classConfig.NodeID = nodeID
						}
						if nodeType, ok := classMap["node_type"].(string); ok {
							classConfig.NodeType = nodeType
						}
						if sourceHandleID, ok := classMap["source_handle_id"].(string); ok {
							classConfig.SourceHandleID = sourceHandleID
						}

						nodeData.Classes = append(nodeData.Classes, classConfig)
					}
				}
			}
		}
		// Parse inputs and outputs
		if err := f.parseInputsOutputs(nodeMap, &nodeData.Inputs, &nodeData.Outputs); err != nil {
			return nil, err
		}
		// Validate the question classifier node data
		if err := nodeData.Validate(); err != nil {
			return nil, fmt.Errorf("question classifier node validation failed: %w", err)
		}
		return nodeData, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}
}

// parseBaseNodeData parses common base node data fields
func (f *NodeFactory) parseBaseNodeData(nodeMap map[string]interface{}, baseData *entities.BaseNodeData) error {
	// Parse ID
	if idStr, ok := nodeMap["id"].(string); ok {
		if id, err := uuid.Parse(idStr); err == nil {
			baseData.ID = id
		} else {
			return fmt.Errorf("invalid node id format: %w", err)
		}
	}

	// Parse title
	if title, ok := nodeMap["title"].(string); ok {
		baseData.Title = title
	}

	// Parse node type
	if nodeTypeStr, ok := nodeMap["node_type"].(string); ok {
		baseData.NodeType = entities.NodeType(nodeTypeStr)
	}

	return nil
}

// parseInputsOutputs parses inputs and outputs for nodes that have them
func (f *NodeFactory) parseInputsOutputs(nodeMap map[string]interface{}, inputs *[]*entities.VariableEntity, outputs *[]*entities.VariableEntity) error {
	// Parse inputs
	if inputsData, exists := nodeMap["inputs"]; exists {
		if inputsSlice, ok := inputsData.([]interface{}); ok {
			for _, input := range inputsSlice {
				if inputMap, ok := input.(map[string]interface{}); ok {
					variable, err := f.parseVariableEntity(inputMap)
					if err != nil {
						return fmt.Errorf("failed to parse input variable: %w", err)
					}
					*inputs = append(*inputs, variable)
				}
			}
		}
	}

	// Parse outputs
	if outputsData, exists := nodeMap["outputs"]; exists {
		if outputsSlice, ok := outputsData.([]interface{}); ok {
			for _, output := range outputsSlice {
				if outputMap, ok := output.(map[string]interface{}); ok {
					variable, err := f.parseVariableEntity(outputMap)
					if err != nil {
						return fmt.Errorf("failed to parse output variable: %w", err)
					}
					*outputs = append(*outputs, variable)
				}
			}
		}
	}

	return nil
}

// parseVariableEntity parses a variable entity from a map
func (f *NodeFactory) parseVariableEntity(variableMap map[string]interface{}) (*entities.VariableEntity, error) {
	variable := entities.NewVariableEntity()

	// Parse name
	if name, ok := variableMap["name"].(string); ok {
		variable.Name = name
	}

	// Parse description
	if desc, ok := variableMap["description"].(string); ok {
		variable.Description = desc
	}

	// Parse required
	if required, ok := variableMap["required"].(bool); ok {
		variable.Required = required
	}

	// Parse type
	if typeStr, ok := variableMap["type"].(string); ok {
		variable.Type = entities.VariableType(typeStr)
	}

	// Parse value
	if valueData, exists := variableMap["value"]; exists {
		if valueMap, ok := valueData.(map[string]interface{}); ok {
			// Parse value type
			if valueType, ok := valueMap["type"].(string); ok {
				variable.Value.Type = entities.VariableValueType(valueType)
			}

			// Parse content based on type
			if variable.Value.Type == entities.VariableValueTypeRef {
				// Parse reference content
				if contentData, exists := valueMap["content"]; exists {
					if contentMap, ok := contentData.(map[string]interface{}); ok {
						content := &entities.VariableContent{}

						if refNodeIDStr, ok := contentMap["ref_node_id"].(string); ok {
							if refNodeID, err := uuid.Parse(refNodeIDStr); err == nil {
								content.RefNodeID = &refNodeID
							}
						}

						if refVarName, ok := contentMap["ref_var_name"].(string); ok {
							content.RefVarName = refVarName
						}

						variable.Value.Content = content
					}
				}
			} else {
				// Parse constant content
				if content, exists := valueMap["content"]; exists {
					variable.Value.Content = content
				}
			}
		}
	}

	return variable, nil
}
