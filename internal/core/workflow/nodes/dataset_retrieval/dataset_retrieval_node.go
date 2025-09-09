package dataset_retrieval

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// DatasetRetrievalNode represents a dataset retrieval workflow node
type DatasetRetrievalNode struct {
	nodeData         *DatasetRetrievalNodeData
	retrieverService *retrievers.RetrieverService
	accountID        uuid.UUID
}

// NewDatasetRetrievalNode creates a new dataset retrieval node instance
func NewDatasetRetrievalNode(nodeData *DatasetRetrievalNodeData, retrieverService *retrievers.RetrieverService, accountID uuid.UUID) *DatasetRetrievalNode {
	return &DatasetRetrievalNode{
		nodeData:         nodeData,
		retrieverService: retrieverService,
		accountID:        accountID,
	}
}

// Execute executes the dataset retrieval node with the given workflow state
func (n *DatasetRetrievalNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
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

	// Extract query from inputs
	query, ok := inputsDict["query"].(string)
	if !ok {
		result.Status = entities.NodeStatusFailed
		result.Error = "query input is required and must be a string"
		result.EndTime = time.Now().Unix()
		return result, fmt.Errorf("query input is required and must be a string")
	}

	// Perform retrieval using the retriever service
	retrievalResults, err := n.performRetrieval(ctx, query)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("retrieval failed: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Build output data structure
	outputs := make(map[string]interface{})
	if len(n.nodeData.Outputs) > 0 {
		outputs[n.nodeData.Outputs[0].Name] = retrievalResults
	} else {
		outputs["documents"] = retrievalResults
	}
	result.Outputs = outputs

	// Set success status
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// extractVariablesFromState extracts input variables from the workflow state
func (n *DatasetRetrievalNode) extractVariablesFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
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

// performRetrieval performs the actual retrieval using the retriever service
func (n *DatasetRetrievalNode) performRetrieval(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// Get retrieval configuration
	topK := 5
	if topKValue, exists := n.nodeData.RetrievalConfig["top_k"]; exists {
		if topKInt, ok := topKValue.(int); ok {
			topK = topKInt
		}
	}

	retrievalMode := "hybrid"
	if modeValue, exists := n.nodeData.RetrievalConfig["retrieval_mode"]; exists {
		if modeStr, ok := modeValue.(string); ok {
			retrievalMode = modeStr
		}
	}

	// Use the first dataset ID for retrieval (in a real implementation, you might want to handle multiple datasets)
	if len(n.nodeData.DatasetIDs) == 0 {
		return nil, fmt.Errorf("no dataset IDs configured for retrieval")
	}

	// Create retriever configuration
	config := n.retrieverService.GetDefaultConfig(retrievers.RetrieverType(retrievalMode), n.nodeData.DatasetIDs[0])
	config.TopK = topK

	// Perform retrieval using the service
	documents, err := n.retrieverService.RetrieveDocuments(ctx, query, config)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// Convert documents to map format
	results := make([]map[string]interface{}, len(documents))
	for i, doc := range documents {
		results[i] = map[string]interface{}{
			"content":  doc.Content,
			"metadata": doc.MetaData,
		}
	}

	return results, nil
}

// GetNodeData returns the node data
func (n *DatasetRetrievalNode) GetNodeData() *DatasetRetrievalNodeData {
	return n.nodeData
}
