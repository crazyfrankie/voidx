package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type OAuthDao struct {
	db *gorm.DB
}

func NewOAuthDao(db *gorm.DB) *OAuthDao {
	return &OAuthDao{db: db}
}

func (d *OAuthDao) GetAccountOAuthByOpenID(ctx context.Context, providerName string, openID string) (*entity.AccountOAuth, error) {
	var accountOAuth *entity.AccountOAuth
	if err := d.db.WithContext(ctx).Model(&entity.AccountOAuth{}).
		Where("open_id = ? AND provider = ?", openID, providerName).
		First(&accountOAuth).Error; err != nil {
		return nil, err
	}

	return accountOAuth, nil
}

func (d *OAuthDao) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	var account *entity.Account
	if err := d.db.WithContext(ctx).Model(&entity.Account{}).
		Where("email = ?", email).
		First(&account).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (d *OAuthDao) GetAccountByID(ctx context.Context, accountID uuid.UUID) (*entity.Account, error) {
	var account *entity.Account
	if err := d.db.WithContext(ctx).Model(&entity.Account{}).
		Where("id = ?", accountID).
		First(&account).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (d *OAuthDao) CreateAccount(ctx context.Context, account *entity.Account) (*entity.Account, error) {
	err := d.db.WithContext(ctx).Model(&entity.Account{}).Create(account).Error
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (d *OAuthDao) CreateAccountOAuth(ctx context.Context, auth *entity.AccountOAuth) error {
	return d.db.WithContext(ctx).Model(&entity.AccountOAuth{}).Create(auth).Error
}

func (d *OAuthDao) UpdateAccountInfo(ctx context.Context, accountID uuid.UUID, account *entity.Account, accountOAuth *entity.AccountOAuth) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Account{}).Where("id = ?", accountID).Save(account).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.AccountOAuth{}).Where("account_id = ?", accountID).Save(accountOAuth).Error; err != nil {
			return err
		}

		return nil
	})
}
