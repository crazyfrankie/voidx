package ioc

import "github.com/silenceper/wechat/v2"

func InitWechat() *wechat.Wechat {
	return wechat.NewWechat()
}
