package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AuthDao struct {
	db *gorm.DB
}

func NewAuthDao(db *gorm.DB) *AuthDao {
	return &AuthDao{db: db}
}

func (d *AuthDao) CreateAccount(ctx context.Context, account *entity.Account) (entity.Account, error) {
	if err := d.db.WithContext(ctx).Model(&entity.Account{}).Create(&account).Error; err != nil {
		return entity.Account{}, err
	}

	return *account, nil
}

func (d *AuthDao) GetAccountByEmail(ctx context.Context, email string) (entity.Account, error) {
	var account entity.Account
	err := d.db.WithContext(ctx).Model(&entity.Account{}).
		Where("email = ?", email).Select([]string{"id", "password"}).Find(&account).Error
	if err != nil {
		return entity.Account{}, err
	}

	return account, nil
}

func (d *AuthDao) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	var account *entity.Account
	if err := d.db.WithContext(ctx).Model(&entity.Account{}).
		Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (d *AuthDao) UpdatePassword(ctx context.Context, id uuid.UUID, passwd string) error {
	return d.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", id).Update("password", passwd).Error
}

func (d *AuthDao) UpdateAccount(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", id).Updates(updates).Error
}
