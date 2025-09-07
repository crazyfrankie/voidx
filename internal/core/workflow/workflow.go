package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/code"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/dataset_retrieval"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/end"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/http_request"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/iteration"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/llm"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/question_classifier"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/start"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/template_transform"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes/tool"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// Workflow 工作流LangChain工具类
type Workflow struct {
	workflowConfig *entities.WorkflowConfig
}

// NewWorkflow 构造函数，完成工作流函数的初始化
func NewWorkflow(values map[string]any) (*Workflow, error) {
	wf := &Workflow{workflowConfig: entities.NewWorkflowConfig()}
	if err := wf.ValidateWorkflowConfig(values); err != nil {
		return nil, err
	}

	return wf, nil
}

// Name 返回工作流名称
func (w *Workflow) Name() string {
	return w.workflowConfig.Name
}

// Description 返回工作流描述
func (w *Workflow) Description() string {
	return w.workflowConfig.Description
}

// Call 执行工作流
func (w *Workflow) Call(ctx context.Context, input string) (string, error) {
	// 解析输入参数
	var inputMap map[string]any
	if input != "" {
		if err := sonic.Unmarshal([]byte(input), &inputMap); err != nil {
			return "", fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		inputMap = make(map[string]any)
	}

	// 创建工作流状态
	state := &entities.WorkflowState{}
	if err := state.SetInputsFromMap(inputMap); err != nil {
		return "", fmt.Errorf("failed to set inputs: %w", err)
	}

	// 执行工作流
	result, err := w.executeWorkflow(ctx, state)
	if err != nil {
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}

	// 获取输出结果
	outputs, err := result.GetOutputsAsMap()
	if err != nil {
		return "", fmt.Errorf("failed to get outputs: %w", err)
	}

	// 将输出结果序列化为字符串
	outputBytes, err := sonic.Marshal(outputs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outputs: %w", err)
	}

	return string(outputBytes), nil
}

// executeWorkflow 执行工作流的核心逻辑
func (w *Workflow) executeWorkflow(ctx context.Context, state *entities.WorkflowState) (*entities.WorkflowState, error) {
	// 这里实现工作流的执行逻辑
	// 由于Go版本中没有具体的节点实现，这里提供一个基础框架

	// 1. 找到开始节点
	var startNode *entities.BaseNodeData
	for _, node := range w.workflowConfig.Nodes {
		if node.NodeType == entities.NodeTypeStart {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return nil, fmt.Errorf("start node not found")
	}

	// 2. 构建节点执行顺序（这里简化处理，实际需要根据边的关系来确定执行顺序）
	executionOrder := w.buildExecutionOrder()

	// 3. 按顺序执行节点
	for _, nodeID := range executionOrder {
		node := w.findNodeByID(nodeID)
		if node == nil {
			continue
		}

		// 执行节点（这里需要根据具体的节点类型来实现）
		nodeResult, err := w.executeNode(ctx, node, state)
		if err != nil {
			return nil, fmt.Errorf("failed to execute node %s: %w", node.Title, err)
		}

		// 将节点结果添加到状态中
		state.NodeResults = append(state.NodeResults, nodeResult)
	}

	return state, nil
}

// buildExecutionOrder 构建节点执行顺序
func (w *Workflow) buildExecutionOrder() []string {
	// 这里应该实现拓扑排序来确定节点的执行顺序
	// 简化处理，直接返回节点ID列表
	var order []string
	for _, node := range w.workflowConfig.Nodes {
		order = append(order, node.ID.String())
	}
	return order
}

// findNodeByID 根据ID查找节点
func (w *Workflow) findNodeByID(nodeID string) *entities.BaseNodeData {
	for _, node := range w.workflowConfig.Nodes {
		if node.ID.String() == nodeID {
			return node
		}
	}
	return nil
}

// executeNode 执行单个节点
func (w *Workflow) executeNode(ctx context.Context, node *entities.BaseNodeData, state *entities.WorkflowState) (*entities.NodeResult, error) {
	// 创建节点结果
	result := entities.NewNodeResult(node)

	// 获取当前状态的输入数据
	inputs, err := state.GetInputsAsMap()
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to get inputs: %v", err)
		return result, err
	}

	result.Inputs = inputs

	// 根据节点类型执行不同的逻辑
	switch node.NodeType {
	case entities.NodeTypeStart:
		// 开始节点：直接传递输入到输出
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded

		// 更新状态的输出
		if err := state.SetOutputsFromMap(result.Outputs); err != nil {
			result.Status = entities.NodeStatusFailed
			result.Error = fmt.Sprintf("failed to set outputs: %v", err)
			return result, err
		}

	case entities.NodeTypeEnd:
		// 结束节点：输出最终结果
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded

	default:
		// 其他节点类型的处理需要根据具体实现来完成
		result.Outputs = inputs
		result.Status = entities.NodeStatusSucceeded
	}

	return result, nil
}

// Stream 工作流流式输出每个节点对应的结果
func (w *Workflow) Stream(ctx context.Context, input string) (<-chan *entities.NodeResult, error) {
	// 解析输入参数
	var inputMap map[string]any
	if input != "" {
		if err := sonic.Unmarshal([]byte(input), &inputMap); err != nil {
			return nil, fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		inputMap = make(map[string]any)
	}

	// 创建工作流状态
	state := &entities.WorkflowState{}
	if err := state.SetInputsFromMap(inputMap); err != nil {
		return nil, fmt.Errorf("failed to set inputs: %w", err)
	}

	// 创建结果通道
	resultChan := make(chan *entities.NodeResult, len(w.workflowConfig.Nodes))

	// 启动goroutine执行工作流
	go func() {
		defer close(resultChan)

		executionOrder := w.buildExecutionOrder()
		for _, nodeID := range executionOrder {
			node := w.findNodeByID(nodeID)
			if node == nil {
				continue
			}

			nodeResult, err := w.executeNode(ctx, node, state)
			if err != nil {
				nodeResult.Status = entities.NodeStatusFailed
				nodeResult.Error = err.Error()
			}

			// 发送节点结果到通道
			select {
			case resultChan <- nodeResult:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}

func (w *Workflow) ValidateWorkflowConfig(values map[string]any) error {
	wc := w.workflowConfig
	// 1.获取工作流名字name，并校验是否符合规则
	if values["name"] == "" || !entities.WorkflowConfigNamePattern.MatchString(values["name"].(string)) {
		return fmt.Errorf("工作流名字仅支持字母、数字和下划线，且以字母/下划线为开头")
	}
	wc.Name = values["name"].(string)

	// 2.校验工作流的描述信息，该描述信息是传递给LLM使用，长度不能超过1024个字符
	if values["description"] == "" || len(values["description"].(string)) > entities.WorkflowConfigDescriptionMaxLength {
		return fmt.Errorf("工作流名字仅支持字母、数字和下划线，且以字母/下划线为开头")
	}
	wc.Description = values["description"].(string)

	// 3.校验nodes/edges数据类型和内容不能为空
	nodes := values["nodes"]
	edges := values["edges"]
	var nodesMap, edgesMap []map[string]any
	var ok bool
	if nodesMap, ok = nodes.([]map[string]any); !ok || len(nodesMap) <= 0 {
		return fmt.Errorf("工作流节点列表信息错误，请核实后重试")
	}
	if edgesMap, ok = edges.([]map[string]any); !ok || len(edgesMap) <= 0 {
		return fmt.Errorf("工作流边列表信息错误，请核实后重试")
	}

	// 4.循环遍历所有节点
	nodeDataDict := make(map[uuid.UUID]entities.NodeDataInterface)
	baseNodeDataDict := make(map[uuid.UUID]*entities.BaseNodeData)
	startNodes := 0
	endNodes := 0

	for _, nodeMap := range nodesMap {
		// 5.解析节点数据
		nodeData, err := parseNodeFromMapWithInterface(nodeMap)
		if err != nil {
			return fmt.Errorf("解析节点数据失败: %w", err)
		}

		baseData := nodeData.GetBaseNodeData()

		// 6.判断开始和结束节点是否唯一
		if baseData.NodeType == entities.NodeTypeStart {
			if startNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个开始节点")
			}
			startNodes++
		} else if baseData.NodeType == entities.NodeTypeEnd {
			if endNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个结束节点")
			}
			endNodes++
		}

		// 7.判断nodes节点数据id是否唯一
		if _, exists := nodeDataDict[baseData.ID]; exists {
			return fmt.Errorf("工作流节点id必须唯一，请核实后重试")
		}

		// 8.判断nodes节点数据title是否唯一
		for _, existingNode := range baseNodeDataDict {
			if strings.TrimSpace(existingNode.Title) == strings.TrimSpace(baseData.Title) {
				return fmt.Errorf("工作流节点title必须唯一，请核实后重试")
			}
		}

		// 9.将数据添加到nodeDataDict中
		nodeDataDict[baseData.ID] = nodeData
		baseNodeDataDict[baseData.ID] = baseData
	}

	// 将解析后的节点数据存储到工作流配置中
	wc.Nodes = make([]*entities.BaseNodeData, 0, len(baseNodeDataDict))
	for _, nodeData := range baseNodeDataDict {
		wc.Nodes = append(wc.Nodes, nodeData)
	}

	// 10.循环遍历edges数据
	edgeDataDict := make(map[uuid.UUID]*entities.BaseEdgeData)
	for _, edgeMap := range edgesMap {
		// 11.解析边数据
		edgeData, err := parseEdgeFromMap(edgeMap)
		if err != nil {
			return fmt.Errorf("解析边数据失败: %w", err)
		}

		// 12.校验边edges的id是否唯一
		if _, exists := edgeDataDict[edgeData.ID]; exists {
			return fmt.Errorf("工作流边数据id必须唯一，请核实后重试")
		}

		// 13.校验边中的source/target/source_type/target_type必须和nodes对得上
		sourceNode, sourceExists := nodeDataDict[edgeData.Source]
		targetNode, targetExists := nodeDataDict[edgeData.Target]

		if !sourceExists || edgeData.SourceType != sourceNode.GetBaseNodeData().NodeType ||
			!targetExists || edgeData.TargetType != targetNode.GetBaseNodeData().NodeType {
			return fmt.Errorf("工作流边起点/终点对应的节点不存在或类型错误，请核实后重试")
		}

		// 14.校验边Edges里的边必须唯一(source+target+source_handle_id必须唯一)
		for _, existingEdge := range edgeDataDict {
			if existingEdge.Source == edgeData.Source &&
				existingEdge.Target == edgeData.Target &&
				((existingEdge.SourceHandleID == nil && edgeData.SourceHandleID == nil) ||
					(existingEdge.SourceHandleID != nil && edgeData.SourceHandleID != nil &&
						*existingEdge.SourceHandleID == *edgeData.SourceHandleID)) {
				return fmt.Errorf("工作流边数据不能重复添加")
			}
		}

		// 15.基础数据校验通过，将数据添加到edgeDataDict中
		edgeDataDict[edgeData.ID] = edgeData
	}

	// 将解析后的边数据存储到工作流配置中
	wc.Edges = make([]*entities.BaseEdgeData, 0, len(edgeDataDict))
	for _, edgeData := range edgeDataDict {
		wc.Edges = append(wc.Edges, edgeData)
	}

	// 16.构建邻接表、逆邻接表、入度以及出度
	adjList := buildAdjList(wc.Edges)
	reverseAdjList := buildReverseAdjList(wc.Edges)
	inDegree, outDegree := buildDegrees(wc.Edges)

	// 17.从边的关系中校验是否有唯一的开始/结束节点(入度为0即为开始，出度为0即为结束)
	var startNodesList []*entities.BaseNodeData
	var endNodesList []*entities.BaseNodeData

	for _, node := range baseNodeDataDict {
		if inDegree[node.ID] == 0 {
			startNodesList = append(startNodesList, node)
		}
		if outDegree[node.ID] == 0 {
			endNodesList = append(endNodesList, node)
		}
	}

	if len(startNodesList) != 1 || len(endNodesList) != 1 ||
		startNodesList[0].NodeType != entities.NodeTypeStart || endNodesList[0].NodeType != entities.NodeTypeEnd {
		return fmt.Errorf("工作流中有且只有一个开始/结束节点作为图结构的起点和终点")
	}

	// 18.获取唯一的开始节点
	startNodeData := startNodesList[0]

	// 19.使用edges边信息校验图的连通性，确保没有孤立的节点
	if !isConnected(adjList, startNodeData.ID, len(baseNodeDataDict)) {
		return fmt.Errorf("工作流中存在不可到达节点，图不联通，请核实后重试")
	}

	// 20.校验edges中是否存在环路（即循环边结构）
	if isCycle(wc.Nodes, adjList, inDegree) {
		return fmt.Errorf("工作流中存在环路，请核实后重试")
	}

	// 21.校验nodes+edges中的数据引用是否正确，即inputs/outputs对应的数据
	if err := validateInputsRefDetailed(nodeDataDict, reverseAdjList); err != nil {
		return err
	}

	return nil
}

// buildAdjList 构建邻接表，邻接表的key为节点的id，值为该节点的所有直接子节点(后继节点)
func buildAdjList(edges []*entities.BaseEdgeData) map[uuid.UUID][]uuid.UUID {
	adjList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
	}
	return adjList
}

// buildReverseAdjList 构建逆邻接表，逆邻接表的key是每个节点的id，值为该节点的直接父节点
func buildReverseAdjList(edges []*entities.BaseEdgeData) map[uuid.UUID][]uuid.UUID {
	reverseAdjList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		reverseAdjList[edge.Target] = append(reverseAdjList[edge.Target], edge.Source)
	}
	return reverseAdjList
}

// buildDegrees 根据传递的边信息，计算每个节点的入度(inDegree)和出度(outDegree)
func buildDegrees(edges []*entities.BaseEdgeData) (map[uuid.UUID]int, map[uuid.UUID]int) {
	inDegree := make(map[uuid.UUID]int)
	outDegree := make(map[uuid.UUID]int)

	for _, edge := range edges {
		inDegree[edge.Target]++
		outDegree[edge.Source]++
	}

	return inDegree, outDegree
}

// isConnected 根据传递的邻接表+开始节点id，使用BFS广度优先搜索遍历，检查图是否流通
func isConnected(adjList map[uuid.UUID][]uuid.UUID, startNodeID uuid.UUID, totalNodes int) bool {
	visited := make(map[uuid.UUID]bool)
	queue := []uuid.UUID{startNodeID}
	visited[startNodeID] = true

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjList[nodeID] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// 计算已访问的节点数量是否和总节点数相等，如果不相等则表示存在孤立节点，图不连通
	return len(visited) == totalNodes
}

// isCycle 根据传递的节点列表、邻接表、入度数据，使用拓扑排序(Kahn算法)检测图中是否存在环
func isCycle(nodes []*entities.BaseNodeData, adjList map[uuid.UUID][]uuid.UUID, inDegree map[uuid.UUID]int) bool {
	// 复制入度映射，避免修改原始数据
	inDegreeCopy := make(map[uuid.UUID]int)
	for k, v := range inDegree {
		inDegreeCopy[k] = v
	}

	// 存储所有入度为0的节点id，即开始节点
	var zeroInDegreeNodes []uuid.UUID
	for _, node := range nodes {
		if inDegreeCopy[node.ID] == 0 {
			zeroInDegreeNodes = append(zeroInDegreeNodes, node.ID)
		}
	}

	visitedCount := 0

	// 循环遍历入度为0的节点信息
	for len(zeroInDegreeNodes) > 0 {
		// 从队列左侧取出一个入度为0的节点，并记录访问+1
		nodeID := zeroInDegreeNodes[0]
		zeroInDegreeNodes = zeroInDegreeNodes[1:]
		visitedCount++

		// 循环遍历取到的节点的所有子节点
		for _, neighbor := range adjList[nodeID] {
			// 将子节点的入度-1，并判断是否为0，如果是则添加到队列中
			inDegreeCopy[neighbor]--

			// Kahn算法的核心是，如果存在环，那么至少有一个非结束节点的入度大于等于2，并且该入度无法消减到0
			// 这就会导致该节点后续的所有子节点在该算法下都无法浏览，那么访问次数肯定小于总节点数
			if inDegreeCopy[neighbor] == 0 {
				zeroInDegreeNodes = append(zeroInDegreeNodes, neighbor)
			}
		}
	}

	// 判断访问次数和总节点数是否相等，如果不等/小于则说明存在环
	return visitedCount != len(nodes)
}

// getPredecessors 根据传递的逆邻接表+目标节点id，获取该节点的所有前置节点
func getPredecessors(reverseAdjList map[uuid.UUID][]uuid.UUID, targetNodeID uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	var predecessors []uuid.UUID

	var dfs func(uuid.UUID)
	dfs = func(nodeID uuid.UUID) {
		if !visited[nodeID] {
			visited[nodeID] = true
			if nodeID != targetNodeID {
				predecessors = append(predecessors, nodeID)
			}
			for _, neighbor := range reverseAdjList[nodeID] {
				dfs(neighbor)
			}
		}
	}

	dfs(targetNodeID)
	return predecessors
}

// validateInputsRef 校验输入数据引用是否正确，如果出错则直接抛出异常
func validateInputsRef(nodeDataDict map[uuid.UUID]*entities.BaseNodeData, reverseAdjList map[uuid.UUID][]uuid.UUID) error {
	// 由于Go版本中没有具体的节点输入输出变量信息，这里只做基础校验
	// 实际使用时需要根据具体的节点类型来实现详细的变量引用校验

	for _, nodeData := range nodeDataDict {
		// 提取该节点的所有前置节点
		predecessors := getPredecessors(reverseAdjList, nodeData.ID)

		// 如果节点数据类型不是START则校验输入数据引用（因为开始节点不需要校验）
		if nodeData.NodeType != entities.NodeTypeStart {
			// 这里需要根据具体的节点实现来校验变量引用
			// 由于当前Go版本中节点结构较简单，暂时跳过详细的变量引用校验
			_ = predecessors // 避免未使用变量警告
		}
	}

	return nil
}

// validateInputsRefDetailed 详细校验输入数据引用是否正确
func validateInputsRefDetailed(nodeDataDict map[uuid.UUID]entities.NodeDataInterface, reverseAdjList map[uuid.UUID][]uuid.UUID) error {
	// 循环遍历所有节点数据逐个处理
	for _, nodeData := range nodeDataDict {
		// 提取该节点的所有前置节点
		predecessors := getPredecessors(reverseAdjList, nodeData.GetBaseNodeData().ID)

		// 如果节点数据类型不是START则校验输入数据引用（因为开始节点不需要校验）
		if nodeData.GetBaseNodeData().NodeType != entities.NodeTypeStart {
			// 这里需要根据具体的节点实现来校验变量引用
			// 由于当前Go版本中节点结构较简单，我们可以添加基础的校验逻辑

			// 检查前置节点是否存在
			if len(predecessors) == 0 && nodeData.GetBaseNodeData().NodeType != entities.NodeTypeStart {
				// 除了开始节点，其他节点都应该有前置节点
				continue // 这里可以根据具体需求决定是否报错
			}

			// 这里可以添加更详细的变量引用校验逻辑
			// 例如检查节点的输入变量是否正确引用了前置节点的输出变量
		}
	}

	return nil
}

// getNodeDataClass 根据节点类型获取对应的节点数据类
func getNodeDataClass(nodeType entities.NodeType) (*entities.BaseNodeData, error) {
	switch nodeType {
	case entities.NodeTypeStart:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeStart}, nil
	case entities.NodeTypeEnd:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeEnd}, nil
	case entities.NodeTypeLLM:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeLLM}, nil
	case entities.NodeTypeTool:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeTool}, nil
	case entities.NodeTypeCode:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeCode}, nil
	case entities.NodeTypeDatasetRetrieval:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeDatasetRetrieval}, nil
	case entities.NodeTypeHTTPRequest:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeHTTPRequest}, nil
	case entities.NodeTypeTemplateTransform:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeTemplateTransform}, nil
	case entities.NodeTypeQuestionClassifier:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeQuestionClassifier}, nil
	case entities.NodeTypeIteration:
		return &entities.BaseNodeData{NodeType: entities.NodeTypeIteration}, nil
	default:
		return nil, fmt.Errorf("unsupported node type: %v", nodeType)
	}
}

