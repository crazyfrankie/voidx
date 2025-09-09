package resp

import "github.com/crazyfrankie/voidx/internal/core/llm/entities"

// ProviderResp 提供商响应
type ProviderResp struct {
	Name        string               `json:"name"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Icon        string               `json:"icon"`
	Background  string               `json:"background"`
	ModelTypes  []entities.ModelType `json:"model_types"`
}

// ModelResp 模型响应
type ModelResp struct {
	ModelName       string               `json:"model_name"`
	Label           string               `json:"label"`
	ModelType       string               `json:"model_type"`
	Features        []string             `json:"features"`
	ContextWindow   int                  `json:"context_window"`
	MaxOutputTokens int                  `json:"max_output_tokens"`
	Parameters      []ModelParameterResp `json:"parameters"`
	Metadata        map[string]any       `json:"metadata"`
}

// ModelEntityResp 模型实体响应
type ModelEntityResp struct {
	ModelName       string               `json:"model_name"`
	Label           string               `json:"label"`
	ModelType       string               `json:"model_type"`
	Features        []string             `json:"features"`
	ContextWindow   int                  `json:"context_window"`
	MaxOutputTokens int                  `json:"max_output_tokens"`
	Attributes      map[string]any       `json:"attributes"`
	Parameters      []ModelParameterResp `json:"parameters"`
	Metadata        map[string]any       `json:"metadata"`
}

// ModelParameterResp 模型参数响应
type ModelParameterResp struct {
	Name      string                     `json:"name"`
	Label     string                     `json:"label"`
	Type      string                     `json:"type"`
	Help      string                     `json:"help"`
	Required  bool                       `json:"required"`
	Default   any                        `json:"default"`
	Min       *float64                   `json:"min,omitempty"`
	Max       *float64                   `json:"max,omitempty"`
	Precision int                        `json:"precision"`
	Options   []ModelParameterOptionResp `json:"options"`
}

// ModelParameterOptionResp 模型参数选项响应
type ModelParameterOptionResp struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}
