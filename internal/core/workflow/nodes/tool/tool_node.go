package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// ToolNodeData represents the data structure for tool workflow nodes
type ToolNodeData struct {
	*entities.BaseNodeData
	ToolName   string                     `json:"tool_name"`
	ToolConfig map[string]interface{}     `json:"tool_config"`
	Inputs     []*entities.VariableEntity `json:"inputs"`
	Outputs    []*entities.VariableEntity `json:"outputs"`
}

// NewToolNodeData creates a new tool node data instance
func NewToolNodeData() *ToolNodeData {
	return &ToolNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeTool},
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
		ToolConfig:   make(map[string]interface{}),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (t *ToolNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return t.BaseNodeData
}

// ToolNode represents a tool workflow node
type ToolNode struct {
	nodeData *ToolNodeData
	tool     tool.InvokableTool
}

// NewToolNode creates a new tool node instance
func NewToolNode(nodeData *ToolNodeData, tool tool.InvokableTool) *ToolNode {
	return &ToolNode{
		nodeData: nodeData,
		tool:     tool,
	}
}

// Execute executes the tool node with the given workflow state
func (n *ToolNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Extract input variables from state
	inputsDict, err := n.extractVariablesFromState(state)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to extract input variables: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}
	result.Inputs = inputsDict

	// Prepare tool arguments
	toolArgs, err := sonic.Marshal(inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to marshal tool arguments: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Execute the tool
	toolResult, err := n.tool.InvokableRun(ctx, string(toolArgs))
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("tool execution failed: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Build output data structure
	outputs := make(map[string]interface{})
	if len(n.nodeData.Outputs) > 0 {
		outputs[n.nodeData.Outputs[0].Name] = toolResult
	} else {
		outputs["output"] = toolResult
	}
	result.Outputs = outputs

	// Set success status
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// extractVariablesFromState extracts input variables from the workflow state
func (n *ToolNode) extractVariablesFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
	inputsDict := make(map[string]interface{})

	for _, input := range n.nodeData.Inputs {
		var value interface{}
		var found bool

		// Check if it's a reference to another node's output
		if input.Value.Type == entities.VariableValueTypeRef {
			if content, ok := input.Value.Content.(*entities.VariableContent); ok {
				if content.RefNodeID != nil {
					// Find the referenced node's output in state
					for _, nodeResult := range state.NodeResults {
						if nodeResult.NodeID == *content.RefNodeID {
							if refValue, exists := nodeResult.Outputs[content.RefVarName]; exists {
								value = refValue
								found = true
								break
							}
						}
					}
				}
			}
		} else {
			// It's a constant value
			value = input.Value.Content
			found = true
		}

		if !found && input.Required {
			return nil, fmt.Errorf("required input variable %s not found", input.Name)
		}

		if found {
			inputsDict[input.Name] = value
		}
	}

	// Also include workflow inputs
	for key, value := range state.Inputs {
		if _, exists := inputsDict[key]; !exists {
			inputsDict[key] = value
		}
	}

	return inputsDict, nil
}

// GetNodeData returns the node data
func (n *ToolNode) GetNodeData() *ToolNodeData {
	return n.nodeData
}
