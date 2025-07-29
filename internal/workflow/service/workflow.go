package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	corewf "github.com/crazyfrankie/voidx/internal/core/workflow"
	wfentities "github.com/crazyfrankie/voidx/internal/core/workflow/entities"
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
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/workflow/repository"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type WorkflowService struct {
	repo            *repository.WorkflowRepo
	builtinProvider *builtin.BuiltinProviderManager
}

func NewWorkflowService(repo *repository.WorkflowRepo, builtinProvider *builtin.BuiltinProviderManager) *WorkflowService {
	return &WorkflowService{
		repo:            repo,
		builtinProvider: builtinProvider,
	}
}

// CreateWorkflow 根据传递的请求信息创建工作流
func (s *WorkflowService) CreateWorkflow(ctx context.Context, userID uuid.UUID, createReq req.CreateWorkflowReq) (*entity.Workflow, error) {
	// 1. 检查工具调用名称是否重复
	existing, err := s.repo.GetWorkflowByToolCallName(ctx, userID, strings.TrimSpace(createReq.ToolCallName))
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errno.ErrValidate.AppendBizMessage(fmt.Sprintf("在当前账号下已创建[%s]工作流，不支持重名", createReq.ToolCallName))
	}

	// 2. 创建工作流
	workflow := &entity.Workflow{
		AccountID:     userID,
		Name:          createReq.Name,
		ToolCallName:  strings.TrimSpace(createReq.ToolCallName),
		Icon:          createReq.Icon,
		Description:   createReq.Description,
		IsDebugPassed: false,
		Status:        consts.WorkflowStatusDraft,
	}

	err = s.repo.CreateWorkflow(ctx, workflow)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

// GetWorkflow 根据传递的工作流id，获取指定的工作流基础信息
func (s *WorkflowService) GetWorkflow(ctx context.Context, workflowID, userID uuid.UUID) (*resp.GetWorkflowResp, error) {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	// 计算节点数量
	draftGraph := workflow.DraftGraph
	nodeCount := 0
	if nodes, ok := draftGraph["nodes"].([]any); ok {
		nodeCount = len(nodes)
	}

	publishedAt := int64(0)
	if workflow.PublishedAt != 0 {
		publishedAt = workflow.PublishedAt
	}

	return &resp.GetWorkflowResp{
		ID:            workflow.ID,
		Name:          workflow.Name,
		ToolCallName:  workflow.ToolCallName,
		Icon:          workflow.Icon,
		Description:   workflow.Description,
		Status:        string(workflow.Status),
		IsDebugPassed: workflow.IsDebugPassed,
		NodeCount:     nodeCount,
		PublishedAt:   publishedAt,
		Utime:         workflow.Utime,
		Ctime:         workflow.Ctime,
	}, nil
}

// DeleteWorkflow 根据传递的工作流id+账号信息，删除指定的工作流
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, workflowID, userID uuid.UUID) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	return s.repo.DeleteWorkflow(ctx, workflowID)
}

// UpdateWorkflow 根据传递的工作流id+请求更新工作流基础信息
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, workflowID, userID uuid.UUID, updateReq req.UpdateWorkflowReq) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	// 检查工具调用名称是否重复
	toolCallName := strings.TrimSpace(updateReq.ToolCallName)
	existing, err := s.repo.GetWorkflowByToolCallName(ctx, userID, toolCallName)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != workflowID {
		return errno.ErrValidate.AppendBizMessage(fmt.Sprintf("在当前账号下已创建[%s]工作流，不支持重名", toolCallName))
	}

	updates := map[string]any{
		"name":           updateReq.Name,
		"tool_call_name": toolCallName,
		"icon":           updateReq.Icon,
		"description":    updateReq.Description,
	}

	return s.repo.UpdateWorkflow(ctx, workflowID, updates)
}

