package consts

// App相关常量定义

// GenerateIconPromptTemplate 生成icon描述提示词模板
const GenerateIconPromptTemplate = `你是一个拥有10年经验的AI绘画工程师，可以将用户传递的` + "`应用名称`" + `和` + "`应用描述`" + `转换为对应应用的icon描述。
该描述主要用于DallE AI绘画，并且该描述是英文，用户传递的数据如下:

应用名称: {%s}。
应用描述: {%s}。

并且除了icon描述提示词外，其他什么都不要生成`

// AppStatus 应用状态
type AppStatus string

const (
	AppStatusDraft     AppStatus = "draft"
	AppStatusPublished AppStatus = "published"
)

// AppConfigType 应用配置类型
type AppConfigType string

const (
	AppConfigTypeDraft     AppConfigType = "draft"
	AppConfigTypePublished AppConfigType = "published"
)

// DefaultAppConfig 应用默认配置信息
var DefaultAppConfig = map[string]any{
	"model_config": map[string]any{
		"provider": "openai",
		"model":    "gpt-4o-mini",
		"parameters": map[string]any{
			"temperature":       0.5,
			"top_p":             0.85,
			"frequency_penalty": 0.2,
			"presence_penalty":  0.2,
			"max_tokens":        8192,
		},
	},
	"dialog_round":  3,
	"preset_prompt": "",
	"tools":         []any{},
	"workflows":     []any{},
	"datasets":      []any{},
	"retrieval_config": map[string]any{
		"retrieval_strategy": "semantic",
		"k":                  10,
		"score":              0.5,
	},
	"long_term_memory": map[string]any{
		"enable": false,
	},
	"opening_statement": "",
	"opening_questions": []any{},
	"speech_to_text": map[string]any{
		"enable": false,
	},
	"text_to_speech": map[string]any{
		"enable":    false,
		"voice":     "echo",
		"auto_play": false,
	},
	"suggested_after_answer": map[string]any{
		"enable": true,
	},
	"review_config": map[string]any{
		"enable":   false,
		"keywords": []any{},
		"inputs_config": map[string]any{
			"enable":          false,
			"preset_response": "",
		},
		"outputs_config": map[string]any{
			"enable": false,
		},
	},
}
