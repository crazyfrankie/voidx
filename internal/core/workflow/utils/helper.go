package utils

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// ExtractVariablesFromState extracts variables from workflow state based on variable definitions
func ExtractVariablesFromState(variables []*entities.VariableEntity, state *entities.WorkflowState) (map[string]any, error) {
	result := make(map[string]any)

	for _, variable := range variables {
		value, err := extractVariableValue(variable, state)
		if err != nil {
			if variable.Required {
				return nil, fmt.Errorf("failed to extract required variable %s: %w", variable.Name, err)
			}
			// For optional variables, use default value or skip
			continue
		}
		result[variable.Name] = value
	}

	return result, nil
}

// extractVariableValue extracts a single variable value from the workflow state
func extractVariableValue(variable *entities.VariableEntity, state *entities.WorkflowState) (any, error) {
	switch variable.Value.Type {
	case entities.VariableValueTypeConstant:
		// Return the constant value directly
		return variable.Value.Content, nil

	case entities.VariableValueTypeRef:
		// Extract value from referenced node
		content, ok := variable.Value.Content.(*entities.VariableContent)
		if !ok {
			return nil, fmt.Errorf("invalid reference content for variable %s", variable.Name)
		}

		if content.RefNodeID == nil {
			return nil, fmt.Errorf("reference node ID is nil for variable %s", variable.Name)
		}

		// Find the referenced node result
		var referencedResult *entities.NodeResult
		for i := range state.NodeResults {
			if state.NodeResults[i].NodeID == *content.RefNodeID {
				referencedResult = &state.NodeResults[i]
				break
			}
		}

		if referencedResult == nil {
			return nil, fmt.Errorf("referenced node result not found for variable %s", variable.Name)
		}

		// Extract the specific variable from the referenced node's outputs
		if content.RefVarName == "" {
			return nil, fmt.Errorf("reference variable name is empty for variable %s", variable.Name)
		}

		value, exists := referencedResult.Outputs[content.RefVarName]
		if !exists {
			return nil, fmt.Errorf("referenced variable %s not found in node outputs", content.RefVarName)
		}

		return value, nil

	default:
		return nil, fmt.Errorf("unsupported variable value type: %s", variable.Value.Type)
	}
}

// FindNodeResultByID finds a node result by node ID in the workflow state
func FindNodeResultByID(state *entities.WorkflowState, nodeID uuid.UUID) *entities.NodeResult {
	for i := range state.NodeResults {
		if state.NodeResults[i].NodeID == nodeID {
			return &state.NodeResults[i]
		}
	}
	return nil
}

// GetNodeOutputVariable gets a specific output variable from a node result
func GetNodeOutputVariable(result *entities.NodeResult, variableName string) (any, bool) {
	if result == nil || result.Outputs == nil {
		return nil, false
	}

	value, exists := result.Outputs[variableName]
	return value, exists
}

// ValidateVariableType validates if a value matches the expected variable type
func ValidateVariableType(value any, expectedType entities.VariableType) error {
	switch expectedType {
	case entities.VariableTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case entities.VariableTypeNumber:
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid number types
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case entities.VariableTypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case entities.VariableTypeArray:
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case entities.VariableTypeObject:
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported variable type: %s", expectedType)
	}
	return nil
}