// GetWorkflowsWithPage 根据传递的信息获取工作流分页列表数据
func (s *WorkflowService) GetWorkflowsWithPage(ctx context.Context, userID uuid.UUID, pageReq req.GetWorkflowsWithPageReq) ([]resp.GetWorkflowsWithPageResp, resp.Paginator, error) {
	workflows, total, err := s.repo.GetWorkflowsByAccountID(ctx, userID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式
	workflowResps := make([]resp.GetWorkflowsWithPageResp, len(workflows))
	for i, workflow := range workflows {
		// 计算节点数量
		graph := workflow.Graph
		nodeCount := 0
		if nodes, ok := graph["nodes"].([]any); ok {
			nodeCount = len(nodes)
		}

		publishedAt := int64(0)
		if workflow.PublishedAt != 0 {
			publishedAt = workflow.PublishedAt
		}

		workflowResps[i] = resp.GetWorkflowsWithPageResp{
			ID:            workflow.ID,
			Name:          workflow.Name,
			ToolCallName:  workflow.ToolCallName,
			Icon:          workflow.Icon,
			Description:   workflow.Description,
			Status:        string(workflow.Status),
			IsDebugPassed: workflow.IsDebugPassed,
			NodeCount:     nodeCount,
			PublishedAt:   publishedAt,
			Utime:         workflow.Utime,
			Ctime:         workflow.Ctime,
		}
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return workflowResps, paginator, nil
}

// UpdateDraftGraph 根据传递的工作流id+草稿图配置+账号更新工作流的草稿图
func (s *WorkflowService) UpdateDraftGraph(ctx context.Context, workflowID, userID uuid.UUID, draftGraph map[string]any) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	validateDraftGraph, err := s.validateGraph(ctx, workflowID, draftGraph, userID)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"draft_graph":     validateDraftGraph,
		"is_debug_passed": false,
	}

	return s.repo.UpdateWorkflow(ctx, workflowID, updates)
}

// GetDraftGraph 根据传递的工作流id+账号信息，获取指定工作流的草稿配置信息
func (s *WorkflowService) GetDraftGraph(ctx context.Context, workflowID, userID uuid.UUID) (map[string]any, error) {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	validateDraftGraph, err := s.validateGraph(ctx, workflowID, workflow.DraftGraph, userID)
	if err != nil {
		return nil, err
	}

	for _, node := range validateDraftGraph["nodes"].([]*wfentities.BaseNodeData) {
		if node.NodeType == wfentities.NodeTypeTool {

		}
	}

	return validateDraftGraph, nil
}

// DebugWorkflow 调试指定的工作流API接口，该接口为流式事件输出
func (s *WorkflowService) DebugWorkflow(ctx context.Context, workflowID, userID uuid.UUID, inputs map[string]any) (<-chan resp.WorkflowDebugEvent, error) {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	// 创建事件通道
	eventChan := make(chan resp.WorkflowDebugEvent, 100)

	// 启动异步处理
	go s.processWorkflowDebug(ctx, workflow, inputs, eventChan)

	return eventChan, nil
}

// PublishWorkflow 根据传递的工作流id，发布指定的工作流
func (s *WorkflowService) PublishWorkflow(ctx context.Context, workflowID, userID uuid.UUID) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	if !workflow.IsDebugPassed {
		return errno.ErrValidate.AppendBizMessage("该工作流未调试通过，请调试通过后发布")
	}

	if _, err := corewf.NewWorkflow(map[string]any{
		"account_id":  userID,
		"name":        workflow.Name,
		"description": workflow.Description,
		"nodes":       workflow.DraftGraph["nodes"],
		"edges":       workflow.DraftGraph["edges"],
	}); err != nil {
		workflow.IsDebugPassed = false
		if err := s.repo.SaveWorkflow(ctx, workflow); err != nil {
			return errno.ErrInternalServer.AppendBizMessage(fmt.Sprintf("保存工作流状态失败: %v", err))
		}
	}

	now := time.Now().UnixMilli()
	updates := map[string]any{
		"graph":           workflow.DraftGraph,
		"status":          consts.WorkflowStatusPublished,
		"is_debug_passed": false,
		"published_at":    now,
	}

	return s.repo.UpdateWorkflow(ctx, workflowID, updates)
}