// parseNodeFromMap 从map中解析节点数据
func parseNodeFromMap(nodeMap map[string]any) (*entities.BaseNodeData, error) {
	// 解析基础字段
	idStr, ok := nodeMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid node id format: %w", err)
	}

	title, ok := nodeMap["title"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node title")
	}

	nodeTypeStr, ok := nodeMap["node_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	nodeType := entities.NodeType(nodeTypeStr)

	// 创建基础节点数据
	nodeData := &entities.BaseNodeData{
		ID:       id,
		Title:    title,
		NodeType: nodeType,
	}

	return nodeData, nil
}

// parseEdgeFromMap 从map中解析边数据
func parseEdgeFromMap(edgeMap map[string]any) (*entities.BaseEdgeData, error) {
	// 解析基础字段
	idStr, ok := edgeMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge id format: %w", err)
	}

	sourceStr, ok := edgeMap["source"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge source")
	}

	source, err := uuid.Parse(sourceStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge source format: %w", err)
	}

	targetStr, ok := edgeMap["target"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge target")
	}

	target, err := uuid.Parse(targetStr)
	if err != nil {
		return nil, fmt.Errorf("invalid edge target format: %w", err)
	}

	sourceTypeStr, ok := edgeMap["source_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge source_type")
	}

	targetTypeStr, ok := edgeMap["target_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid edge target_type")
	}

	// 创建边数据
	edgeData := &entities.BaseEdgeData{
		ID:         id,
		Source:     source,
		Target:     target,
		SourceType: entities.NodeType(sourceTypeStr),
		TargetType: entities.NodeType(targetTypeStr),
	}

	// 解析可选字段
	if sourceHandleID, exists := edgeMap["source_handle_id"]; exists && sourceHandleID != nil {
		if handleIDStr, ok := sourceHandleID.(string); ok {
			edgeData.SourceHandleID = &handleIDStr
		}
	}

	return edgeData, nil
}

