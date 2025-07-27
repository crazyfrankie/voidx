package dalle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bytedance/sonic"
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
func (t *Dalle3Tool) Run(ctx context.Context, input string) (string, error) {
	args := make(map[string]any)
	if err := sonic.UnmarshalString(input, &args); err != nil {
		return "", err
	}

	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required and must be a string")
	}

	if t.APIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set")
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

	jsonData, err := sonic.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var dalleResp Dalle3Response
	if err := sonic.Unmarshal(body, &dalleResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(dalleResp.Data) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	result := map[string]any{
		"image_url":       dalleResp.Data[0].URL,
		"revised_prompt":  dalleResp.Data[0].RevisedPrompt,
		"original_prompt": query,
		"size":            size,
		"style":           style,
	}

	res, err := sonic.MarshalString(result)
	if err != nil {
		return "", err
	}

	return res, nil
}

// Dalle3 is the exported function for dynamic loading
func Dalle3(ctx context.Context, input string) (string, error) {
	tool := NewDalle3Tool()
	return tool.Run(ctx, input)
}
