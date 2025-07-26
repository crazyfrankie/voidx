package dalle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Dalle3Tool represents a tool for DALLE-3 image generation
type Dalle3Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIKey      string `json:"-"`
}

// Dalle3Request represents the request structure for DALLE-3 API
type Dalle3Request struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	Style          string `json:"style"`
	ResponseFormat string `json:"response_format"`
}

// Dalle3Response represents the response structure from DALLE-3 API
type Dalle3Response struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
}

// NewDalle3Tool creates a new Dalle3Tool instance
func NewDalle3Tool() *Dalle3Tool {
	return &Dalle3Tool{
		Name:        "dalle3",
		Description: "DALLE-3是一个将文本转换成图片的绘图工具",
		APIKey:      os.Getenv("OPENAI_API_KEY"),
	}
}

// Run executes the DALLE-3 image generation
func (t *Dalle3Tool) Run(args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required and must be a string")
	}

	if t.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Get optional parameters with defaults
	size := "1024x1024"
	if s, ok := args["size"].(string); ok {
		size = s
	}

	style := "vivid"
	if s, ok := args["style"].(string); ok {
		style = s
	}

	// Prepare request
	reqBody := Dalle3Request{
		Model:          "dall-e-3",
		Prompt:         query,
		N:              1,
		Size:           size,
		Style:          style,
		ResponseFormat: "url",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var dalleResp Dalle3Response
	if err := json.Unmarshal(body, &dalleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(dalleResp.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	result := map[string]interface{}{
		"image_url":       dalleResp.Data[0].URL,
		"revised_prompt":  dalleResp.Data[0].RevisedPrompt,
		"original_prompt": query,
		"size":           size,
		"style":          style,
	}

	return result, nil
}

// Dalle3 is the exported function for dynamic loading
func Dalle3(args map[string]interface{}) (interface{}, error) {
	tool := NewDalle3Tool()
	return tool.Run(args)
}
