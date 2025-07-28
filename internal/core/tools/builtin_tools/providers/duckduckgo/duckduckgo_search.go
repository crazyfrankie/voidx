package duckduckgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bytedance/sonic"
)

// DuckDuckGoSearchTool represents a tool for DuckDuckGo search
type DuckDuckGoSearchTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DDGResult represents a search result from DuckDuckGo
type DDGResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// NewDuckDuckGoSearchTool creates a new DuckDuckGoSearchTool instance
func NewDuckDuckGoSearchTool() *DuckDuckGoSearchTool {
	return &DuckDuckGoSearchTool{
		Name:        "duckduckgo_search",
		Description: "一个注重隐私的搜索工具，当你需要搜索时事时可以使用该工具，工具的输入是一个查询语句",
	}
}

// Run executes the DuckDuckGo search
func (t *DuckDuckGoSearchTool) Run(ctx context.Context, query string) (string, error) {
	// Use DuckDuckGo instant answer API
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to make search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the JSON response
	var result map[string]any
	if err := sonic.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var results []string

	// Check for instant answer
	if abstract, ok := result["Abstract"].(string); ok && abstract != "" {
		results = append(results, fmt.Sprintf("摘要: %s", abstract))
		if abstractURL, ok := result["AbstractURL"].(string); ok && abstractURL != "" {
			results = append(results, fmt.Sprintf("来源: %s", abstractURL))
		}
	}

	// Check for definition
	if definition, ok := result["Definition"].(string); ok && definition != "" {
		results = append(results, fmt.Sprintf("定义: %s", definition))
		if definitionURL, ok := result["DefinitionURL"].(string); ok && definitionURL != "" {
			results = append(results, fmt.Sprintf("来源: %s", definitionURL))
		}
	}

	// Check for related topics
	if relatedTopics, ok := result["RelatedTopics"].([]any); ok && len(relatedTopics) > 0 {
		results = append(results, "相关主题:")
		for i, topic := range relatedTopics {
			if i >= 3 { // Limit to 3 related topics
				break
			}
			if topicMap, ok := topic.(map[string]any); ok {
				if text, ok := topicMap["Text"].(string); ok && text != "" {
					if firstURL, ok := topicMap["FirstURL"].(string); ok && firstURL != "" {
						results = append(results, fmt.Sprintf("  %d. %s - %s", i+1, text, firstURL))
					} else {
						results = append(results, fmt.Sprintf("  %d. %s", i+1, text))
					}
				}
			}
		}
	}

	// If no results found, try a simple web search approach
	if len(results) == 0 {
		// Fallback to a simple HTML scraping approach (simplified)
		results = append(results, fmt.Sprintf("搜索查询: %s", query))
		results = append(results, "注意: DuckDuckGo API 没有返回详细结果。建议使用更具体的搜索词。")
	}

	return strings.Join(results, "\n"), nil
}

// DuckduckgoSearch is the exported function for dynamic loading
func DuckduckgoSearch(ctx context.Context, input string) (string, error) {
	tool := NewDuckDuckGoSearchTool()
	return tool.Run(ctx, input)
}
