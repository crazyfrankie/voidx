package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/voidx/internal/auth/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AuthService struct {
	repo  *repository.AuthRepo
	token *jwt.TokenService
}

func NewAuthService(repo *repository.AuthRepo, token *jwt.TokenService) *AuthService {
	return &AuthService{repo: repo, token: token}
}

func (s *AuthService) Login(ctx context.Context, ua string, loginReq req.LoginReq) ([]string, error) {
	account, err := s.repo.GetAccountByEmail(ctx, loginReq.Email)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("账号不存在或者密码错误，请核实后重试")
	}

	var uid uuid.UUID
	// create user
	if account.ID == uuid.Nil {
		if hashPasswd, err := bcrypt.GenerateFromPassword([]byte(loginReq.Password), bcrypt.DefaultCost); err == nil {
			newUser, err := s.repo.CreateAccount(ctx, &entity.Account{
				Email:    loginReq.Email,
				Password: string(hashPasswd),
			})
			if err != nil {
				return nil, err
			}
			uid = newUser.ID
		} else {
			return nil, err
		}
	} else {
		uid = account.ID
		if err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(loginReq.Password)); err != nil {
			return nil, errno.ErrNotFound.AppendBizMessage("账号不存在或者密码错误，请核实后重试")
		}
	}

	return s.token.GenerateToken(uid, ua)
}

func (s *AuthService) Logout(ctx context.Context, ua string) error {
	uid, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.token.CleanToken(ctx, uid, ua)
}
