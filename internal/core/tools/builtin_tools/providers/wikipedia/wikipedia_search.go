package wikipedia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WikipediaSearchTool represents a tool for Wikipedia search
type WikipediaSearchTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WikipediaSearchResponse represents the response from Wikipedia search API
type WikipediaSearchResponse struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"search"`
	} `json:"query"`
}

// WikipediaPageResponse represents the response from Wikipedia page API
type WikipediaPageResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

// NewWikipediaSearchTool creates a new WikipediaSearchTool instance
func NewWikipediaSearchTool() *WikipediaSearchTool {
	return &WikipediaSearchTool{
		Name:        "wikipedia_search",
		Description: "一个用于执行维基百科搜索并提取片段和网页的工具",
	}
}

// Run executes the Wikipedia search
func (t *WikipediaSearchTool) Run(args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required and must be a string")
	}

	// Step 1: Search for pages
	searchURL := fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&srlimit=3",
		url.QueryEscape(query))

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search Wikipedia: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %w", err)
	}

	var searchResp WikipediaSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	if len(searchResp.Query.Search) == 0 {
		return fmt.Sprintf("没有找到关于 '%s' 的维基百科条目", query), nil
	}

	// Step 2: Get detailed content for the first result
	firstResult := searchResp.Query.Search[0]
	pageURL := fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&titles=%s&prop=extracts&exintro=true&explaintext=true&format=json&exsectionformat=plain",
		url.QueryEscape(firstResult.Title))

	resp, err = http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get page content: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read page response: %w", err)
	}

	var pageResp WikipediaPageResponse
	if err := json.Unmarshal(body, &pageResp); err != nil {
		return nil, fmt.Errorf("failed to parse page response: %w", err)
	}

	// Format results
	var results []string
	
	// Add main content
	for _, page := range pageResp.Query.Pages {
		if page.Extract != "" {
			// Limit extract length
			extract := page.Extract
			if len(extract) > 1000 {
				extract = extract[:1000] + "..."
			}
			results = append(results, fmt.Sprintf("标题: %s\n内容: %s", page.Title, extract))
		}
		break // Only process the first page
	}

	// Add related search results
	if len(searchResp.Query.Search) > 1 {
		results = append(results, "\n相关条目:")
		for i, result := range searchResp.Query.Search[1:] {
			if i >= 2 { // Limit to 2 additional results
				break
			}
			// Clean HTML tags from snippet
			snippet := strings.ReplaceAll(result.Snippet, "<span class=\"searchmatch\">", "")
			snippet = strings.ReplaceAll(snippet, "</span>", "")
			results = append(results, fmt.Sprintf("  %d. %s - %s", i+1, result.Title, snippet))
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("没有找到关于 '%s' 的详细信息", query), nil
	}

	return strings.Join(results, "\n"), nil
}

// WikipediaSearch is the exported function for dynamic loading
func WikipediaSearch(args map[string]interface{}) (interface{}, error) {
	tool := NewWikipediaSearchTool()
	return tool.Run(args)
}
