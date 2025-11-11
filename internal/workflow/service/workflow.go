package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	corewf "github.com/crazyfrankie/voidx/internal/core/workflow"
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/workflow/repository"
	"github.com/crazyfrankie/voidx/pkg/sonic"
	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
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
		return nil, errno.ErrValidate.AppendBizMessage(fmt.Errorf("在当前账号下已创建[%s]工作流，不支持重名", createReq.ToolCallName))
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
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
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
		return errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	return s.repo.DeleteWorkflow(ctx, workflowID)
}

// UpdateWorkflow 根据传递的工作流id+请求更新工作流基础信息
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, workflowID, userID uuid.UUID, updateReq req.UpdateWorkflowReq) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	// 检查工具调用名称是否重复
	toolCallName := strings.TrimSpace(updateReq.ToolCallName)
	existing, err := s.repo.GetWorkflowByToolCallName(ctx, userID, toolCallName)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != workflowID {
		return errno.ErrValidate.AppendBizMessage(fmt.Errorf("在当前账号下已创建[%s]工作流，不支持重名", toolCallName))
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
		return errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
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
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	validateDraftGraph, err := s.validateGraph(ctx, workflowID, workflow.DraftGraph, userID)
	if err != nil {
		return nil, err
	}

	for _, node := range validateDraftGraph["nodes"].([]*entities.BaseNodeData) {
		if node.NodeType == entities.NodeTypeTool {

		}
	}

	return validateDraftGraph, nil
}

// DebugWorkflow 调试指定的工作流API接口，该接口为流式事件输出
func (s *WorkflowService) DebugWorkflow(ctx context.Context, workflowID, userID uuid.UUID, inputs map[string]any) (<-chan resp.WorkflowDebugEvent, error) {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	workflowTool, err := corewf.NewWorkflow(map[string]any{
		"account_id":  userID,
		"name":        workflow.ToolCallName,
		"description": workflow.Description,
		"nodes":       workflow.DraftGraph["nodes"],
		"edges":       workflow.DraftGraph["edges"],
	})

	// 创建事件通道
	eventChan := make(chan resp.WorkflowDebugEvent, 100)

	// 启动异步处理
	go s.processWorkflowDebug(ctx, workflow, workflowTool, inputs, eventChan)

	return eventChan, nil
}

