package req

// CreateModelReq 创建模型请求
type CreateModelReq struct {
	Provider  string                 `json:"provider" binding:"required"`
	ModelName string                 `json:"model_name" binding:"required"`
	Config    map[string]interface{} `json:"config" binding:"required"`
}

// ValidateModelConfigReq 验证模型配置请求
type ValidateModelConfigReq struct {
	Provider  string                 `json:"provider" binding:"required"`
	ModelName string                 `json:"model_name" binding:"required"`
	Config    map[string]interface{} `json:"config" binding:"required"`
}

// GetModelsReq 获取模型列表请求
type GetModelsReq struct {
	Provider string `uri:"provider" binding:"required"`
}

// GetModelsByTypeReq 根据类型获取模型列表请求
type GetModelsByTypeReq struct {
	ModelType string `uri:"model_type" binding:"required"`
}
