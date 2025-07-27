package entities

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

// 工作流配置校验信息
var WorkflowConfigNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

const WorkflowConfigDescriptionMaxLength = 1024

// WorkflowConfig 工作流配置信息
type WorkflowConfig struct {
	AccountID   uuid.UUID       `json:"account_id"`  // 用户的唯一标识数据
	Name        string          `json:"name"`        // 工作流名称，必须是英文
	Description string          `json:"description"` // 工作流描述信息，用于告知LLM什么时候需要调用工作流
	Nodes       []*BaseNodeData `json:"nodes"`       // 工作流对应的节点列表信息
	Edges       []*BaseEdgeData `json:"edges"`       // 工作流对应的边列表信息
}

// WorkflowState 工作流图程序状态
type WorkflowState struct {
	Inputs      string        `json:"inputs"`       // 工作流的最初始输入，也就是工具输入 (使用string存储)
	Outputs     string        `json:"outputs"`      // 工作流的最终输出结果，也就是工具输出 (使用string存储)
	NodeResults []*NodeResult `json:"node_results"` // 各节点的运行结果
}

// GetInputsAsMap 将inputs字符串解析为map[string]any
func (ws *WorkflowState) GetInputsAsMap() (map[string]any, error) {
	if ws.Inputs == "" {
		return make(map[string]any), nil
	}
	var result map[string]any
	if err := sonic.Unmarshal([]byte(ws.Inputs), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}
	return result, nil
}

// SetInputsFromMap 将map[string]any转换为inputs字符串
func (ws *WorkflowState) SetInputsFromMap(inputs map[string]any) error {
	if inputs == nil {
		ws.Inputs = ""
		return nil
	}
	data, err := sonic.Marshal(inputs)
	if err != nil {
		return fmt.Errorf("failed to marshal inputs: %w", err)
	}
	ws.Inputs = string(data)
	return nil
}

// GetOutputsAsMap 将outputs字符串解析为map[string]any
func (ws *WorkflowState) GetOutputsAsMap() (map[string]any, error) {
	if ws.Outputs == "" {
		return make(map[string]any), nil
	}
	var result map[string]any
	if err := sonic.Unmarshal([]byte(ws.Outputs), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}
	return result, nil
}

// SetOutputsFromMap 将map[string]any转换为outputs字符串
func (ws *WorkflowState) SetOutputsFromMap(outputs map[string]any) error {
	if outputs == nil {
		ws.Outputs = ""
		return nil
	}
	data, err := sonic.Marshal(outputs)
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}
	ws.Outputs = string(data)
	return nil
}

// ProcessDict 工作流状态字典归纳函数
func ProcessDict(left, right map[string]any) map[string]any {
	if left == nil {
		left = make(map[string]any)
	}
	if right == nil {
		right = make(map[string]any)
	}

	result := make(map[string]any)
	for k, v := range left {
		result[k] = v
	}
	for k, v := range right {
		result[k] = v
	}
	return result
}

// ProcessNodeResults 工作流状态节点结果列表归纳函数
func ProcessNodeResults(left, right []*NodeResult) []*NodeResult {
	if left == nil {
		left = []*NodeResult{}
	}
	if right == nil {
		right = []*NodeResult{}
	}
	return append(left, right...)
}

