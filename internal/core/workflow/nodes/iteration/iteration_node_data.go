package iteration

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// IterationNodeData represents the data structure for iteration nodes
type IterationNodeData struct {
	*entities.BaseNodeData
	WorkflowIDs []uuid.UUID                `json:"workflow_ids"`
	Inputs      []*entities.VariableEntity `json:"inputs"`
	Outputs     []*entities.VariableEntity `json:"outputs"`
}

// NewIterationNodeData creates a new iteration node data instance
func NewIterationNodeData() *IterationNodeData {
	return &IterationNodeData{
		BaseNodeData: &entities.BaseNodeData{
			NodeType: entities.NodeTypeIteration,
		},
		WorkflowIDs: make([]uuid.UUID, 0),
		Inputs: []*entities.VariableEntity{
			{
				Name:     "inputs",
				Type:     entities.VariableTypeArray,
				Required: true,
				Value: entities.VariableValue{
					Type:    entities.VariableValueTypeConstant,
					Content: []any{},
				},
			},
		},
		Outputs: []*entities.VariableEntity{
			{
				Name: "outputs",
				Value: entities.VariableValue{
					Type: entities.VariableValueTypeConstant,
				},
			},
		},
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (d *IterationNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return d.BaseNodeData
}

// Validate validates the iteration node data
func (d *IterationNodeData) Validate() error {
	// Validate workflow IDs - only one workflow is allowed
	if len(d.WorkflowIDs) > 1 {
		return fmt.Errorf("iteration node can only bind to one workflow")
	}

	// Validate inputs - must have exactly one input variable
	if len(d.Inputs) != 1 {
		return fmt.Errorf("iteration node input variable information error")
	}

	// Validate input variable properties
	iterationInput := d.Inputs[0]
	allowedTypes := []entities.VariableType{
		entities.VariableTypeArray,
	}

	typeAllowed := false
	for _, allowedType := range allowedTypes {
		if iterationInput.Type == allowedType {
			typeAllowed = true
			break
		}
	}

	if iterationInput.Name != "inputs" || !typeAllowed || !iterationInput.Required {
		return fmt.Errorf("iteration node input variable name/type/required property error")
	}

	return nil
}
