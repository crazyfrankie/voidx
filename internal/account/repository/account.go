package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/account/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AccountRepo struct {
	dao *dao.AccountDao
}

func NewAccountRepo(d *dao.AccountDao) *AccountRepo {
	return &AccountRepo{dao: d}
}

func (r *AccountRepo) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	return r.dao.GetAccountByID(ctx, id)
}

func (r *AccountRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwd string) error {
	return r.dao.UpdatePassword(ctx, id, passwd)
}

func (r *AccountRepo) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return r.dao.UpdateName(ctx, id, name)
}

func (r *AccountRepo) UpdateAvatar(ctx context.Context, id uuid.UUID, avatar string) error {
	return r.dao.UpdateAvatar(ctx, id, avatar)
}
