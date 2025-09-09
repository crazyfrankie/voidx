package question_classifier

import (
	"fmt"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// ClassConfig represents the configuration for a question classifier class
type ClassConfig struct {
	Query          string `json:"query"`            // Question classification query description
	NodeID         string `json:"node_id"`          // Connected node ID
	NodeType       string `json:"node_type"`        // Connected node type
	SourceHandleID string `json:"source_handle_id"` // Source handle ID
}

// QuestionClassifierNodeData represents the data structure for question classifier nodes
type QuestionClassifierNodeData struct {
	*entities.BaseNodeData
	Inputs  []*entities.VariableEntity `json:"inputs"`
	Outputs []*entities.VariableEntity `json:"outputs"`
	Classes []*ClassConfig             `json:"classes"`
}

// NewQuestionClassifierNodeData creates a new question classifier node data instance
func NewQuestionClassifierNodeData() *QuestionClassifierNodeData {
	return &QuestionClassifierNodeData{
		BaseNodeData: &entities.BaseNodeData{
			NodeType: entities.NodeTypeQuestionClassifier,
		},
		Inputs: []*entities.VariableEntity{
			{
				Name:     "query",
				Type:     entities.VariableTypeString,
				Required: true,
				Value: entities.VariableValue{
					Type: entities.VariableValueTypeConstant,
				},
			},
		},
		Outputs: []*entities.VariableEntity{}, // Question classifier has no outputs
		Classes: make([]*ClassConfig, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (d *QuestionClassifierNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return d.BaseNodeData
}

// Validate validates the question classifier node data
func (d *QuestionClassifierNodeData) Validate() error {
	// Validate inputs - must have exactly one input variable
	if len(d.Inputs) != 1 {
		return fmt.Errorf("question classifier node input variable information error")
	}

	// Validate input variable properties
	queryInput := d.Inputs[0]
	if queryInput.Name != "query" || !queryInput.Required {
		return fmt.Errorf("question classifier node input variable name/type/required property error")
	}

	// Validate classes
	if len(d.Classes) == 0 {
		return fmt.Errorf("question classifier node must have at least one class")
	}

	for i, class := range d.Classes {
		if class.Query == "" {
			return fmt.Errorf("class %d query cannot be empty", i)
		}
		if class.SourceHandleID == "" {
			return fmt.Errorf("class %d source handle ID cannot be empty", i)
		}
	}

	return nil
}

// GetClassNames returns all class names for this question classifier
func (d *QuestionClassifierNodeData) GetClassNames() []string {
	classNames := make([]string, len(d.Classes))
	for i, class := range d.Classes {
		classNames[i] = fmt.Sprintf("qc_source_handle_%s", class.SourceHandleID)
	}
	return classNames
}

// GetClassBySourceHandleID finds a class by its source handle ID
func (d *QuestionClassifierNodeData) GetClassBySourceHandleID(sourceHandleID string) *ClassConfig {
	for _, class := range d.Classes {
		if class.SourceHandleID == sourceHandleID {
			return class
		}
	}
	return nil
}
