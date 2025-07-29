package req

// CreateApiKeyReq 创建API秘钥请求
type CreateApiKeyReq struct {
	IsActive bool   `json:"is_active"`
	Remark   string `json:"remark" binding:"max=100"`
}

// UpdateApiKeyReq 更新API秘钥请求
type UpdateApiKeyReq struct {
	IsActive bool   `json:"is_active"`
	Remark   string `json:"remark" binding:"max=100"`
}

// UpdateApiKeyIsActiveReq 更新API秘钥激活请求
type UpdateApiKeyIsActiveReq struct {
	IsActive bool `json:"is_active"`
}

// GetApiKeysWithPageReq 获取API秘钥分页列表请求
type GetApiKeysWithPageReq struct {
	CurrentPage int `form:"current_page" binding:"min=1"`
	PageSize    int `form:"page_size" binding:"min=1,max=100"`
}
