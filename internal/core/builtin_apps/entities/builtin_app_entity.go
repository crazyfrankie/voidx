package entities

// BuiltinAppEntity represents a built-in application configuration
type BuiltinAppEntity struct {
	// ID represents the unique identifier of the built-in app
	ID string `yaml:"id" json:"id"`

	// Category represents the category of the built-in app
	Category string `yaml:"category" json:"category"`

	// Name represents the name of the built-in app
	Name string `yaml:"name" json:"name"`

	// Icon represents the icon URL or identifier
	Icon string `yaml:"icon" json:"icon"`

	// Description provides a detailed description of the app
	Description string `yaml:"description" json:"description"`

	// ModelConfig contains LLM-specific configurations
	ModelConfig map[string]any `yaml:"model_config" json:"model_config"`

	// DialogRound represents the maximum number of conversation turns
	DialogRound int `yaml:"dialog_round" json:"dialog_round"`

	// PresetPrompt contains the initial system prompt
	PresetPrompt string `yaml:"preset_prompt" json:"preset_prompt"`

	// Tools contains the list of available tools and their configurations
	Tools []map[string]any `yaml:"tools" json:"tools"`

	// RetrievalConfig contains configurations for knowledge retrieval
	RetrievalConfig map[string]any `yaml:"retrieval_config" json:"retrieval_config"`

	// LongTermMemory contains configurations for persistent memory
	LongTermMemory map[string]any `yaml:"long_term_memory" json:"long_term_memory"`

	// OpeningStatement is the initial message shown to users
	OpeningStatement string `yaml:"opening_statement" json:"opening_statement"`

	// OpeningQuestions contains suggested initial questions
	OpeningQuestions []string `yaml:"opening_questions" json:"opening_questions"`

	// SpeechToText contains speech recognition configurations
	SpeechToText map[string]any `yaml:"speech_to_text" json:"speech_to_text"`

	// TextToSpeech contains text-to-speech configurations
	TextToSpeech map[string]any `yaml:"text_to_speech" json:"text_to_speech"`

	// SuggestedAfterAnswer contains follow-up suggestion configurations
	SuggestedAfterAnswer map[string]any `yaml:"suggested_after_answer" json:"suggested_after_answer"`

	// ReviewConfig contains content review configurations
	ReviewConfig map[string]any `yaml:"review_config" json:"review_config"`

	// CreatedAt represents the creation timestamp
	Ctime int64 `yaml:"ctime" json:"ctime"`
}

// DefaultAppConfig provides default configuration values
var DefaultAppConfig = map[string]any{
	"model_config": map[string]any{
		// Add default model configuration here
	},
	"dialog_round":     10,
	"preset_prompt":    "",
	"retrieval_config": map[string]any{
		// Add default retrieval configuration here
	},
	"long_term_memory": map[string]any{
		"enable": false,
	},
	"opening_statement": "",
	"opening_questions": []string{},
	"speech_to_text": map[string]any{
		"enable": false,
	},
	"text_to_speech": map[string]any{
		"enable": false,
	},
	"suggested_after_answer": map[string]any{
		"enable": false,
	},
	"review_config": map[string]any{
		"enable": false,
	},
}
