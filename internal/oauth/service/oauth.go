package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/oauth/repository"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/crazyfrankie/voidx/pkg/oauth"
)

type OAuthService struct {
	repo     *repository.OAuthRepo
	token    *jwt.TokenService
	oauthMap map[string]oauth.OAuth
}

func NewOAuthService(repo *repository.OAuthRepo, token *jwt.TokenService) *OAuthService {
	return &OAuthService{
		repo:  repo,
		token: token,
		oauthMap: map[string]oauth.OAuth{
			"github": &oauth.GithubOAuth{
				BaseOAuth: oauth.BaseOAuth{
					ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
					ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
					RedirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
				},
			},
		},
	}
}

func (s *OAuthService) GetOAuthByProviderName(ctx context.Context, providerName string) (oauth.OAuth, error) {
	oAuth := s.oauthMap[providerName]
	if oAuth == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(fmt.Sprintf("该授权方式 %s 不存在", providerName))
	}

	return oAuth, nil
}

func (s *OAuthService) OAuthLogin(ctx context.Context, providerName string, code string, ua string) ([]string, error) {
	// 1.根据传递的provider_name获取oauth
	oAuth := s.oauthMap[providerName]

	// 2.根据code从第三方登录服务中获取access_token
	accessToken, err := oAuth.GetAccessToken(code)
	if err != nil {
		return nil, err
	}

	// 3.根据获取到的token提取user_info信息
	oAuthUserInfo, err := oAuth.GetUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	// 4.根据provider_name+openid获取授权记录
	accountAuth, err := s.repo.GetAccountOAuthByOpenID(ctx, providerName, oAuthUserInfo.ID)
	if err != nil {
		return nil, err
	}

	var account *entity.Account
	if accountAuth == nil {
		// 5.该授权认证方式是第一次登录，查询邮箱是否存在
		account, err = s.repo.GetAccountByEmail(ctx, oAuthUserInfo.Email)
		if err != nil {
			return nil, err
		}
		// 6.账号不存在，注册账号
		if account == nil {
			if account, err = s.repo.CreateAccount(ctx, &entity.Account{
				Name:  oAuthUserInfo.Name,
				Email: oAuthUserInfo.Email,
			}); err != nil {
				return nil, err
			}
		}
		// 7.添加授权认证记录
		if err := s.repo.CreateAccountOAuth(ctx, &entity.AccountOAuth{
			AccountID:      account.ID,
			Provider:       providerName,
			OpenID:         oAuthUserInfo.ID,
			EncryptedToken: accessToken,
		}); err != nil {
			return nil, err
		}
	} else {
		// 8.查找账号信息
		account, err = s.repo.GetAccountByID(ctx, accountAuth.AccountID)
		if err != nil {
			return nil, err
		}
	}

	// 9.更新账号信息，涵盖最后一次登录时间，以及ip地址
	if err := s.repo.UpdateAccountInfo(ctx, account.ID, &entity.Account{
		LastLoginAt: time.Now().Unix(),
		LastLoginIP: s.getLastLoginIP(ctx),
	}, &entity.AccountOAuth{
		EncryptedToken: accessToken,
	}); err != nil {
		return nil, err
	}

	tokens, err := s.token.GenerateToken(account.ID, ua)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *OAuthService) getLastLoginIP(ctx context.Context) string {
	ip := ctx.Value("last_login_ip")
	if res, ok := ip.(string); ok {
		return res
	} else {
		return ""
	}
}
