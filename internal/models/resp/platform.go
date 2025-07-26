package resp

import "github.com/google/uuid"

// GetWechatConfigResp 获取微信配置响应
type GetWechatConfigResp struct {
	AppID           uuid.UUID `json:"app_id"`
	URL             string    `json:"url"`
	IP              string    `json:"ip"`
	WechatAppID     string    `json:"wechat_app_id"`
	WechatAppSecret string    `json:"wechat_app_secret"`
	WechatToken     string    `json:"wechat_token"`
	Status          string    `json:"status"`
	Ctime           int64     `json:"ctime"`
	Utime           int64     `json:"utime"`
}
