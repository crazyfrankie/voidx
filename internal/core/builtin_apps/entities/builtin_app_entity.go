package entities

import (
	"time"
)

// BuiltinAppEntity represents a built-in application configuration
type BuiltinAppEntity struct {
	// ID represents the unique identifier of the built-in app
	ID string `json:"id"`

	// Category represents the category of the built-in app
	Category string `json:"category"`

	// Name represents the name of the built-in app
	Name string `json:"name"`

	// Icon represents the icon URL or identifier
	Icon string `json:"icon"`

	// Description provides a detailed description of the app
	Description string `json:"description"`

	// LanguageModelConfig contains LLM-specific configurations
	LanguageModelConfig map[string]any `json:"language_model_config"`

	// DialogRound represents the maximum number of conversation turns
	DialogRound int `json:"dialog_round"`

	// PresetPrompt contains the initial system prompt
	PresetPrompt string `json:"preset_prompt"`

	// Tools contains the list of available tools and their configurations
	Tools []map[string]any `json:"tools"`

	// RetrievalConfig contains configurations for knowledge retrieval
	RetrievalConfig map[string]any `json:"retrieval_config"`

	// LongTermMemory contains configurations for persistent memory
	LongTermMemory map[string]any `json:"long_term_memory"`

	// OpeningStatement is the initial message shown to users
	OpeningStatement string `json:"opening_statement"`

	// OpeningQuestions contains suggested initial questions
	OpeningQuestions []string `json:"opening_questions"`

	// SpeechToText contains speech recognition configurations
	SpeechToText map[string]any `json:"speech_to_text"`

	// TextToSpeech contains text-to-speech configurations
	TextToSpeech map[string]any `json:"text_to_speech"`

	// SuggestedAfterAnswer contains follow-up suggestion configurations
	SuggestedAfterAnswer map[string]any `json:"suggested_after_answer"`

	// ReviewConfig contains content review configurations
	ReviewConfig map[string]any `json:"review_config"`

	// CreatedAt represents the creation timestamp
	CreatedAt time.Time `json:"created_at"`
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
