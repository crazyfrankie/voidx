package template_transform

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// TemplateTransformNodeData represents the data structure for template transform workflow nodes
type TemplateTransformNodeData struct {
	*entities.BaseNodeData
	Template string                     `json:"template"`
	Inputs   []*entities.VariableEntity `json:"inputs"`
	Outputs  []*entities.VariableEntity `json:"outputs"`
}

// NewTemplateTransformNodeData creates a new template transform node data instance
func NewTemplateTransformNodeData() *TemplateTransformNodeData {
	return &TemplateTransformNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeTemplateTransform},
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (t *TemplateTransformNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return t.BaseNodeData
}
