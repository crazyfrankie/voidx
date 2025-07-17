package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/voidx/internal/account/repository"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AccountService struct {
	repo *repository.AccountRepo
}

func NewAccountService(repo *repository.AccountRepo) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) GetAccountByID(ctx context.Context, id uuid.UUID) (resp.Account, error) {
	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return resp.Account{}, errno.ErrNotFound.AppendBizMessage("用户标识错误")
	}

	return resp.Account{
		ID:       account.ID,
		Avatar:   account.Avatar,
		Name:     account.Name,
		Email:    account.Email,
		Password: account.Password,
	}, nil
}

func (s *AccountService) UpdatePassword(ctx context.Context, id uuid.UUID, passwd string) error {
	newPasswd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(newPasswd))
}

func (s *AccountService) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdateName(ctx, id, name)
}

func (s *AccountService) UpdateAvatar(ctx context.Context, id uuid.UUID, avatar string) error {
	return s.repo.UpdateAvatar(ctx, id, avatar)
}
