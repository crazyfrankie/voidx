package resp

import "github.com/google/uuid"

// GetWechatConfigResp 获取微信配置响应
//
//	app_id = fields.UUID(dump_default="")
//	  url = fields.String(dump_default="")
//	  ip = fields.String(dump_default="")
//	  wechat_app_id = fields.String(dump_default="")
//	  wechat_app_secret = fields.String(dump_default="")
//	  wechat_token = fields.String(dump_default="")
//	  status = fields.String(dump_default="")
//	  updated_at = fields.Integer(dump_default=0)
//	  created_at = fields.Integer(dump_default=0)
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