// parseNodeFromMapWithInterface 从map中解析节点数据，返回NodeDataInterface
func parseNodeFromMapWithInterface(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	// 解析节点类型
	nodeTypeStr, ok := nodeMap["node_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	nodeType := entities.NodeType(nodeTypeStr)

	// 根据节点类型创建对应的节点数据
	switch nodeType {
	case entities.NodeTypeStart:
		return parseStartNodeFromMap(nodeMap)
	case entities.NodeTypeEnd:
		return parseEndNodeFromMap(nodeMap)
	case entities.NodeTypeLLM:
		return parseLLMNodeFromMap(nodeMap)
	case entities.NodeTypeTool:
		return parseToolNodeFromMap(nodeMap)
	case entities.NodeTypeCode:
		return parseCodeNodeFromMap(nodeMap)
	case entities.NodeTypeDatasetRetrieval:
		return parseDatasetRetrievalNodeFromMap(nodeMap)
	case entities.NodeTypeHTTPRequest:
		return parseHTTPRequestNodeFromMap(nodeMap)
	case entities.NodeTypeTemplateTransform:
		return parseTemplateTransformNodeFromMap(nodeMap)
	case entities.NodeTypeQuestionClassifier:
		return parseQuestionClassifierNodeFromMap(nodeMap)
	case entities.NodeTypeIteration:
		return parseIterationNodeFromMap(nodeMap)
	default:
		return nil, fmt.Errorf("unsupported node type: %v", nodeType)
	}
}

// parseStartNodeFromMap 解析开始节点
func parseStartNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	startNode := &start.StartNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
	}

	// 解析inputs字段
	if inputsData, exists := nodeMap["inputs"]; exists {
		if inputsList, ok := inputsData.([]interface{}); ok {
			for _, inputData := range inputsList {
				if inputMap, ok := inputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(inputMap)
					if err != nil {
						return nil, fmt.Errorf("解析输入变量失败: %w", err)
					}
					startNode.Inputs = append(startNode.Inputs, variable)
				}
			}
		}
	}

	return startNode, nil
}

