package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/start"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// Workflow represents a workflow implementation using eino's compose framework
type Workflow struct {
	workflowConfig *entities.WorkflowConfig
	runnable       compose.Runnable[map[string]interface{}, map[string]interface{}]
	nodeFactory    *nodes.NodeFactory
	accountID      uuid.UUID
	nodeExecutors  map[uuid.UUID]nodes.NodeExecutor
}

// SetNodeFactory sets the node factory for the workflow
func (w *Workflow) SetNodeFactory(factory *nodes.NodeFactory) {
	w.nodeFactory = factory
}

// SetAccountID sets the account ID for the workflow
func (w *Workflow) SetAccountID(accountID uuid.UUID) {
	w.accountID = accountID
}

// RebuildWorkflowGraph rebuilds the workflow graph after nodeFactory is set
func (w *Workflow) RebuildWorkflowGraph() error {
	return w.buildWorkflowGraph()
}

// NewWorkflow creates a new workflow instance
func NewWorkflow(values map[string]interface{}) (*Workflow, error) {
	wf := &Workflow{
		workflowConfig: entities.NewWorkflowConfig(),
		// nodeFactory and accountID will be set by WorkflowManager
	}
	if err := wf.ValidateWorkflowConfig(values); err != nil {
		return nil, err
	}

	// Build the workflow graph using eino's compose framework
	if err := wf.buildWorkflowGraph(); err != nil {
		return nil, fmt.Errorf("failed to build workflow graph: %w", err)
	}

	return wf, nil
}

// Name returns the workflow name
func (w *Workflow) Name() string {
	return w.workflowConfig.Name
}

// Description returns the workflow description
func (w *Workflow) Description() string {
	return w.workflowConfig.Description
}

// Info returns tool information for the workflow (implements tool.BaseTool)
func (w *Workflow) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        w.workflowConfig.Name,
		Desc:        w.workflowConfig.Description,
		ParamsOneOf: w.buildParametersSchema(),
	}, nil
}

// InvokableRun executes the workflow (implements tool.InvokableTool)
func (w *Workflow) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Parse input arguments
	var inputMap map[string]interface{}
	if argumentsInJSON != "" {
		if err := sonic.Unmarshal([]byte(argumentsInJSON), &inputMap); err != nil {
			return "", fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		inputMap = make(map[string]interface{})
	}

	// Execute workflow using the compiled runnable
	result, err := w.runnable.Invoke(ctx, inputMap)
	if err != nil {
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}

	// Serialize result
	outputBytes, err := sonic.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outputs: %w", err)
	}

	return string(outputBytes), nil
}

