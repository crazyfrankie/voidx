package entities

import (
	"regexp"

	"github.com/google/uuid"
)

// NodeType represents the type of a workflow node
type NodeType string

const (
	NodeTypeStart              NodeType = "start"
	NodeTypeEnd                NodeType = "end"
	NodeTypeLLM                NodeType = "llm"
	NodeTypeTool               NodeType = "tool"
	NodeTypeCode               NodeType = "code"
	NodeTypeDatasetRetrieval   NodeType = "dataset_retrieval"
	NodeTypeHTTPRequest        NodeType = "http_request"
	NodeTypeTemplateTransform  NodeType = "template_transform"
	NodeTypeQuestionClassifier NodeType = "question_classifier"
	NodeTypeIteration          NodeType = "iteration"
)

// NodeStatus represents the execution status of a node
type NodeStatus string

const (
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusSucceeded NodeStatus = "succeeded"
	NodeStatusFailed    NodeStatus = "failed"
)

// WorkflowConfig represents the configuration of a workflow
type WorkflowConfig struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Nodes       []*BaseNodeData `json:"nodes"`
	Edges       []*BaseEdgeData `json:"edges"`
	AccountID   uuid.UUID       `json:"account_id,omitempty"`
}

// NewWorkflowConfig creates a new workflow configuration
func NewWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		Nodes: make([]*BaseNodeData, 0),
		Edges: make([]*BaseEdgeData, 0),
	}
}

// WorkflowState represents the state during workflow execution
type WorkflowState struct {
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	NodeResults []NodeResult           `json:"node_results"`
}

// NewWorkflowState creates a new workflow state
func NewWorkflowState() *WorkflowState {
	return &WorkflowState{
		Inputs:      make(map[string]interface{}),
		Outputs:     make(map[string]interface{}),
		NodeResults: make([]NodeResult, 0),
	}
}

// SetInputsFromMap sets inputs from a map
func (ws *WorkflowState) SetInputsFromMap(inputs map[string]interface{}) error {
	ws.Inputs = inputs
	return nil
}

// GetInputsAsMap returns inputs as a map
func (ws *WorkflowState) GetInputsAsMap() (map[string]interface{}, error) {
	return ws.Inputs, nil
}

// SetOutputsFromMap sets outputs from a map
func (ws *WorkflowState) SetOutputsFromMap(outputs map[string]interface{}) error {
	ws.Outputs = outputs
	return nil
}

// GetOutputsAsMap returns outputs as a map
func (ws *WorkflowState) GetOutputsAsMap() (map[string]interface{}, error) {
	return ws.Outputs, nil
}

// NodeResult represents the result of a node execution
type NodeResult struct {
	NodeID    uuid.UUID              `json:"node_id"`
	NodeType  NodeType               `json:"node_type"`
	Status    NodeStatus             `json:"status"`
	Inputs    map[string]interface{} `json:"inputs,omitempty"`
	Outputs   map[string]interface{} `json:"outputs,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartTime int64                  `json:"start_time,omitempty"`
	EndTime   int64                  `json:"end_time,omitempty"`
}

// NewNodeResult creates a new node result
func NewNodeResult(node *BaseNodeData) *NodeResult {
	return &NodeResult{
		NodeID:   node.ID,
		NodeType: node.NodeType,
		Status:   NodeStatusRunning,
		Inputs:   make(map[string]interface{}),
		Outputs:  make(map[string]interface{}),
	}
}

// Constants for workflow configuration validation
const (
	WorkflowConfigDescriptionMaxLength = 1024
)

// WorkflowConfigNamePattern is the regex pattern for workflow names
var WorkflowConfigNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// NodeDataInterface represents the interface for all node data types
type NodeDataInterface interface {
	GetBaseNodeData() *BaseNodeData
}
