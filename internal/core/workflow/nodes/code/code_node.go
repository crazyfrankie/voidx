package code

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// CodeNode represents a code execution workflow node
type CodeNode struct {
	nodeData *CodeNodeData
}

// NewCodeNode creates a new code node instance
func NewCodeNode(nodeData *CodeNodeData) *CodeNode {
	return &CodeNode{
		nodeData: nodeData,
	}
}

// Execute executes the code node with the given workflow state
func (n *CodeNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
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

	// Execute Python code
	codeResult, err := n.executePythonCode(ctx, inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("code execution failed: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Build output data structure
	outputs := make(map[string]interface{})
	if len(n.nodeData.Outputs) > 0 {
		outputs[n.nodeData.Outputs[0].Name] = codeResult
	} else {
		outputs["result"] = codeResult
	}
	result.Outputs = outputs

	// Set success status
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// extractVariablesFromState extracts input variables from the workflow state
func (n *CodeNode) extractVariablesFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
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

// executePythonCode executes the Python code with the given variables
func (n *CodeNode) executePythonCode(ctx context.Context, variables map[string]interface{}) (interface{}, error) {
	// For security reasons, we'll implement a simple mock execution
	// In a production environment, you would want to use a sandboxed Python execution environment

	// Create a simple Python script that includes the variables and code
	pythonScript := n.buildPythonScript(variables)

	// Execute Python script (this is a simplified implementation)
	// In production, you should use a proper sandboxed execution environment
	cmd := exec.CommandContext(ctx, "python3", "-c", pythonScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("python execution failed: %w", err)
	}

	// Parse the output as JSON to get the result
	var result interface{}
	if err := sonic.Unmarshal(output, &result); err != nil {
		// If JSON parsing fails, return the raw output as string
		return strings.TrimSpace(string(output)), nil
	}

	return result, nil
}

// buildPythonScript builds a Python script with variables and user code
func (n *CodeNode) buildPythonScript(variables map[string]interface{}) string {
	var script strings.Builder

	// Add imports
	script.WriteString("import json\n")
	script.WriteString("import sys\n\n")

	// Add variables
	for key, value := range variables {
		valueJSON, _ := sonic.Marshal(value)
		script.WriteString(fmt.Sprintf("%s = json.loads('%s')\n", key, string(valueJSON)))
	}

	script.WriteString("\n")

	// Add user code
	script.WriteString(n.nodeData.Code)
	script.WriteString("\n\n")

	// Add result output
	script.WriteString("if 'result' in locals():\n")
	script.WriteString("    print(json.dumps(result))\n")
	script.WriteString("else:\n")
	script.WriteString("    print(json.dumps(None))\n")

	return script.String()
}

// GetNodeData returns the node data
func (n *CodeNode) GetNodeData() *CodeNodeData {
	return n.nodeData
}
