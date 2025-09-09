package iteration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/utils"
)

// IterationNode represents an iteration workflow node
type IterationNode struct {
	nodeData *IterationNodeData
	workflow WorkflowExecutor // Interface for executing sub-workflows
}

// WorkflowExecutor defines the interface for executing workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)
	GetInputSchema() map[string]any
}

// NewIterationNode creates a new iteration node instance
func NewIterationNode(nodeData *IterationNodeData) *IterationNode {
	return &IterationNode{
		nodeData: nodeData,
		workflow: nil,
	}
}

// SetWorkflow sets the sub-workflow executor for this iteration node
func (n *IterationNode) SetWorkflow(workflow WorkflowExecutor) {
	n.workflow = workflow
}

// Execute executes the iteration node
func (n *IterationNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Extract input variables from state
	inputsDict, err := utils.ExtractVariablesFromState(n.nodeData.Inputs, state)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to extract input variables: %v", err)
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	result.Inputs = inputsDict

	// Get the inputs array
	inputs, ok := inputsDict["inputs"]
	if !ok {
		result.Status = entities.NodeStatusFailed
		result.Error = "inputs not found in input variables"
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Convert inputs to slice
	inputsSlice, ok := inputs.([]any)
	if !ok {
		result.Status = entities.NodeStatusFailed
		result.Error = "inputs must be an array"
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Validate workflow and inputs
	if n.workflow == nil || len(inputsSlice) == 0 {
		result.Status = entities.NodeStatusFailed
		result.Error = "workflow not available or inputs empty"
		result.Outputs = map[string]any{"outputs": []any{}}
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Get workflow input schema to determine parameter key
	inputSchema := n.workflow.GetInputSchema()
	if len(inputSchema) != 1 {
		result.Status = entities.NodeStatusFailed
		result.Error = "workflow must have exactly one input parameter"
		result.Outputs = map[string]any{"outputs": []any{}}
		result.EndTime = time.Now().Unix()
		return result, nil
	}

	// Get the parameter key (first key in schema)
	var paramKey string
	for key := range inputSchema {
		paramKey = key
		break
	}

	// Execute workflow for each input item
	outputs := make([]string, 0, len(inputsSlice))
	for _, item := range inputsSlice {
		// Build input data for workflow
		workflowInput := map[string]any{
			paramKey: item,
		}

		// Execute the workflow
		iterationResult, err := n.workflow.Execute(ctx, workflowInput)
		if err != nil {
			// On error, add error message to outputs
			outputs = append(outputs, fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			continue
		}

		// Convert result to JSON string
		resultJSON, err := json.Marshal(iterationResult)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf(`{"error": "failed to marshal result: %s"}`, err.Error()))
			continue
		}

		outputs = append(outputs, string(resultJSON))
	}

	// Set successful result
	result.Status = entities.NodeStatusSucceeded
	result.Outputs = map[string]any{"outputs": outputs}
	result.EndTime = time.Now().Unix()

	return result, nil
}