// parseEndNodeFromMap 解析结束节点
func parseEndNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	endNode := &end.EndNodeData{
		BaseNodeData: baseData,
		Outputs:      make([]*entities.VariableEntity, 0),
	}

	// 解析outputs字段
	if outputsData, exists := nodeMap["outputs"]; exists {
		if outputsList, ok := outputsData.([]interface{}); ok {
			for _, outputData := range outputsList {
				if outputMap, ok := outputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(outputMap)
					if err != nil {
						return nil, fmt.Errorf("解析输出变量失败: %w", err)
					}
					endNode.Outputs = append(endNode.Outputs, variable)
				}
			}
		}
	}

	return endNode, nil
}

// parseLLMNodeFromMap 解析LLM节点
func parseLLMNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	llmNode := &llm.LLMNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
		Model:        "gpt-3.5-turbo",
		MaxTokens:    1000,
		Temperature:  0.7,
	}

	// 解析inputs和outputs字段
	if inputsData, exists := nodeMap["inputs"]; exists {
		if inputsList, ok := inputsData.([]interface{}); ok {
			for _, inputData := range inputsList {
				if inputMap, ok := inputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(inputMap)
					if err != nil {
						return nil, fmt.Errorf("解析输入变量失败: %w", err)
					}
					llmNode.Inputs = append(llmNode.Inputs, variable)
				}
			}
		}
	}

	if outputsData, exists := nodeMap["outputs"]; exists {
		if outputsList, ok := outputsData.([]interface{}); ok {
			for _, outputData := range outputsList {
				if outputMap, ok := outputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(outputMap)
					if err != nil {
						return nil, fmt.Errorf("解析输出变量失败: %w", err)
					}
					llmNode.Outputs = append(llmNode.Outputs, variable)
				}
			}
		}
	}

	// 解析其他字段
	if model, exists := nodeMap["model"]; exists {
		if modelStr, ok := model.(string); ok {
			llmNode.Model = modelStr
		}
	}

	return llmNode, nil
}

