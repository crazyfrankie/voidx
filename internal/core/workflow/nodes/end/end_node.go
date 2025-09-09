package end

import (
	"context"
	"fmt"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// EndNodeData represents the data structure for end workflow nodes
type EndNodeData struct {
	*entities.BaseNodeData
	Outputs []*entities.VariableEntity `json:"outputs"`
}

// NewEndNodeData creates a new end node data instance
func NewEndNodeData() *EndNodeData {
	return &EndNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeEnd},
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (e *EndNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return e.BaseNodeData
}

// EndNode represents an end workflow node
type EndNode struct {
	nodeData *EndNodeData
}

// NewEndNode creates a new end node instance
func NewEndNode(nodeData *EndNodeData) *EndNode {
	return &EndNode{
		nodeData: nodeData,
	}
}

// Execute executes the end node with the given workflow state
func (n *EndNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Extract output variables from state
	outputsDict, err := n.extractOutputsFromState(state)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to extract output variables: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	result.Inputs = outputsDict
	result.Outputs = outputsDict
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	// Update workflow state outputs
	state.Outputs = outputsDict

	return result, nil
}

// extractOutputsFromState extracts output variables from the workflow state
func (n *EndNode) extractOutputsFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
	outputsDict := make(map[string]interface{})

	for _, output := range n.nodeData.Outputs {
		var value interface{}
		var found bool

		// Check if it's a reference to another node's output
		if output.Value.Type == entities.VariableValueTypeRef {
			if content, ok := output.Value.Content.(*entities.VariableContent); ok {
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
			value = output.Value.Content
			found = true
		}

		if !found && output.Required {
			return nil, fmt.Errorf("required output variable %s not found", output.Name)
		}

		if found {
			outputsDict[output.Name] = value
		}
	}

	return outputsDict, nil
}

// GetNodeData returns the node data
func (n *EndNode) GetNodeData() *EndNodeData {
	return n.nodeData
}