// PublishWorkflow 根据传递的工作流id，发布指定的工作流
func (s *WorkflowService) PublishWorkflow(ctx context.Context, workflowID, userID uuid.UUID) error {
	workflow, err := s.repo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	if !workflow.IsDebugPassed {
		return errno.ErrValidate.AppendBizMessage(errors.New("该工作流未调试通过，请调试通过后发布"))
	}

	if _, err := corewf.NewWorkflow(map[string]any{
		"account_id":  userID,
		"name":        workflow.Name,
		"description": workflow.Description,
		"nodes":       workflow.DraftGraph["nodes"],
		"edges":       workflow.DraftGraph["edges"],
	}); err != nil {
		if err := s.repo.UpdateWorkflow(ctx, workflowID, map[string]any{
			"is_debug_passed": false,
		}); err != nil {
			return errno.ErrInternalServer.AppendBizMessage(fmt.Errorf("保存工作流状态失败: %v", err))
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
		return errno.ErrNotFound.AppendBizMessage(errors.New("该工作流不存在，请核实后重试"))
	}

	if workflow.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	if workflow.Status != consts.WorkflowStatusPublished {
		return errno.ErrValidate.AppendBizMessage(errors.New("该工作流未发布无法取消发布"))
	}

	emptyGraphJSON, _ := sonic.Marshal(map[string]any{})
	updates := map[string]any{
		"graph":           emptyGraphJSON,
		"status":          consts.WorkflowStatusDraft,
		"is_debug_passed": false,
	}

	return s.repo.UpdateWorkflow(ctx, workflowID, updates)
}

// validateGraph 校验传递的graph信息，涵盖nodes和edges对应的数据
func (s *WorkflowService) validateGraph(ctx context.Context, workflowID uuid.UUID, graph map[string]any, accountID uuid.UUID) (map[string]any, error) {
	// 1. 提取 nodes 和 edges 数据
	nodes, _ := graph["nodes"].([]any)
	edges, _ := graph["edges"].([]any)

	// 2. 循环校验 nodes 中各个节点对应的数据
	nodeDataDict := make(map[uuid.UUID]*entities.BaseNodeData)
	startNodes := 0
	endNodes := 0

	for _, nodeInterface := range nodes {
		nodeMap, ok := nodeInterface.(map[string]any)
		if !ok {
			continue // 跳过无效节点
		}

		// 解析节点数据
		nodeData, err := s.parseNodeFromMap(nodeMap)
		if err != nil {
			continue // 跳过解析失败的节点
		}

		// 判断节点 id 是否唯一
		if _, exists := nodeDataDict[nodeData.ID]; exists {
			return nil, fmt.Errorf("工作流节点id必须唯一，请核实后重试")
		}

		// 判断节点 title 是否唯一
		for _, existingNode := range nodeDataDict {
			if strings.TrimSpace(existingNode.Title) == strings.TrimSpace(nodeData.Title) {
				return nil, fmt.Errorf("工作流节点title必须唯一，请核实后重试")
			}
		}

		// 对特殊节点进行判断
		switch nodeData.NodeType {
		case entities.NodeTypeStart:
			if startNodes >= 1 {
				return nil, fmt.Errorf("工作流中只允许有1个开始节点")
			}
			startNodes++
		case entities.NodeTypeEnd:
			if endNodes >= 1 {
				return nil, fmt.Errorf("工作流中只允许有1个结束节点")
			}
			endNodes++
		}

		nodeDataDict[nodeData.ID] = nodeData
	}

	// 3. 循环校验 edges 中各个节点对应的数据
	edgeDataDict := make(map[uuid.UUID]*entities.BaseEdgeData)
	for _, edgeInterface := range edges {
		edgeMap, ok := edgeInterface.(map[string]any)
		if !ok {
			continue // 跳过无效边
		}

		edgeData, err := s.parseEdgeFromMap(edgeMap)
		if err != nil {
			continue // 跳过解析失败的边
		}

		// 校验边 edges 的 id 是否唯一
		if _, exists := edgeDataDict[edgeData.ID]; exists {
			continue // 跳过重复边
		}

		// 校验边中的 source/target/source_type/target_type 必须和 nodes 对得上
		sourceNode, sourceExists := nodeDataDict[edgeData.Source]
		targetNode, targetExists := nodeDataDict[edgeData.Target]
		if !sourceExists || !targetExists ||
			edgeData.SourceType != sourceNode.NodeType ||
			edgeData.TargetType != targetNode.NodeType {
			continue // 跳过无效边
		}

		// 校验边 Edges 里的边必须唯一
		isDuplicate := false
		for _, existingEdge := range edgeDataDict {
			if existingEdge.Source == edgeData.Source &&
				existingEdge.Target == edgeData.Target &&
				((existingEdge.SourceHandleID == nil && edgeData.SourceHandleID == nil) ||
					(existingEdge.SourceHandleID != nil && edgeData.SourceHandleID != nil &&
						*existingEdge.SourceHandleID == *edgeData.SourceHandleID)) {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue // 跳过重复边
		}

		edgeDataDict[edgeData.ID] = edgeData
	}

	// 转换结果
	resNodes := make([]*entities.BaseNodeData, 0, len(nodeDataDict))
	for _, nodeData := range nodeDataDict {
		resNodes = append(resNodes, nodeData)
	}

	resEdges := make([]*entities.BaseEdgeData, 0, len(edgeDataDict))
	for _, edgeData := range edgeDataDict {
		resEdges = append(resEdges, edgeData)
	}

	result := map[string]any{
		"nodes": resNodes,
		"edges": resEdges,
	}

	return result, nil
}

// parseNodeFromMap 从 map 解析节点数据
func (s *WorkflowService) parseNodeFromMap(nodeMap map[string]any) (*entities.BaseNodeData, error) {
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

// parseEdgeFromMap 从 map 解析边数据
func (s *WorkflowService) parseEdgeFromMap(edgeMap map[string]any) (*entities.BaseEdgeData, error) {
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

// processWorkflowDebug 处理工作流调试
func (s *WorkflowService) processWorkflowDebug(ctx context.Context, workflow *entity.Workflow, workflowTool *corewf.Workflow,
	inputs map[string]any, eventChan chan<- resp.WorkflowDebugEvent) {
	defer close(eventChan)

	var nodeResults []map[string]any

	// 创建工作流运行结果记录
	workflowResult := &entity.WorkflowResult{
		AccountID:  workflow.AccountID,
		WorkflowID: workflow.ID,
		Graph:      workflow.DraftGraph,
		Status:     consts.WorkflowResultStatusRunning,
	}

	if err := s.repo.CreateWorkflowResult(ctx, workflowResult); err != nil {
		eventChan <- resp.WorkflowDebugEvent{
			Error: errno.ErrInternalServer.AppendBizMessage(errors.New("创建工作流结果记录失败")).Error(),
		}
		return
	}

	startTime := time.Now()

	defer func() {
		// 在函数结束时更新工作流结果和工作流状态
		latency := time.Since(startTime).Milliseconds()

		updateFields := map[string]any{
			"state":   nodeResults,
			"latency": latency,
		}

		if err := s.repo.UpdateWorkflowResult(ctx, workflowResult.ID, updateFields); err != nil {
			// 记录错误但不中断流程
		}

		if err := s.repo.UpdateWorkflow(ctx, workflow.ID, map[string]any{
			"is_debug_passed": workflowResult.Status == consts.WorkflowResultStatusSucceeded,
		}); err != nil {
			// 记录错误但不中断流程
		}
	}()

	// 执行工作流
	streamChan, err := workflowTool.Stream(ctx, inputs)
	if err != nil {
		workflowResult.Status = consts.WorkflowResultStatusFailed
		eventChan <- resp.WorkflowDebugEvent{
			Error: errno.ErrInternalServer.AppendBizMessage(errors.New("创建工作流流式处理失败")).Error(),
		}
		return
	}

	// 处理流式结果
	for nodeResult := range streamChan {
		// 转换为调试事件
		debugEvent := resp.WorkflowDebugEvent{
			ID:          uuid.New().String(),
			NodeID:      nodeResult.NodeID.String(),
			NodeType:    string(nodeResult.NodeType),
			Title:       nodeResult.NodeID.String(), // 可以根据需要调整
			Status:      string(nodeResult.Status),
			Inputs:      nodeResult.Inputs,
			Outputs:     nodeResult.Outputs,
			ElapsedTime: float64(time.Since(startTime).Milliseconds()) / 1000.0,
		}

		if nodeResult.Error != "" {
			debugEvent.Error = nodeResult.Error
		}

		// 记录节点结果
		nodeResultMap := map[string]any{
			"node_id":      nodeResult.NodeID.String(),
			"node_type":    string(nodeResult.NodeType),
			"status":       string(nodeResult.Status),
			"inputs":       nodeResult.Inputs,
			"outputs":      nodeResult.Outputs,
			"elapsed_time": debugEvent.ElapsedTime,
		}
		if nodeResult.Error != "" {
			nodeResultMap["error"] = nodeResult.Error
		}
		nodeResults = append(nodeResults, nodeResultMap)

		select {
		case eventChan <- debugEvent:
		case <-ctx.Done():
			return
		}
	}

	// 更新工作流运行结果
	latency := time.Since(startTime).Seconds()
	stateJSON, _ := sonic.Marshal(nodeResults)
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
