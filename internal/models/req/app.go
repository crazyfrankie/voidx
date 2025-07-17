package req

import (
	"github.com/google/uuid"
)

// CreateAppReq 创建应用请求
type CreateAppReq struct {
	FileList []struct {
		Uid     string `json:"uid"`
		File    any    `json:"file"`
		Url     string `json:"url"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		Percent int    `json:"percent"`
	} `json:"fileList"`
	Name        string `json:"name" binding:"required,max=100"`
	Icon        string `json:"icon" binding:"required"`
	Description string `json:"description" binding:"required,max=500"`
}

type T struct {
	FileList []struct {
	} `json:"fileList"`
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateAppReq 更新应用请求
type UpdateAppReq struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Icon        string `json:"icon"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// GetAppsWithPageReq 获取应用分页列表请求
type GetAppsWithPageReq struct {
	CurrentPage int    `form:"current_page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
	SearchWord  string `form:"search_word"`
}

// FallbackHistoryToDraftReq 回退历史配置到草稿请求
type FallbackHistoryToDraftReq struct {
	AppConfigVersionID uuid.UUID `json:"app_config_version_id" binding:"required"`
}

// GetPublishHistoriesWithPageReq 获取发布历史分页列表请求
type GetPublishHistoriesWithPageReq struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// UpdateDebugConversationSummaryReq 更新调试会话长期记忆请求
type UpdateDebugConversationSummaryReq struct {
	Summary string `json:"summary" binding:"required"`
}

// DebugChatReq 调试对话请求
type DebugChatReq struct {
	Query     string   `json:"query" binding:"required,max=2000"`
	ImageUrls []string `json:"image_urls"`
}

// GetDebugConversationMessagesWithPageReq 获取调试会话消息分页列表请求
type GetDebugConversationMessagesWithPageReq struct {
	Page     int   `form:"page" binding:"required,min=1"`
	PageSize int   `form:"page_size" binding:"required,min=1,max=100"`
	Ctime    int64 `form:"ctime"` // 时间戳，用于游标分页
}
