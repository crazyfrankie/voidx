package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type ProcessRuleDao struct {
	db *gorm.DB
}

func NewProcessRuleDao(db *gorm.DB) *ProcessRuleDao {
	return &ProcessRuleDao{db: db}
}

func (d *ProcessRuleDao) CreateProcessRule(ctx context.Context, processRule *entity.ProcessRule) error {
	return d.db.WithContext(ctx).Create(processRule).Error
}

func (d *ProcessRuleDao) GetProcessRuleByID(ctx context.Context, id uuid.UUID) (*entity.ProcessRule, error) {
	var processRule entity.ProcessRule
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&processRule).Error
	if err != nil {
		return nil, err
	}
	return &processRule, nil
}

func (d *ProcessRuleDao) GetProcessRuleByDatasetID(ctx context.Context, datasetID uuid.UUID) (*entity.ProcessRule, error) {
	var processRule entity.ProcessRule
	err := d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).First(&processRule).Error
	if err != nil {
		return nil, err
	}
	return &processRule, nil
}

func (d *ProcessRuleDao) UpdateProcessRule(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.ProcessRule{}).Where("id = ?", id).Updates(updates).Error
}

func (d *ProcessRuleDao) DeleteProcessRule(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Delete(&entity.ProcessRule{}, "id = ?", id).Error
}

func (d *ProcessRuleDao) GetProcessRulesByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.ProcessRule, error) {
	var processRules []*entity.ProcessRule
	err := d.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&processRules).Error
	if err != nil {
		return nil, err
	}
	return processRules, nil
}
