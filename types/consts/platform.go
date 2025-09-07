package consts

// Platform相关常量定义

// WechatConfigStatus 微信配置状态
type WechatConfigStatus string

const (
	WechatConfigStatusConfigured   WechatConfigStatus = "configured"   // 已配置
	WechatConfigStatusUnconfigured WechatConfigStatus = "unconfigured" // 未配置
)