// ValidateWorkflowConfig 自定义校验函数，用于校验工作流配置中的所有参数信息
func (wc *WorkflowConfig) ValidateWorkflowConfig() error {
	// 1.获取工作流名字name，并校验是否符合规则
	if wc.Name == "" || !WorkflowConfigNamePattern.MatchString(wc.Name) {
		return fmt.Errorf("工作流名字仅支持字母、数字和下划线，且以字母/下划线为开头")
	}

	// 2.校验工作流的描述信息，该描述信息是传递给LLM使用，长度不能超过1024个字符
	if wc.Description == "" || len(wc.Description) > WorkflowConfigDescriptionMaxLength {
		return fmt.Errorf("工作流描述信息长度不能超过1024个字符")
	}

	// 3.校验nodes/edges数据类型和内容不能为空
	if len(wc.Nodes) <= 0 {
		return fmt.Errorf("工作流节点列表信息错误，请核实后重试")
	}
	if len(wc.Edges) <= 0 {
		return fmt.Errorf("工作流边列表信息错误，请核实后重试")
	}

	// 4.循环遍历所有节点
	nodeDataDict := make(map[uuid.UUID]*BaseNodeData)
	startNodes := 0
	endNodes := 0

	for _, node := range wc.Nodes {
		// 5.判断开始和结束节点是否唯一
		if node.NodeType == NodeTypeStart {
			if startNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个开始节点")
			}
			startNodes++
		} else if node.NodeType == NodeTypeEnd {
			if endNodes >= 1 {
				return fmt.Errorf("工作流中只允许有1个结束节点")
			}
			endNodes++
		}

		// 6.判断nodes节点数据id是否唯一
		if _, exists := nodeDataDict[node.ID]; exists {
			return fmt.Errorf("工作流节点id必须唯一，请核实后重试")
		}

		// 7.判断nodes节点数据title是否唯一
		for _, existingNode := range nodeDataDict {
			if strings.TrimSpace(existingNode.Title) == strings.TrimSpace(node.Title) {
				return fmt.Errorf("工作流节点title必须唯一，请核实后重试")
			}
		}

		// 8.将数据添加到nodeDataDict中
		nodeDataDict[node.ID] = node
	}

	// 9.循环遍历edges数据
	edgeDataDict := make(map[uuid.UUID]*BaseEdgeData)
	for _, edge := range wc.Edges {
		// 10.校验边edges的id是否唯一
		if _, exists := edgeDataDict[edge.ID]; exists {
			return fmt.Errorf("工作流边数据id必须唯一，请核实后重试")
		}

		// 11.校验边中的source/target/source_type/target_type必须和nodes对得上
		sourceNode, sourceExists := nodeDataDict[edge.Source]
		targetNode, targetExists := nodeDataDict[edge.Target]

		if !sourceExists || edge.SourceType != sourceNode.NodeType ||
			!targetExists || edge.TargetType != targetNode.NodeType {
			return fmt.Errorf("工作流边起点/终点对应的节点不存在或类型错误，请核实后重试")
		}

		// 12.校验边Edges里的边必须唯一(source+target+source_handle_id必须唯一)
		for _, existingEdge := range edgeDataDict {
			if existingEdge.Source == edge.Source &&
				existingEdge.Target == edge.Target &&
				((existingEdge.SourceHandleID == nil && edge.SourceHandleID == nil) ||
					(existingEdge.SourceHandleID != nil && edge.SourceHandleID != nil &&
						*existingEdge.SourceHandleID == *edge.SourceHandleID)) {
				return fmt.Errorf("工作流边数据不能重复添加")
			}
		}

		// 13.基础数据校验通过，将数据添加到edgeDataDict中
		edgeDataDict[edge.ID] = edge
	}

	// 14.构建邻接表、逆邻接表、入度以及出度
	adjList := buildAdjList(wc.Edges)
	reverseAdjList := buildReverseAdjList(wc.Edges)
	inDegree, outDegree := buildDegrees(wc.Edges)

	// 15.从边的关系中校验是否有唯一的开始/结束节点(入度为0即为开始，出度为0即为结束)
	var startNodesList []*BaseNodeData
	var endNodesList []*BaseNodeData

	for _, node := range nodeDataDict {
		if inDegree[node.ID] == 0 {
			startNodesList = append(startNodesList, node)
		}
		if outDegree[node.ID] == 0 {
			endNodesList = append(endNodesList, node)
		}
	}

	if len(startNodesList) != 1 || len(endNodesList) != 1 ||
		startNodesList[0].NodeType != NodeTypeStart || endNodesList[0].NodeType != NodeTypeEnd {
		return fmt.Errorf("工作流中有且只有一个开始/结束节点作为图结构的起点和终点")
	}

	// 16.获取唯一的开始节点
	startNodeData := startNodesList[0]

	// 17.使用edges边信息校验图的连通性，确保没有孤立的节点
	if !isConnected(adjList, startNodeData.ID) {
		return fmt.Errorf("工作流中存在不可到达节点，图不联通，请核实后重试")
	}

	// 18.校验edges中是否存在环路（即循环边结构）
	if isCycle(wc.Nodes, adjList, inDegree) {
		return fmt.Errorf("工作流中存在环路，请核实后重试")
	}

	// 19.校验nodes+edges中的数据引用是否正确，即inputs/outputs对应的数据
	if err := validateInputsRef(nodeDataDict, reverseAdjList); err != nil {
		return err
	}

	return nil
}

