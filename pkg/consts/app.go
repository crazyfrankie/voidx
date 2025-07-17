package consts

var DefaultAppConfig = map[string]interface{}{
	"model_config": map[string]interface{}{
		"provider": "moonshot",
		"model":    "moonshot-v1-8k",
		"parameters": map[string]interface{}{
			"temperature":       0.5,
			"top_p":             0.85,
			"frequency_penalty": 0.2,
			"presence_penalty":  0.2,
			"max_tokens":        8192,
		},
	},
	"dialog_round":  3,
	"preset_prompt": "",
	"tools":         []interface{}{},
	"workflows":     []interface{}{},
	"datasets":      []interface{}{},
	"retrieval_config": map[string]interface{}{
		"retrieval_strategy": "semantic",
		"k":                  10,
		"score":              0.5,
	},
	"long_term_memory": map[string]interface{}{
		"enable": false,
	},
	"opening_statement":      "",
	"opening_questions":      []interface{}{},
	"speech_to_text":         map[string]interface{}{"enable": false},
	"text_to_speech":         map[string]interface{}{"enable": false, "voice": "echo", "auto_play": false},
	"suggested_after_answer": map[string]interface{}{"enable": true},
	"review_config": map[string]interface{}{
		"enable":         false,
		"keywords":       []interface{}{},
		"inputs_config":  map[string]interface{}{"enable": false, "preset_response": ""},
		"outputs_config": map[string]interface{}{"enable": false},
	},
}
