package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GithubOAuth 实现
type GithubOAuth struct {
	BaseOAuth
}

const (
	githubAuthorizeURL   = "https://github.com/login/oauth/authorize"
	githubAccessTokenURL = "https://github.com/login/oauth/access_token"
	githubUserInfoURL    = "https://api.github.com/user"
	githubEmailInfoURL   = "https://api.github.com/user/emails"
)

func NewGithubOAuth(clientID, clientSecret, redirectURI string) *GithubOAuth {
	return &GithubOAuth{
		BaseOAuth: BaseOAuth{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
		},
	}
}

func (g *GithubOAuth) GetProvider() string {
	return "github"
}

func (g *GithubOAuth) GetAuthorizationURL() string {
	params := url.Values{}
	params.Add("client_id", g.ClientID)
	params.Add("redirect_uri", g.RedirectURI)
	params.Add("scope", "user:email")
	return fmt.Sprintf("%s?%s", githubAuthorizeURL, params.Encode())
}

func (g *GithubOAuth) GetAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Add("client_id", g.ClientID)
	data.Add("client_secret", g.ClientSecret)
	data.Add("code", code)
	data.Add("redirect_uri", g.RedirectURI)

	req, err := http.NewRequest("POST", githubAccessTokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.URL.RawQuery = data.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("Github OAuth授权失败: %s", result.Error)
	}

	return result.AccessToken, nil
}

func (g *GithubOAuth) GetRawUserInfo(token string) (map[string]interface{}, error) {
	// 获取用户基本信息
	userInfo, err := g.getUserInfo(token)
	if err != nil {
		return nil, err
	}

	// 获取用户邮箱
	emailInfo, err := g.getEmailInfo(token)
	if err != nil {
		return nil, err
	}

	// 合并结果
	userInfo["email"] = emailInfo
	return userInfo, nil
}

func (g *GithubOAuth) TransformUserInfo(rawInfo map[string]interface{}) (*OAuthUserInfo, error) {
	// 提取邮箱，如果不存在设置一个默认邮箱
	email, ok := rawInfo["email"].(string)
	if !ok || email == "" {
		id, _ := rawInfo["id"].(float64)
		login, _ := rawInfo["login"].(string)
		email = fmt.Sprintf("%.0f+%s@user.no-reply.github.com", id, login)
	}

	// 提取名称
	name, ok := rawInfo["name"].(string)
	if !ok {
		name = ""
	}

	// 提取ID
	id, ok := rawInfo["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user id")
	}

	return &OAuthUserInfo{
		ID:    fmt.Sprintf("%.0f", id),
		Name:  name,
		Email: email,
	}, nil
}

func (g *GithubOAuth) GetUserInfo(token string) (*OAuthUserInfo, error) {
	rawInfo, err := g.GetRawUserInfo(token)
	if err != nil {
		return nil, err
	}
	return g.TransformUserInfo(rawInfo)
}

func (g *GithubOAuth) getUserInfo(token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", githubUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (g *GithubOAuth) getEmailInfo(token string) (string, error) {
	req, err := http.NewRequest("GET", githubEmailInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emailInfo []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&emailInfo); err != nil {
		return "", err
	}

	// 查找主邮箱
	for _, email := range emailInfo {
		if primary, ok := email["primary"].(bool); ok && primary {
			if emailStr, ok := email["email"].(string); ok {
				return emailStr, nil
			}
		}
	}

	return "", fmt.Errorf("no primary email found")
}
