package dao

import (
	"context"
	"errors"
	"github.com/crazyfrankie/voidx/pkg/consts"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type WorkflowDao struct {
	db *gorm.DB
}

func NewWorkflowDao(db *gorm.DB) *WorkflowDao {
	return &WorkflowDao{db: db}
}

// CreateWorkflow 创建工作流
func (d *WorkflowDao) CreateWorkflow(ctx context.Context, workflow *entity.Workflow) error {
	return d.db.WithContext(ctx).Create(workflow).Error
}

// GetWorkflowByID 根据ID获取工作流
func (d *WorkflowDao) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	var workflow entity.Workflow
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// GetWorkflowByToolCallName 根据工具调用名称获取工作流
func (d *WorkflowDao) GetWorkflowByToolCallName(ctx context.Context, accountID uuid.UUID, toolCallName string) (*entity.Workflow, error) {
	var workflow entity.Workflow
	err := d.db.WithContext(ctx).
		Where("account_id = ? AND tool_call_name = ?", accountID, toolCallName).
		First(&workflow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

// GetWorkflowsByAccountID 根据账户ID获取工作流列表
func (d *WorkflowDao) GetWorkflowsByAccountID(ctx context.Context, accountID uuid.UUID,
	pageReq req.GetWorkflowsWithPageReq) ([]entity.Workflow, int64, error) {
	query := d.db.WithContext(ctx).Where("account_id = ?", accountID)

	// 添加筛选条件
	if pageReq.SearchWord != "" {
		query = query.Where("name ILIKE ?", "%"+pageReq.SearchWord+"%")
	}
	if pageReq.Status != "" {
		query = query.Where("status = ?", pageReq.Status)
	}

	// 获取总数
	var total int64
	if err := query.Model(&entity.Workflow{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var workflows []entity.Workflow
	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").Offset(offset).Limit(pageReq.PageSize).Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

// UpdateWorkflow 更新工作流
func (d *WorkflowDao) UpdateWorkflow(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Workflow{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteWorkflow 删除工作流
func (d *WorkflowDao) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Workflow{}).Error
}

// CreateWorkflowResult 创建工作流运行结果
func (d *WorkflowDao) CreateWorkflowResult(ctx context.Context, result *entity.WorkflowResult) error {
	return d.db.WithContext(ctx).Create(result).Error
}

// UpdateWorkflowResult 更新工作流运行结果
func (d *WorkflowDao) UpdateWorkflowResult(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.WorkflowResult{}).Where("id = ?", id).Updates(updates).Error
}

// GetWorkflowResultsByWorkflowID 根据工作流ID获取运行结果列表
func (d *WorkflowDao) GetWorkflowResultsByWorkflowID(ctx context.Context, workflowID uuid.UUID,
	page, pageSize int) ([]entity.WorkflowResult, int64, error) {
	query := d.db.WithContext(ctx).Where("workflow_id = ?", workflowID)

	// 获取总数
	var total int64
	if err := query.Model(&entity.WorkflowResult{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var results []entity.WorkflowResult
	offset := (page - 1) * pageSize
	err := query.Order("ctime DESC").Offset(offset).Limit(pageSize).Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (d *WorkflowDao) GetDatasets(ctx context.Context, accountID uuid.UUID, datasetIDs []uuid.UUID) ([]entity.Dataset, error) {
	var datasets []entity.Dataset

	err := d.db.WithContext(ctx).
		Where("id IN (?) AND account_id = ?", datasets, accountID).
		Find(&datasets).Error
	if err != nil {
		return nil, err
	}

	return datasets, nil
}

func (d *WorkflowDao) GetWorkflows(ctx context.Context, workflowIDs []uuid.UUID, accountId uuid.UUID) ([]entity.Workflow, error) {
	var workflows []entity.Workflow
	err := d.db.WithContext(ctx).
		Where("id IN (?) AND account_id = ? AND status = ?", workflowIDs, accountId, consts.WorkflowStatusPublished).
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	return workflows, nil
}
