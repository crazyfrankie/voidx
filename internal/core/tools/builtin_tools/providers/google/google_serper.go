package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// GoogleSerperTool represents a tool for Google Serper search
type GoogleSerperTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIKey      string `json:"-"`
}

// SerperRequest represents the request structure for Serper API
type SerperRequest struct {
	Q  string `json:"q"`
	GL string `json:"gl,omitempty"`
	HL string `json:"hl,omitempty"`
}

// SerperResponse represents the response structure from Serper API
type SerperResponse struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
	AnswerBox struct {
		Answer string `json:"answer"`
	} `json:"answerBox"`
}

// NewGoogleSerperTool creates a new GoogleSerperTool instance
func NewGoogleSerperTool() *GoogleSerperTool {
	return &GoogleSerperTool{
		Name:        "google_serper",
		Description: "这是一个低成本的谷歌搜索API。当你需要搜索时事的时候，可以使用该工具，该工具的输入是一个查询语句",
		APIKey:      os.Getenv("SERPER_API_KEY"),
	}
}

// Run executes the Google Serper search
func (t *GoogleSerperTool) Run(ctx context.Context, query string) (string, error) {
	if t.APIKey == "" {
		return "", fmt.Errorf("SERPER_API_KEY environment variable is not set")
	}

	// Prepare request
	reqBody := SerperRequest{
		Q:  query,
		GL: "cn",
		HL: "zh-cn",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-KEY", t.APIKey)
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
	var serperResp SerperResponse
	if err := json.Unmarshal(body, &serperResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Format results
	var results string

	// Add answer box if available
	if serperResp.AnswerBox.Answer != "" {
		results += fmt.Sprintf("答案: %s", serperResp.AnswerBox.Answer)
	}

	// Add organic results
	for i, result := range serperResp.Organic {
		if i >= 5 { // Limit to top 5 results
			break
		}
		results += fmt.Sprintf("%d. %s\n   链接: %s\n   摘要: %s",
			i+1, result.Title, result.Link, result.Snippet)
	}

	return results, nil
}

// GoogleSerper is the exported function for dynamic loading
func GoogleSerper(ctx context.Context, input string) (string, error) {
	tool := NewGoogleSerperTool()
	return tool.Run(ctx, input)
}
