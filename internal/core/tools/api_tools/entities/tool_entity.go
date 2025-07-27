package entities

import "github.com/crazyfrankie/voidx/internal/models/entity"

// ToolEntity API工具实体信息，记录了创建工具所需的配置信息
type ToolEntity struct {
	ID          string           `json:"id"`          // API工具提供者对应的id
	Name        string           `json:"name"`        // API工具的名称
	URL         string           `json:"url"`         // API工具发起请求的URL地址
	Method      string           `json:"method"`      // API工具发起请求的方法
	Description string           `json:"description"` // API工具的描述信息
	Headers     []entity.Header  `json:"headers"`     // API工具的请求头信息
	Parameters  []map[string]any `json:"parameters"`  // API工具的参数列表信息
}

// NewToolEntity 创建新的工具实体
func NewToolEntity(id, name, url, method, description string) *ToolEntity {
	return &ToolEntity{
		ID:          id,
		Name:        name,
		URL:         url,
		Method:      method,
		Description: description,
		Headers:     make([]entity.Header, 0),
		Parameters:  make([]map[string]any, 0),
	}
}

// AddHeader 添加请求头
func (t *ToolEntity) AddHeader(key, value string) {
	t.Headers = append(t.Headers, entity.Header{
		Key:   key,
		Value: value,
	})
}

// AddParameter 添加参数
func (t *ToolEntity) AddParameter(param map[string]any) {
	t.Parameters = append(t.Parameters, param)
}

// SetHeaders 设置请求头列表
func (t *ToolEntity) SetHeaders(headers []entity.Header) {
	t.Headers = headers
}

// SetParameters 设置参数列表
func (t *ToolEntity) SetParameters(parameters []map[string]any) {
	t.Parameters = parameters
}