// CancelPublishWorkflow 取消发布指定的工作流
func (s *WorkflowService) CancelPublishWorkflow(ctx context.Context, workflowID, userID uuid.UUID) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("该工作流不存在，请核实后重试")
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("当前账号无权限访问该应用，请核实后尝试")
	}

	if workflow.Status != consts.WorkflowStatusPublished {
		return errno.ErrValidate.AppendBizMessage("该工作流未发布无法取消发布")
	}

	emptyGraphJSON, _ := sonic.Marshal(map[string]any{})
	updates := map[string]any{
		"graph":           emptyGraphJSON,
		"status":          consts.WorkflowStatusDraft,
		"is_debug_passed": false,
	}

	return s.repo.UpdateWorkflow(ctx, workflowID, updates)
}

// processWorkflowDebug 处理工作流调试
func (s *WorkflowService) processWorkflowDebug(ctx context.Context, workflow *entity.Workflow,
	inputs map[string]any, eventChan chan<- resp.WorkflowDebugEvent) {
	defer close(eventChan)

	// 创建工作流运行结果记录
	workflowResult := &entity.WorkflowResult{
		AccountID:  workflow.AccountID,
		WorkflowID: workflow.ID,
		Graph:      workflow.DraftGraph,
		Status:     consts.WorkflowResultStatusRunning,
	}

	s.repo.CreateWorkflowResult(ctx, workflowResult)

	startTime := time.Now()

	// 模拟工作流执行过程
	events := []resp.WorkflowDebugEvent{
		{
			ID:          uuid.New().String(),
			NodeID:      "start_node",
			NodeType:    "start",
			Title:       "开始节点",
			Status:      "running",
			Inputs:      inputs,
			ElapsedTime: 0.1,
		},
		{
			ID:          uuid.New().String(),
			NodeID:      "start_node",
			NodeType:    "start",
			Title:       "开始节点",
			Status:      "completed",
			Inputs:      inputs,
			Outputs:     map[string]any{"result": "started"},
			ElapsedTime: 0.2,
		},
	}

	// 发送事件
	for _, event := range events {
		select {
		case eventChan <- event:
			time.Sleep(100 * time.Millisecond) // 模拟处理时间
		case <-ctx.Done():
			return
		}
	}

	// 更新工作流运行结果
	latency := time.Since(startTime).Seconds()
	stateJSON, _ := sonic.Marshal(events)
	updates := map[string]any{
		"status":  consts.WorkflowResultStatusSucceeded,
		"state":   stateJSON,
		"latency": latency,
	}
	s.repo.UpdateWorkflowResult(ctx, workflowResult.ID, updates)

	// 更新工作流调试状态
	s.repo.UpdateWorkflow(ctx, workflow.ID, map[string]any{
		"is_debug_passed": true,
	})
}

