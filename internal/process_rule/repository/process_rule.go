package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/process_rule/repository/dao"
)

type ProcessRuleRepo struct {
	dao *dao.ProcessRuleDao
}

func NewProcessRuleRepo(dao *dao.ProcessRuleDao) *ProcessRuleRepo {
	return &ProcessRuleRepo{dao: dao}
}

func (r *ProcessRuleRepo) CreateProcessRule(ctx context.Context, processRule *entity.ProcessRule) error {
	return r.dao.CreateProcessRule(ctx, processRule)
}

func (r *ProcessRuleRepo) GetProcessRuleByID(ctx context.Context, id uuid.UUID) (*entity.ProcessRule, error) {
	return r.dao.GetProcessRuleByID(ctx, id)
}

func (r *ProcessRuleRepo) GetProcessRuleByDatasetID(ctx context.Context, datasetID uuid.UUID) (*entity.ProcessRule, error) {
	return r.dao.GetProcessRuleByDatasetID(ctx, datasetID)
}

func (r *ProcessRuleRepo) UpdateProcessRule(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateProcessRule(ctx, id, updates)
}

func (r *ProcessRuleRepo) DeleteProcessRule(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteProcessRule(ctx, id)
}

func (r *ProcessRuleRepo) GetProcessRulesByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.ProcessRule, error) {
	return r.dao.GetProcessRulesByAccountID(ctx, accountID)
}