// parseVariableFromMap 从map中解析变量
func parseVariableFromMap(varMap map[string]interface{}) (*entities.VariableEntity, error) {
	variable := entities.NewVariableEntity()

	if name, exists := varMap["name"]; exists {
		if nameStr, ok := name.(string); ok {
			variable.Name = nameStr
		}
	}

	if description, exists := varMap["description"]; exists {
		if descStr, ok := description.(string); ok {
			variable.Description = descStr
		}
	}

	if required, exists := varMap["required"]; exists {
		if reqBool, ok := required.(bool); ok {
			variable.Required = reqBool
		}
	}

	if varType, exists := varMap["type"]; exists {
		if typeStr, ok := varType.(string); ok {
			variable.Type = entities.VariableType(typeStr)
		}
	}

	if value, exists := varMap["value"]; exists {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if valueType, exists := valueMap["type"]; exists {
				if valueTypeStr, ok := valueType.(string); ok {
					variable.Value.Type = entities.VariableValueType(valueTypeStr)
				}
			}

			if content, exists := valueMap["content"]; exists {
				if variable.Value.Type == entities.VariableValueTypeRef {
					// 解析引用内容
					if contentMap, ok := content.(map[string]interface{}); ok {
						refContent := &entities.VariableContent{}

						if refNodeID, exists := contentMap["ref_node_id"]; exists {
							if refNodeIDStr, ok := refNodeID.(string); ok {
								if id, err := uuid.Parse(refNodeIDStr); err == nil {
									refContent.RefNodeID = &id
								}
							}
						}

						if refVarName, exists := contentMap["ref_var_name"]; exists {
							if refVarNameStr, ok := refVarName.(string); ok {
								refContent.RefVarName = refVarNameStr
							}
						}

						variable.Value.Content = refContent
					}
				} else {
					variable.Value.Content = content
				}
			}
		}
	}

	return variable, nil
}

