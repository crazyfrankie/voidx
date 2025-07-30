package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/workflow/repository/dao"
)

type WorkflowRepo struct {
	dao *dao.WorkflowDao
}

func NewWorkflowRepo(d *dao.WorkflowDao) *WorkflowRepo {
	return &WorkflowRepo{dao: d}
}

// CreateWorkflow 创建工作流
func (r *WorkflowRepo) CreateWorkflow(ctx context.Context, workflow *entity.Workflow) error {
	return r.dao.CreateWorkflow(ctx, workflow)
}

// GetWorkflowByID 根据ID获取工作流
func (r *WorkflowRepo) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	return r.dao.GetWorkflowByID(ctx, id)
}

// GetWorkflowByToolCallName 根据工具调用名称获取工作流
func (r *WorkflowRepo) GetWorkflowByToolCallName(ctx context.Context, accountID uuid.UUID, toolCallName string) (*entity.Workflow, error) {
	return r.dao.GetWorkflowByToolCallName(ctx, accountID, toolCallName)
}

// GetWorkflowsByAccountID 根据账户ID获取工作流列表
func (r *WorkflowRepo) GetWorkflowsByAccountID(ctx context.Context, accountID uuid.UUID, pageReq req.GetWorkflowsWithPageReq) ([]entity.Workflow, int64, error) {
	return r.dao.GetWorkflowsByAccountID(ctx, accountID, pageReq)
}

// UpdateWorkflow 更新工作流
func (r *WorkflowRepo) UpdateWorkflow(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateWorkflow(ctx, id, updates)
}

// DeleteWorkflow 删除工作流
func (r *WorkflowRepo) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteWorkflow(ctx, id)
}

// CreateWorkflowResult 创建工作流运行结果
func (r *WorkflowRepo) CreateWorkflowResult(ctx context.Context, result *entity.WorkflowResult) error {
	return r.dao.CreateWorkflowResult(ctx, result)
}

// UpdateWorkflowResult 更新工作流运行结果
func (r *WorkflowRepo) UpdateWorkflowResult(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateWorkflowResult(ctx, id, updates)
}

// GetWorkflowResultsByWorkflowID 根据工作流ID获取运行结果列表
func (r *WorkflowRepo) GetWorkflowResultsByWorkflowID(ctx context.Context, workflowID uuid.UUID, page, pageSize int) ([]entity.WorkflowResult, int64, error) {
	return r.dao.GetWorkflowResultsByWorkflowID(ctx, workflowID, page, pageSize)
}

// GetDatasets 剔除关联知识库列表中不属于当前账户的数据
func (r *WorkflowRepo) GetDatasets(ctx context.Context, accountID uuid.UUID, datasetIDs []uuid.UUID) ([]entity.Dataset, error) {
	return r.dao.GetDatasets(ctx, accountID, datasetIDs)
}

// GetWorkflows 剔除不属于当前账户并且未发布的工作流
func (r *WorkflowRepo) GetWorkflows(ctx context.Context, workflowIDs []uuid.UUID, accountId uuid.UUID) ([]entity.Workflow, error) {
	return r.dao.GetWorkflows(ctx, workflowIDs, accountId)
}
