package dao

import (
	"context"

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
