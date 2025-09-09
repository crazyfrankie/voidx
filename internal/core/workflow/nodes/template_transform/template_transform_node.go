package template_transform

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// TemplateTransformNode represents a template transformation workflow node
type TemplateTransformNode struct {
	nodeData *TemplateTransformNodeData
}

// NewTemplateTransformNode creates a new template transform node instance
func NewTemplateTransformNode(nodeData *TemplateTransformNodeData) *TemplateTransformNode {
	return &TemplateTransformNode{
		nodeData: nodeData,
	}
}

// Execute executes the template transform node with the given workflow state
func (n *TemplateTransformNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
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

	// Render template using extracted variables
	transformedValue, err := n.renderTemplate(inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to render template: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Build output data structure
	outputs := make(map[string]interface{})
	if len(n.nodeData.Outputs) > 0 {
		outputs[n.nodeData.Outputs[0].Name] = transformedValue
	} else {
		outputs["output"] = transformedValue
	}
	result.Outputs = outputs

	// Set success status
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// extractVariablesFromState extracts input variables from the workflow state
func (n *TemplateTransformNode) extractVariablesFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
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

// renderTemplate renders the template with the given variables
func (n *TemplateTransformNode) renderTemplate(variables map[string]interface{}) (string, error) {
	// Use Go's text/template to render the template
	tmpl, err := template.New("transform").Parse(n.nodeData.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// GetNodeData returns the node data
func (n *TemplateTransformNode) GetNodeData() *TemplateTransformNodeData {
	return n.nodeData
}
