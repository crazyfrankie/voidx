package nodes

import (
	"context"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// NodeExecutor defines the interface for all workflow nodes
type NodeExecutor interface {
	Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error)
}

// BaseNode provides common functionality for all workflow nodes
type BaseNode struct {
	nodeData *entities.BaseNodeData
}

// NewBaseNode creates a new base node instance
func NewBaseNode(nodeData *entities.BaseNodeData) *BaseNode {
	return &BaseNode{
		nodeData: nodeData,
	}
}

// GetNodeData returns the node data
func (n *BaseNode) GetNodeData() *entities.BaseNodeData {
	return n.nodeData
}
