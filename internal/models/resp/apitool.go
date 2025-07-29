package resp

import (
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// ApiToolProviderResp API工具提供商响应
type ApiToolProviderResp struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Icon          string          `json:"icon"`
	Description   string          `json:"description"`
	OpenAPISchema string          `json:"openapi_schema"`
	Headers       []entity.Header `json:"headers"`
	Ctime         int64           `json:"ctime"`
}

// ApiToolResp API工具响应
type ApiToolResp struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Inputs      []map[string]any `json:"inputs"`
	Provider    ApiToolProvider  `json:"provider"`
}

type ApiToolProvider struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Icon        string          `json:"icon"`
	Description string          `json:"description"`
	Headers     []entity.Header `json:"headers"`
}

type GetApiToolProvidersWithPageResp struct {
	List      []GetApiToolProvidersWithPage `json:"list"`
	Paginator Paginator                     `json:"paginator"`
}

type GetApiToolProvidersWithPage struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Icon        string          `json:"icon"`
	Description string          `json:"description"`
	Headers     []entity.Header `json:"headers"`
	Tools       []ApiTool       `json:"tools"`
	Ctime       int64           `json:"ctime"`
}

type ApiTool struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Inputs      map[string]any `json:"inputs"`
}