// parseToolNodeFromMap 解析工具节点
func parseToolNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	toolNode := tool.NewToolNodeData()
	toolNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, toolNode); err != nil {
		return nil, err
	}

	return toolNode, nil
}

func parseCodeNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	codeNode := code.NewCodeNodeData()
	codeNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, codeNode); err != nil {
		return nil, err
	}

	return codeNode, nil
}

func parseDatasetRetrievalNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	datasetNode := dataset_retrieval.NewDatasetRetrievalNodeData()
	datasetNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, datasetNode); err != nil {
		return nil, err
	}

	return datasetNode, nil
}

func parseHTTPRequestNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	httpNode := http_request.NewHttpRequestNodeData()
	httpNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, httpNode); err != nil {
		return nil, err
	}

	return httpNode, nil
}

func parseTemplateTransformNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	templateNode := template_transform.NewTemplateTransformNodeData()
	templateNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, templateNode); err != nil {
		return nil, err
	}

	return templateNode, nil
}

func parseQuestionClassifierNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	qcNode := question_classifier.NewQuestionClassifierNodeData()
	qcNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, qcNode); err != nil {
		return nil, err
	}

	return qcNode, nil
}

func parseIterationNodeFromMap(nodeMap map[string]any) (entities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	iterNode := iteration.NewIterationNodeData()
	iterNode.BaseNodeData = baseData

	// 解析inputs和outputs字段
	if err := parseNodeInputsOutputs(nodeMap, iterNode); err != nil {
		return nil, err
	}

	return iterNode, nil
}

