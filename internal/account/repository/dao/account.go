package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AccountDao struct {
	db *gorm.DB
}

func NewAccountDao(db *gorm.DB) *AccountDao {
	return &AccountDao{db: db}
}

func (d *AccountDao) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	var account *entity.Account
	if err := d.db.WithContext(ctx).Model(&entity.Account{}).
		Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (d *AccountDao) UpdatePassword(ctx context.Context, id uuid.UUID, passwd string) error {
	return d.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", id).Update("password", passwd).Error
}

func (d *AccountDao) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return d.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", id).Update("name", name).Error
}

func (d *AccountDao) UpdateAvatar(ctx context.Context, id uuid.UUID, avatar string) error {
	return d.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", id).Update("avatar", avatar).Error
}
