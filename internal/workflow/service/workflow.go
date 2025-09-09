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