// parseNodeInputsOutputs 解析节点的inputs和outputs字段的通用方法
func parseNodeInputsOutputs(nodeMap map[string]any, nodeData entities.NodeDataInterface) error {
	// 解析inputs字段
	if inputsData, exists := nodeMap["inputs"]; exists {
		if inputsList, ok := inputsData.([]interface{}); ok {
			inputs := make([]*entities.VariableEntity, 0, len(inputsList))
			for _, inputData := range inputsList {
				if inputMap, ok := inputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(inputMap)
					if err != nil {
						return fmt.Errorf("解析输入变量失败: %w", err)
					}
					inputs = append(inputs, variable)
				}
			}
			setNodeInputs(nodeData, inputs)
		}
	}

	// 解析outputs字段
	if outputsData, exists := nodeMap["outputs"]; exists {
		if outputsList, ok := outputsData.([]interface{}); ok {
			outputs := make([]*entities.VariableEntity, 0, len(outputsList))
			for _, outputData := range outputsList {
				if outputMap, ok := outputData.(map[string]interface{}); ok {
					variable, err := parseVariableFromMap(outputMap)
					if err != nil {
						return fmt.Errorf("解析输出变量失败: %w", err)
					}
					outputs = append(outputs, variable)
				}
			}
			// 设置outputs
			setNodeOutputs(nodeData, outputs)
		}
	}

	return nil
}