func (s *WorkflowService) validateGraph(ctx context.Context, workflowID uuid.UUID, graph map[string]any, accountID uuid.UUID) (map[string]any, error) {
	// 1. 提取 nodes 和 edges 数据
	nodes, _ := graph["nodes"].([]any)
	edges, _ := graph["edges"].([]any)

	// 2. 构建节点类型与节点数据类映射
	nodeDataClasses := map[string]any{
		"START":               &start.StartNodeData{},
		"END":                 &end.EndNodeData{},
		"LLM":                 &llm.LLMNodeData{},
		"TEMPLATE_TRANSFORM":  &template_transform.TemplateTransformNodeData{},
		"DATASET_RETRIEVAL":   &dataset_retrieval.DatasetRetrievalNodeData{},
		"CODE":                &code.CodeNodeData{},
		"TOOL":                &tool.ToolNodeData{},
		"HTTP_REQUEST":        &http_request.HttpRequestNodeData{},
		"QUESTION_CLASSIFIER": &question_classifier.QuestionClassifierNodeData{},
		"ITERATION":           &iteration.IterationNodeData{},
	}

	// 3. 循环校验 nodes 中各个节点对应的数据
	nodeDataDict := make(map[uuid.UUID]wfentities.NodeDataInterface)
	startNodes := 0
	endNodes := 0

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("工作流节点数据类型出错，请核实后重试")
		}

		// 5. 提取节点的 node_type 类型，并判断类型是否正确
		nodeType, ok := nodeMap["node_type"].(string)
		if !ok {
			return nil, fmt.Errorf("工作流节点类型出错，请核实后重试")
		}

		nodeDataCls, exists := nodeDataClasses[nodeType]
		if !exists {
			return nil, fmt.Errorf("工作流节点类型出错，请核实后重试")
		}

		// 6. 实例化节点数据类型
		nodeData, err := s.createNodeData(nodeDataCls, nodeMap)
		if err != nil {
			return nil, err
		}

		// 7. 判断节点 id 是否唯一
		if _, exists := nodeDataDict[nodeData.GetBaseNodeData().ID]; exists {
			return nil, fmt.Errorf("工作流节点id必须唯一，请核实后重试")
		}

		// 8. 判断节点 title 是否唯一
		for _, existingNode := range nodeDataDict {
			node, ok := existingNode.(wfentities.NodeDataInterface)
			if !ok {
				return nil, errno.ErrValidate
			}
			if node.GetBaseNodeData().Title == nodeData.GetBaseNodeData().Title {
				return nil, fmt.Errorf("工作流节点title必须唯一，请核实后重试")
			}
		}

		// 9. 对特殊节点进行判断
		switch nodeType {
		case "START":
			if startNodes >= 1 {
				return nil, fmt.Errorf("工作流中只允许有1个开始节点")
			}
			startNodes++
		case "END":
			if endNodes >= 1 {
				return nil, fmt.Errorf("工作流中只允许有1个结束节点")
			}
			endNodes++
		case "DATASET_RETRIEVAL":
			retrievalNode, ok := nodeData.(*dataset_retrieval.DatasetRetrievalNodeData)
			if !ok {
				return nil, errno.ErrValidate
			}
			// 10. 剔除关联知识库列表中不属于当前账户的数据
			datasets, err := s.repo.GetDatasets(ctx, accountID, retrievalNode.DatasetIDs[:5])
			if err != nil {
				return nil, err
			}
			var validDatasetIDs []uuid.UUID
			for _, dataset := range datasets {
				validDatasetIDs = append(validDatasetIDs, dataset.ID)
			}
			retrievalNode.DatasetIDs = validDatasetIDs
		case "ITERATION":
			iterationNode, ok := nodeData.(*iteration.IterationNodeData)
			if !ok {
				return nil, errno.ErrValidate
			}
			// 11. 校验迭代节点
			workflows, err := s.repo.GetWorkflows(ctx, iterationNode.WorkflowIDs[:1], accountID)
			if err != nil {
				return nil, err
			}
			// 剔除当前工作流
			filteredIDs := make([]uuid.UUID, 0)
			for _, w := range workflows {
				if w.ID != workflowID {
					filteredIDs = append(filteredIDs, w.ID)
				}
			}
			iterationNode.WorkflowIDs = filteredIDs
		}

		nodeDataDict[nodeData.GetBaseNodeData().ID] = nodeData
	}

	// 14. 循环校验 edges 中各个节点对应的数据
	edgeDataDict := make(map[uuid.UUID]*wfentities.BaseEdgeData)
	for _, edge := range edges {
		edgeMap, ok := edge.(map[string]any)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage("工作流边数据类型出错，请核实后重试")
		}
		edgeData, err := s.createEdgeData(edgeMap)
		if err != nil {
			continue
		}
		// 16. 校验边 edges 的 id 是否唯一
		if _, exists := edgeDataDict[edgeData.ID]; exists {
			continue
		}

		// 17. 校验边中的 source/target/source_type/target_type 必须和 nodes 对得上
		sourceNode, sourceExists := nodeDataDict[edgeData.Source]
		targetNode, targetExists := nodeDataDict[edgeData.Target]
		if !sourceExists || !targetExists ||
			edgeData.SourceType != sourceNode.GetBaseNodeData().NodeType ||
			edgeData.TargetType != targetNode.GetBaseNodeData().NodeType {
			continue
		}

		// 18. 校验边 Edges 里的边必须唯一
		for _, existingEdge := range edgeDataDict {
			if existingEdge.Source == edgeData.Source &&
				existingEdge.Target == edgeData.Target &&
				existingEdge.SourceHandleID == edgeData.SourceHandleID {
				continue
			}
		}

		edgeDataDict[edgeData.ID] = edgeData
	}

	// 转换结果
	resNodes, err := s.convertNodesToMap(nodeDataDict)
	if err != nil {
		return nil, err
	}
	resEdges, err := s.convertEdgesToMap(edgeDataDict)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"nodes": resNodes,
		"edges": resEdges,
	}

	return result, nil
}

