package entities

import (
	"github.com/google/uuid"
)

// BaseNodeData represents the base data structure for all workflow nodes
type BaseNodeData struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	NodeType NodeType  `json:"node_type"`
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (b *BaseNodeData) GetBaseNodeData() *BaseNodeData {
	return b
}

// BaseEdgeData represents the base data structure for workflow edges
type BaseEdgeData struct {
	ID             uuid.UUID `json:"id"`
	Source         uuid.UUID `json:"source"`
	Target         uuid.UUID `json:"target"`
	SourceType     NodeType  `json:"source_type"`
	TargetType     NodeType  `json:"target_type"`
	SourceHandleID *string   `json:"source_handle_id,omitempty"`
}