// Stream executes the workflow with streaming output
func (w *Workflow) Stream(ctx context.Context, input map[string]interface{}) (<-chan *entities.NodeResult, error) {
	// Create result channel
	resultChan := make(chan *entities.NodeResult, len(w.workflowConfig.Nodes))

	// Start workflow execution in background
	go func() {
		defer close(resultChan)

		// Use eino's streaming capabilities
		streamReader, err := w.runnable.Stream(ctx, input)
		if err != nil {
			// Send error result
			errorResult := &entities.NodeResult{
				NodeID:   uuid.New(),
				NodeType: "error",
				Status:   entities.NodeStatusFailed,
				Error:    err.Error(),
			}
			select {
			case resultChan <- errorResult:
			case <-ctx.Done():
			}
			return
		}

		// Process streaming results
		for {
			result, err := streamReader.Recv()
			if err != nil {
				break
			}

			// Convert result to NodeResult
			nodeResult := &entities.NodeResult{
				NodeID:  uuid.New(),
				Status:  entities.NodeStatusSucceeded,
				Outputs: result,
			}

			select {
			case resultChan <- nodeResult:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}

// buildWorkflowGraph builds the workflow execution graph using eino's compose framework
func (w *Workflow) buildWorkflowGraph() error {
	// Skip building if nodeFactory is not set (will be built later by WorkflowManager)
	if w.nodeFactory == nil {
		return nil
	}

	// Parse node data and create node executors
	nodeExecutors := make(map[uuid.UUID]nodes.NodeExecutor)

	for _, baseNode := range w.workflowConfig.Nodes {
		// Create a map representation of the node for parsing
		nodeMap := map[string]interface{}{
			"id":        baseNode.ID.String(),
			"title":     baseNode.Title,
			"node_type": string(baseNode.NodeType),
		}

		// Parse the node data using the factory
		nodeData, err := w.nodeFactory.ParseNodeData(nodeMap)
		if err != nil {
			return fmt.Errorf("failed to parse node data for node %s: %w", baseNode.ID, err)
		}

		// Create the node executor
		executor, err := w.nodeFactory.CreateNode(nodeData, w.accountID)
		if err != nil {
			return fmt.Errorf("failed to create node executor for node %s: %w", baseNode.ID, err)
		}

		nodeExecutors[baseNode.ID] = executor
	}

	// Store node executors for later use
	w.nodeExecutors = nodeExecutors

	// Create a new workflow using eino's Workflow API
	workflow := compose.NewWorkflow[map[string]interface{}, map[string]interface{}]()

	// Build adjacency list for dependencies
	adjacencyList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range w.workflowConfig.Edges {
		adjacencyList[edge.Target] = append(adjacencyList[edge.Target], edge.Source)
	}

	// Create node map for ID to name mapping
	nodeMap := make(map[uuid.UUID]string)
	workflowNodes := make(map[uuid.UUID]*compose.WorkflowNode)

	// Add nodes to the workflow
	for _, node := range w.workflowConfig.Nodes {
		nodeID := fmt.Sprintf("%s_%s", node.NodeType, node.ID.String())
		nodeMap[node.ID] = nodeID

		// Capture node ID for closure
		capturedNodeID := node.ID

		// Create lambda node that executes the actual workflow node
		lambda := compose.InvokableLambda(func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			return w.executeNode(ctx, capturedNodeID, input)
		})

		// Add the lambda node to workflow
		workflowNode := workflow.AddLambdaNode(nodeID, lambda)
		workflowNodes[node.ID] = workflowNode
	}

	// Add dependencies between nodes based on edges
	for _, edge := range w.workflowConfig.Edges {
		sourceNodeID := nodeMap[edge.Source]

		// Handle end node specially
		if edge.TargetType == entities.NodeTypeEnd {
			workflow.End().AddInput(sourceNodeID)
		} else {
			// Add input dependency for regular nodes
			if targetNode, exists := workflowNodes[edge.Target]; exists {
				targetNode.AddInput(sourceNodeID)
			}
		}
	}

	// Compile the workflow
	compiledWorkflow, err := workflow.Compile(context.Background())
	if err != nil {
		return fmt.Errorf("failed to compile workflow: %w", err)
	}

	w.runnable = compiledWorkflow

	return nil
}

// executeNode executes a single workflow node using the stored node executors
func (w *Workflow) executeNode(ctx context.Context, nodeID uuid.UUID, input map[string]interface{}) (map[string]interface{}, error) {
	// Get the node executor
	executor, exists := w.nodeExecutors[nodeID]
	if !exists {
		return nil, fmt.Errorf("node executor not found for node %s", nodeID)
	}

	// Create workflow state from input
	state := entities.NewWorkflowState()
	state.Inputs = input

	// Execute the node
	result, err := executor.Execute(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("node execution failed: %w", err)
	}

	// Extract outputs from the result
	if result != nil {
		if result.Status == entities.NodeStatusSucceeded {
			return result.Outputs, nil
		} else {
			return nil, fmt.Errorf("node execution failed: %s", result.Error)
		}
	}

	// Fallback to input if no result
	return input, nil
}

// buildParametersSchema builds the parameters schema for the workflow tool
func (w *Workflow) buildParametersSchema() *schema.ParamsOneOf {
	// Find start node and extract its inputs
	for _, node := range w.workflowConfig.Nodes {
		if node.NodeType == entities.NodeTypeStart {
			// If we have nodeFactory, try to get the actual start node data
			if w.nodeFactory != nil {
				nodeMap := map[string]interface{}{
					"id":        node.ID.String(),
					"title":     node.Title,
					"node_type": string(node.NodeType),
				}

				if nodeData, err := w.nodeFactory.ParseNodeData(nodeMap); err == nil {
					// Try to cast to StartNodeData to extract inputs
					if startNodeData, ok := nodeData.(*start.StartNodeData); ok {
						params := make(map[string]*schema.ParameterInfo)

						// Convert each input variable to a parameter
						for _, input := range startNodeData.Inputs {
							var paramType schema.DataType
							switch input.Type {
							case entities.VariableTypeString:
								paramType = schema.String
							case entities.VariableTypeNumber:
								paramType = schema.Number
							case entities.VariableTypeBool:
								paramType = schema.Boolean
							case entities.VariableTypeArray:
								paramType = schema.Array
							case entities.VariableTypeObject:
								paramType = schema.Object
							default:
								paramType = schema.String
							}

							params[input.Name] = &schema.ParameterInfo{
								Type:     paramType,
								Desc:     input.Description,
								Required: input.Required,
							}
						}

						// If no inputs defined, use a default input parameter
						if len(params) == 0 {
							params["input"] = &schema.ParameterInfo{
								Type:     schema.String,
								Desc:     "Workflow input",
								Required: true,
							}
						}

						return schema.NewParamsOneOfByParams(params)
					}
				}
			}

			// Fallback to basic schema
			params := map[string]*schema.ParameterInfo{
				"input": {
					Type:     schema.String,
					Desc:     "Workflow input",
					Required: true,
				},
			}
			return schema.NewParamsOneOfByParams(params)
		}
	}

	// Return empty parameters if no start node found
	return schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{})
}

// ValidateWorkflowConfig validates the workflow configuration
func (w *Workflow) ValidateWorkflowConfig(values map[string]interface{}) error {
	wc := w.workflowConfig

	// Validate workflow name
	name, ok := values["name"].(string)
	if !ok || name == "" || !entities.WorkflowConfigNamePattern.MatchString(name) {
		return fmt.Errorf("工作流名字仅支持字母、数字和下划线，且以字母/下划线为开头")
	}
	wc.Name = name

	// Validate workflow description
	description, ok := values["description"].(string)
	if !ok || description == "" || len(description) > entities.WorkflowConfigDescriptionMaxLength {
		return fmt.Errorf("工作流描述不能为空且长度不能超过%d个字符", entities.WorkflowConfigDescriptionMaxLength)
	}
	wc.Description = description

	// Validate nodes and edges
	nodes, ok := values["nodes"].([]interface{})
	if !ok || len(nodes) == 0 {
		return fmt.Errorf("工作流节点列表信息错误，请核实后重试")
	}

	edges, ok := values["edges"].([]interface{})
	if !ok || len(edges) == 0 {
		return fmt.Errorf("工作流边列表信息错误，请核实后重试")
	}

	// Parse and validate nodes
	nodeDataDict := make(map[uuid.UUID]*entities.BaseNodeData)
	startNodes := 0
	endNodes := 0

	for _, nodeInterface := range nodes {
		nodeMap, ok := nodeInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("节点数据格式错误")
		}

		nodeData, err := parseNodeFromMap(nodeMap)
		if err != nil {
			return fmt.Errorf("解析节点数据失败: %w", err)
		}

		// Check for unique start and end nodes
		if nodeData.NodeType == entities.NodeTypeStart {
			if startNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个开始节点")
			}
			startNodes++
		} else if nodeData.NodeType == entities.NodeTypeEnd {
			if endNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个结束节点")
			}
			endNodes++
		}

		// Check for unique node IDs
		if _, exists := nodeDataDict[nodeData.ID]; exists {
			return fmt.Errorf("工作流节点id必须唯一，请核实后重试")
		}

		// Check for unique node titles
		for _, existingNode := range nodeDataDict {
			if strings.TrimSpace(existingNode.Title) == strings.TrimSpace(nodeData.Title) {
				return fmt.Errorf("工作流节点title必须唯一，请核实后重试")
			}
		}

		nodeDataDict[nodeData.ID] = nodeData
	}

	// Convert nodeDataDict to slice
	wc.Nodes = make([]*entities.BaseNodeData, 0, len(nodeDataDict))
	for _, nodeData := range nodeDataDict {
		wc.Nodes = append(wc.Nodes, nodeData)
	}

	// Parse and validate edges
	edgeDataDict := make(map[uuid.UUID]*entities.BaseEdgeData)
	for _, edgeInterface := range edges {
		edgeMap, ok := edgeInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("边数据格式错误")
		}

		edgeData, err := parseEdgeFromMap(edgeMap)
		if err != nil {
			return fmt.Errorf("解析边数据失败: %w", err)
		}

		// Check for unique edge IDs
		if _, exists := edgeDataDict[edgeData.ID]; exists {
			return fmt.Errorf("工作流边数据id必须唯一，请核实后重试")
		}

		// Validate edge source and target nodes exist
		if _, exists := nodeDataDict[edgeData.Source]; !exists {
			return fmt.Errorf("边的源节点不存在")
		}
		if _, exists := nodeDataDict[edgeData.Target]; !exists {
			return fmt.Errorf("边的目标节点不存在")
		}

		edgeDataDict[edgeData.ID] = edgeData
	}

	// Convert edgeDataDict to slice
	wc.Edges = make([]*entities.BaseEdgeData, 0, len(edgeDataDict))
	for _, edgeData := range edgeDataDict {
		wc.Edges = append(wc.Edges, edgeData)
	}

	return nil
}

