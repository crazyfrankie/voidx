package repository

import (
	"context"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/oauth/repository/dao"
)

type OAuthRepo struct {
	dao *dao.OAuthDao
}

func NewOAuthRepo(dao *dao.OAuthDao) *OAuthRepo {
	return &OAuthRepo{dao: dao}
}

func (r *OAuthRepo) GetAccountOAuthByOpenID(ctx context.Context, providerName string, openID string) (*entity.AccountOAuth, error) {
	return r.dao.GetAccountOAuthByOpenID(ctx, providerName, openID)
}

func (r *OAuthRepo) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	return r.dao.GetAccountByEmail(ctx, email)
}

func (r *OAuthRepo) GetAccountByID(ctx context.Context, accountID uuid.UUID) (*entity.Account, error) {
	return r.dao.GetAccountByID(ctx, accountID)
}

func (r *OAuthRepo) CreateAccount(ctx context.Context, account *entity.Account) (*entity.Account, error) {
	return r.dao.CreateAccount(ctx, account)
}

func (r *OAuthRepo) CreateAccountOAuth(ctx context.Context, accountOAuth *entity.AccountOAuth) error {
	return r.dao.CreateAccountOAuth(ctx, accountOAuth)
}

func (r *OAuthRepo) UpdateAccountInfo(ctx context.Context, accountID uuid.UUID, account *entity.Account, accountOAuth *entity.AccountOAuth) error {
	
}
