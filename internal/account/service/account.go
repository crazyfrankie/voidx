package service

import (
	"context"
	"errors"

	"github.com/crazyfrankie/voidx/types/errno"
	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/voidx/internal/account/repository"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AccountService struct {
	repo *repository.AccountRepo
}

func NewAccountService(repo *repository.AccountRepo) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) GetAccountByID(ctx context.Context) (resp.Account, error) {
	id, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return resp.Account{}, err
	}

	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return resp.Account{}, errno.ErrNotFound.AppendBizMessage(errors.New("用户标识错误"))
	}

	return resp.Account{
		ID:       account.ID,
		Avatar:   account.Avatar,
		Name:     account.Name,
		Email:    account.Email,
		Password: account.Password,
	}, nil
}

func (s *AccountService) UpdatePassword(ctx context.Context, passwd string) error {
	id, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	newPasswd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, id, string(newPasswd))
}

func (s *AccountService) UpdateName(ctx context.Context, name string) error {
	id, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.UpdateName(ctx, id, name)
}

func (s *AccountService) UpdateAvatar(ctx context.Context, avatar string) error {
	id, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.repo.UpdateAvatar(ctx, id, avatar)
}
