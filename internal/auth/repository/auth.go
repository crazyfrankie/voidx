package repository

import (
	"context"

	"github.com/crazyfrankie/voidx/internal/auth/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AuthRepo struct {
	dao *dao.AuthDao
}

func NewAuthRepo(d *dao.AuthDao) *AuthRepo {
	return &AuthRepo{dao: d}
}

func (r *AuthRepo) CreateAccount(ctx context.Context, account *entity.Account) (entity.Account, error) {
	return r.dao.CreateAccount(ctx, account)
}

func (r *AuthRepo) GetAccountByEmail(ctx context.Context, email string) (entity.Account, error) {
	return r.dao.GetAccountByEmail(ctx, email)
}
