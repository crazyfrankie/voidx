package start

import (
	"context"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// StartNodeData represents the data structure for start workflow nodes
type StartNodeData struct {
	*entities.BaseNodeData
	Inputs []*entities.VariableEntity `json:"inputs"`
}

// NewStartNodeData creates a new start node data instance
func NewStartNodeData() *StartNodeData {
	return &StartNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeStart},
		Inputs:       make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (s *StartNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return s.BaseNodeData
}

// StartNode represents a start workflow node
type StartNode struct {
	nodeData *StartNodeData
}

// NewStartNode creates a new start node instance
func NewStartNode(nodeData *StartNodeData) *StartNode {
	return &StartNode{
		nodeData: nodeData,
	}
}

// Execute executes the start node with the given workflow state
func (n *StartNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Start node simply passes through the workflow inputs
	result.Inputs = state.Inputs
	result.Outputs = state.Inputs
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// GetNodeData returns the node data
func (n *StartNode) GetNodeData() *StartNodeData {
	return n.nodeData
}
