package req

// GetModelsReq 获取模型列表请求
type GetModelsReq struct {
	Provider string `uri:"provider" binding:"required"`
}

// GetModelsByTypeReq 根据类型获取模型列表请求
type GetModelsByTypeReq struct {
	ModelType string `uri:"model_type" binding:"required"`
}