// parseNodeFromMap parses a node from a map
func parseNodeFromMap(nodeMap map[string]interface{}) (*entities.BaseNodeData, error) {
	// Parse ID
	idStr, ok := nodeMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid node id format: %w", err)
	}

	// Parse title
	title, ok := nodeMap["title"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node title")
	}

	// Parse node type
	nodeTypeStr, ok := nodeMap["node_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	nodeType := entities.NodeType(nodeTypeStr)

	return &entities.BaseNodeData{
		ID:       id,
		Title:    title,
		NodeType: nodeType,
	}, nil
}

// parseEdgeFromMap parses an edge from a map
func parseEdgeFromMap(edgeMap map[string]interface{}) (*entities.BaseEdgeData, error) {
	// Parse ID
	idStr, ok := edgeMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge id format: %w", err)
	}

	// Parse source
	sourceStr, ok := edgeMap["source"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge source")
	}

	source, err := uuid.Parse(sourceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge source format: %w", err)
	}

	// Parse target
	targetStr, ok := edgeMap["target"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge target")
	}

	target, err := uuid.Parse(targetStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge target format: %w", err)
	}

	// Parse source and target types
	sourceTypeStr, ok := edgeMap["source_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge source_type")
	}

	targetTypeStr, ok := edgeMap["target_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge target_type")
	}

	edgeData := &entities.BaseEdgeData{
		ID:         id,
		Source:     source,
		Target:     target,
		SourceType: entities.NodeType(sourceTypeStr),
		TargetType: entities.NodeType(targetTypeStr),
	}

	// Parse optional source handle ID
	if sourceHandleID, exists := edgeMap["source_handle_id"]; exists && sourceHandleID != nil {
		if handleIDStr, ok := sourceHandleID.(string); ok {
			edgeData.SourceHandleID = &handleIDStr
		}
	}

	return edgeData, nil
}
