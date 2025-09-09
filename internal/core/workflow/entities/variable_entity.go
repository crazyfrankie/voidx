package entities

import (
	"github.com/google/uuid"
)

// VariableType represents the type of a variable
type VariableType string

const (
	VariableTypeString VariableType = "string"
	VariableTypeNumber VariableType = "number"
	VariableTypeBool   VariableType = "bool"
	VariableTypeArray  VariableType = "array"
	VariableTypeObject VariableType = "object"
)

// VariableValueType represents the type of a variable value
type VariableValueType string

const (
	VariableValueTypeConstant VariableValueType = "constant"
	VariableValueTypeRef      VariableValueType = "ref"
)

// VariableEntity represents a variable in the workflow
type VariableEntity struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Type        VariableType  `json:"type"`
	Value       VariableValue `json:"value"`
}

// VariableValue represents the value of a variable
type VariableValue struct {
	Type    VariableValueType `json:"type"`
	Content interface{}       `json:"content"`
}

// VariableContent represents the content of a variable reference
type VariableContent struct {
	RefNodeID  *uuid.UUID `json:"ref_node_id,omitempty"`
	RefVarName string     `json:"ref_var_name,omitempty"`
}

// NewVariableEntity creates a new variable entity
func NewVariableEntity() *VariableEntity {
	return &VariableEntity{
		Type: VariableTypeString,
		Value: VariableValue{
			Type: VariableValueTypeConstant,
		},
	}
}

// VARIABLE_TYPE_MAP maps variable types to Go types
var VARIABLE_TYPE_MAP = map[VariableType]interface{}{
	VariableTypeString: "",
	VariableTypeNumber: 0.0,
	VariableTypeBool:   false,
	VariableTypeArray:  []interface{}{},
	VariableTypeObject: map[string]interface{}{},
}