// 辅助方法
func (s *WorkflowService) createNodeData(nodeType any, data map[string]any) (wfentities.NodeDataInterface, error) {
	switch nodeType {
	case wfentities.NodeTypeStart:
		return parseStartNodeFromMap(data)
	case wfentities.NodeTypeEnd:
		return parseEndNodeFromMap(data)
	case wfentities.NodeTypeLLM:
		return parseLLMNodeFromMap(data)
	case wfentities.NodeTypeTool:
		return parseToolNodeFromMap(data)
	case wfentities.NodeTypeCode:
		return parseCodeNodeFromMap(data)
	case wfentities.NodeTypeDatasetRetrieval:
		return parseDatasetRetrievalNodeFromMap(data)
	case wfentities.NodeTypeHTTPRequest:
		return parseHTTPRequestNodeFromMap(data)
	case wfentities.NodeTypeTemplateTransform:
		return parseTemplateTransformNodeFromMap(data)
	case wfentities.NodeTypeQuestionClassifier:
		return parseQuestionClassifierNodeFromMap(data)
	case wfentities.NodeTypeIteration:
		return parseIterationNodeFromMap(data)
	default:
		return nil, fmt.Errorf("unsupported node type: %v", nodeType)
	}
}

// parseStartNodeFromMap 解析开始节点
func parseStartNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	startNode := &start.StartNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*wfentities.VariableEntity, 0),
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
func parseEndNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	endNode := &end.EndNodeData{
		BaseNodeData: baseData,
		Outputs:      make([]*wfentities.VariableEntity, 0),
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
func parseLLMNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
	baseData, err := parseNodeFromMap(nodeMap)
	if err != nil {
		return nil, err
	}

	llmNode := &llm.LLMNodeData{
		BaseNodeData: baseData,
		Inputs:       make([]*wfentities.VariableEntity, 0),
		Outputs:      make([]*wfentities.VariableEntity, 0),
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
func parseVariableFromMap(varMap map[string]interface{}) (*wfentities.VariableEntity, error) {
	variable := wfentities.NewVariableEntity()

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
			variable.Type = wfentities.VariableType(typeStr)
		}
	}

	if value, exists := varMap["value"]; exists {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if valueType, exists := valueMap["type"]; exists {
				if valueTypeStr, ok := valueType.(string); ok {
					variable.Value.Type = wfentities.VariableValueType(valueTypeStr)
				}
			}

			if content, exists := valueMap["content"]; exists {
				if variable.Value.Type == wfentities.VariableValueTypeRef {
					// 解析引用内容
					if contentMap, ok := content.(map[string]interface{}); ok {
						refContent := &wfentities.VariableContent{}

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
func parseToolNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseCodeNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseDatasetRetrievalNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseHTTPRequestNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseTemplateTransformNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseQuestionClassifierNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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

func parseIterationNodeFromMap(nodeMap map[string]any) (wfentities.NodeDataInterface, error) {
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
func parseNodeInputsOutputs(nodeMap map[string]any, nodeData wfentities.NodeDataInterface) error {
	// 解析inputs字段
	if inputsData, exists := nodeMap["inputs"]; exists {
		if inputsList, ok := inputsData.([]interface{}); ok {
			inputs := make([]*wfentities.VariableEntity, 0, len(inputsList))
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
			outputs := make([]*wfentities.VariableEntity, 0, len(outputsList))
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
func setNodeInputs(nodeData wfentities.NodeDataInterface, inputs []*wfentities.VariableEntity) {
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
func setNodeOutputs(nodeData wfentities.NodeDataInterface, outputs []*wfentities.VariableEntity) {
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

// parseNodeFromMap 从map中解析节点数据
func parseNodeFromMap(nodeMap map[string]any) (*wfentities.BaseNodeData, error) {
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

	nodeType := wfentities.NodeType(nodeTypeStr)

	// 创建基础节点数据
	nodeData := &wfentities.BaseNodeData{
		ID:       id,
		Title:    title,
		NodeType: nodeType,
	}

	return nodeData, nil
}

func (s *WorkflowService) createEdgeData(data map[string]any) (*wfentities.BaseEdgeData, error) {
	edge := wfentities.NewBaseEdgeData()
	target, ok := data["target"].(uuid.UUID)
	if !ok {
		return nil, errno.ErrValidate.AppendBizMessage("工作流边数据类型出错，请核实后重试")
	}
	edge.Target = target
	nodeType, ok := data["node_type"].(wfentities.NodeType)
	if !ok {
		return nil, errno.ErrValidate.AppendBizMessage("工作流边数据类型出错，请核实后重试")
	}
	edge.TargetType = nodeType
	source, ok := data["source"].(uuid.UUID)
	if !ok {
		return nil, errno.ErrValidate.AppendBizMessage("工作流边数据类型出错，请核实后重试")
	}
	edge.Source = source
	sourceType, ok := data["source_type"].(wfentities.NodeType)
	if !ok {
		return nil, errno.ErrValidate.AppendBizMessage("工作流边数据类型出错，请核实后重试")
	}
	edge.SourceType = sourceType

	return edge, nil
}

func (s *WorkflowService) convertNodesToMap(nodes map[uuid.UUID]wfentities.NodeDataInterface) ([]map[string]any, error) {
	res := make([]map[string]any, 0, len(nodes))
	for id, node := range nodes {
		nodeMap, err := s.convertModelToMap(node)
		if err != nil {
			return nil, err
		}
		nodeMap["id"] = id.String()
		res = append(res, nodeMap)
	}
	return res, nil
}

func (s *WorkflowService) convertEdgesToMap(edges map[uuid.UUID]*wfentities.BaseEdgeData) ([]map[string]any, error) {
	res := make([]map[string]any, 0, len(edges))
	for id, edge := range edges {
		edgeMap, err := s.convertModelToMap(edge)
		if err != nil {
			return nil, err
		}
		edgeMap["id"] = id.String()
		res = append(res, edgeMap)
	}
	return res, nil
}

func (s *WorkflowService) convertModelToMap(obj any) (map[string]any, error) {
	// 1. 如果是 nil，直接返回空 map
	if obj == nil {
		return map[string]any{}, nil
	}

	// 2. 如果是 UUID 类型，转换为字符串
	if id, ok := obj.(uuid.UUID); ok {
		return map[string]any{"id": id.String()}, nil
	}

	// 3. 如果是 Enum 类型（实现了 Stringer 接口），调用 String() 方法
	if enum, ok := obj.(fmt.Stringer); ok {
		return map[string]any{"value": enum.String()}, nil
	}

	// 4. 如果是结构体或指针，使用反射处理
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 5. 如果不是结构体，尝试 JSON 序列化（处理基本类型、切片、map 等）
	if val.Kind() != reflect.Struct {
		data, err := sonic.Marshal(obj)
		if err != nil {
			return nil, err
		}
		var result map[string]any
		if err := sonic.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}

	// 6. 处理结构体字段（递归）
	result := make(map[string]any)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 跳过不可导出字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 获取 JSON 标签（字段名）
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		// 递归处理字段值
		if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map {
			nestedMap, err := s.convertModelToMap(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			result[jsonTag] = nestedMap
		} else {
			result[jsonTag] = fieldValue.Interface()
		}
	}

	return result, nil
}