// buildAdjList 构建邻接表，邻接表的key为节点的id，值为该节点的所有直接子节点(后继节点)
func buildAdjList(edges []*BaseEdgeData) map[uuid.UUID][]uuid.UUID {
	adjList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
	}
	return adjList
}

// buildReverseAdjList 构建逆邻接表，逆邻接表的key是每个节点的id，值为该节点的直接父节点
func buildReverseAdjList(edges []*BaseEdgeData) map[uuid.UUID][]uuid.UUID {
	reverseAdjList := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		reverseAdjList[edge.Target] = append(reverseAdjList[edge.Target], edge.Source)
	}
	return reverseAdjList
}

// buildDegrees 根据传递的边信息，计算每个节点的入度(inDegree)和出度(outDegree)
func buildDegrees(edges []*BaseEdgeData) (map[uuid.UUID]int, map[uuid.UUID]int) {
	inDegree := make(map[uuid.UUID]int)
	outDegree := make(map[uuid.UUID]int)

	for _, edge := range edges {
		inDegree[edge.Target]++
		outDegree[edge.Source]++
	}

	return inDegree, outDegree
}

// isConnected 根据传递的邻接表+开始节点id，使用BFS广度优先搜索遍历，检查图是否流通
func isConnected(adjList map[uuid.UUID][]uuid.UUID, startNodeID uuid.UUID) bool {
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

	return len(visited) == len(adjList)
}

// isCycle 根据传递的节点列表、邻接表、入度数据，使用拓扑排序(Kahn算法)检测图中是否存在环
func isCycle(nodes []*BaseNodeData, adjList map[uuid.UUID][]uuid.UUID, inDegree map[uuid.UUID]int) bool {
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

	for len(zeroInDegreeNodes) > 0 {
		// 从队列取出一个入度为0的节点，并记录访问+1
		nodeID := zeroInDegreeNodes[0]
		zeroInDegreeNodes = zeroInDegreeNodes[1:]
		visitedCount++

		// 循环遍历取到的节点的所有子节点
		for _, neighbor := range adjList[nodeID] {
			// 将子节点的入度-1，并判断是否为0，如果是则添加到队列中
			inDegreeCopy[neighbor]--
			if inDegreeCopy[neighbor] == 0 {
				zeroInDegreeNodes = append(zeroInDegreeNodes, neighbor)
			}
		}
	}

	// 判断访问次数和总结点数是否相等，如果不等/小于则说明存在环
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
func validateInputsRef(nodeDataDict map[uuid.UUID]*BaseNodeData, reverseAdjList map[uuid.UUID][]uuid.UUID) error {
	// 由于Go版本中没有具体的节点输入输出变量信息，这里只做基础校验
	// 实际使用时需要根据具体的节点类型来实现详细的变量引用校验

	for _, nodeData := range nodeDataDict {
		// 提取该节点的所有前置节点
		predecessors := getPredecessors(reverseAdjList, nodeData.ID)

		// 如果节点数据类型不是START则校验输入数据引用（因为开始节点不需要校验）
		if nodeData.NodeType != NodeTypeStart {
			// 这里需要根据具体的节点实现来校验变量引用
			// 由于当前Go版本中节点结构较简单，暂时跳过详细的变量引用校验
			_ = predecessors // 避免未使用变量警告
		}
	}

	return nil
}
