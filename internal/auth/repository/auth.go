package repository

import (
	"context"

	"github.com/google/uuid"

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

func (r *AuthRepo) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	return r.dao.GetAccountByID(ctx, id)
}

func (r *AuthRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwd string) error {
	return r.dao.UpdatePassword(ctx, id, passwd)
}

func (r *AuthRepo) UpdateAccount(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateAccount(ctx, id, updates)
}
