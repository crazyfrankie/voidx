package oauth

type OAuthUserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// OAuth 接口对应Python中的抽象类OAuth
type OAuth interface {
	GetProvider() string
	GetAuthorizationURL() string
	GetAccessToken(code string) (string, error)
	GetRawUserInfo(token string) (map[string]interface{}, error)
	TransformUserInfo(rawInfo map[string]interface{}) (*OAuthUserInfo, error)
	GetUserInfo(token string) (*OAuthUserInfo, error)
}

type BaseOAuth struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}
