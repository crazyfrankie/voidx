package req

import "github.com/crazyfrankie/voidx/internal/models/entity"

// UpdateApiToolProviderReq 更新API工具提供商请求
type UpdateApiToolProviderReq struct {
	Name          string          `json:"name" binding:"omitempty,max=100"`
	Icon          string          `json:"icon" binding:"omitempty,url"`
	Headers       []entity.Header `json:"headers" binding:"omitempty,max=500"`
	OpenAPISchema string          `json:"openapi_schema" binding:"omitempty"`
}

// GetApiToolProvidersWithPageReq 获取API工具提供商分页列表请求
type GetApiToolProvidersWithPageReq struct {
	CurrentPage int    `form:"current_page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
	SearchWord  string `form:"search_word"`
}

// CreateApiToolReq 创建API工具请求
type CreateApiToolReq struct {
	Name          string          `json:"name" binding:"required,max=100"`
	OpenAPISchema string          `json:"openapi_schema" binding:"omitempty"`
	Icon          string          `json:"icon"`
	Headers       []entity.Header `json:"headers"`
}

type ValidateOpenApiSchemaReq struct {
	OpenApiSchema string `json:"openapi_schema"`
}
