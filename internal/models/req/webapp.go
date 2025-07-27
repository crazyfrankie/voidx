package req

// WebAppChatReq WebApp对话请求
type WebAppChatReq struct {
	ConversationID string          `json:"conversation_id,omitempty"`
	Query          string          `json:"query" binding:"required,max=2000"`
	ImageUrls      []string        `json:"image_urls,omitempty"`
	Files          []WebAppFileReq `json:"files,omitempty"`
	AutoGenerate   bool            `json:"auto_generate,omitempty"`
	ResponseMode   string          `json:"response_mode,omitempty"` // streaming, blocking
}

// WebAppFileReq WebApp文件请求
type WebAppFileReq struct {
	Type           string `json:"type" binding:"required"`
	TransferMethod string `json:"transfer_method" binding:"required"`
	URL            string `json:"url,omitempty"`
	UploadFileID   string `json:"upload_file_id,omitempty"`
}

// GetWebAppConversationsReq 获取WebApp会话列表请求
type GetWebAppConversationsReq struct {
	IsPinned bool `form:"is_pinned"`
}

// GetWebAppConversationMessagesReq 获取WebApp会话消息列表请求
type GetWebAppConversationMessagesReq struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

// UpdateWebAppConversationNameReq 更新WebApp会话名称请求
type UpdateWebAppConversationNameReq struct {
	Name string `json:"name" binding:"required,max=100"`
}

// UpdateWebAppConversationPinReq 更新WebApp会话置顶状态请求
type UpdateWebAppConversationPinReq struct {
	IsPinned bool `json:"is_pinned"`
}

// StopWebAppChatReq 停止WebApp对话请求
type StopWebAppChatReq struct {
	TaskID string `uri:"task_id" binding:"required"`
}