// setNodeInputs 设置节点的inputs（使用类型断言）
func setNodeInputs(nodeData entities.NodeDataInterface, inputs []*entities.VariableEntity) {
	switch node := nodeData.(type) {
	case *tool.ToolNodeData:
		node.Inputs = inputs
	case *code.CodeNodeData:
		node.Inputs = inputs
	case *dataset_retrieval.DatasetRetrievalNodeData:
		node.Inputs = inputs
	case *http_request.HttpRequestNodeData:
		node.Inputs = inputs
	case *template_transform.TemplateTransformNodeData:
		node.Inputs = inputs
	case *question_classifier.QuestionClassifierNodeData:
		node.Inputs = inputs
	case *iteration.IterationNodeData:
		node.Inputs = inputs
	case *llm.LLMNodeData:
		node.Inputs = inputs
	}
}

// setNodeOutputs 设置节点的outputs（使用类型断言）
func setNodeOutputs(nodeData entities.NodeDataInterface, outputs []*entities.VariableEntity) {
	switch node := nodeData.(type) {
	case *tool.ToolNodeData:
		node.Outputs = outputs
	case *code.CodeNodeData:
		node.Outputs = outputs
	case *dataset_retrieval.DatasetRetrievalNodeData:
		node.Outputs = outputs
	case *http_request.HttpRequestNodeData:
		node.Outputs = outputs
	case *template_transform.TemplateTransformNodeData:
		node.Outputs = outputs
	case *question_classifier.QuestionClassifierNodeData:
		node.Outputs = outputs
	case *iteration.IterationNodeData:
		node.Outputs = outputs
	case *llm.LLMNodeData:
		node.Outputs = outputs
	case *end.EndNodeData:
		node.Outputs = outputs
	}
}
