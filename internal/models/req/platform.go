package req

// UpdateWechatConfigReq 更新微信配置请求
type UpdateWechatConfigReq struct {
	WechatAppID     string `json:"wechat_app_id"`
	WechatAppSecret string `json:"wechat_app_secret"`
	WechatToken     string `json:"wechat_token"`
}
