package duckduckgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// DuckDuckGoSearchRequest DuckDuckGo搜索请求参数
type DuckDuckGoSearchRequest struct {
	Query string `json:"query" jsonschema:"description=需要搜索的查询语句"`
}

// DuckDuckGoSearchResponse DuckDuckGo搜索响应
type DuckDuckGoSearchResponse struct {
	Success bool   `json:"success"`
	Results string `json:"results"`
	Message string `json:"message"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// duckduckgoSearchTool DuckDuckGo搜索工具实现
func duckduckgoSearchTool(ctx context.Context, req DuckDuckGoSearchRequest) (DuckDuckGoSearchResponse, error) {
	if req.Query == "" {
		return DuckDuckGoSearchResponse{
			Success: false,
			Message: "查询参数不能为空",
		}, nil
	}

	results, err := search(ctx, req.Query)
	if err != nil {
		return DuckDuckGoSearchResponse{
			Success: false,
			Message: fmt.Sprintf("搜索失败: %v", err),
		}, nil
	}

	// 格式化结果
	var resultStrings []string
	for i, result := range results {
		if i >= 5 { // 限制结果数量
			break
		}
		resultStrings = append(resultStrings, fmt.Sprintf("标题: %s\n链接: %s\n摘要: %s",
			result.Title, result.Link, result.Snippet))
	}

	formattedResults := strings.Join(resultStrings, "\n\n---\n\n")

	return DuckDuckGoSearchResponse{
		Success: true,
		Results: formattedResults,
		Message: fmt.Sprintf("搜索完成，找到 %d 个结果", len(results)),
	}, nil
}

// search 执行实际的搜索
func search(ctx context.Context, query string) ([]SearchResult, error) {
	// 使用DuckDuckGo的即时答案API
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LLMOps/1.0)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析DuckDuckGo API响应
	var apiResponse struct {
		Abstract      string `json:"Abstract"`
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	var results []SearchResult

	// 添加摘要结果
	if apiResponse.AbstractText != "" {
		results = append(results, SearchResult{
			Title:   "摘要",
			Link:    apiResponse.AbstractURL,
			Snippet: apiResponse.AbstractText,
		})
	}

	// 添加相关主题
	for _, topic := range apiResponse.RelatedTopics {
		if topic.Text != "" && topic.FirstURL != "" {
			// 提取标题（通常在第一个破折号之前）
			title := topic.Text
			if idx := strings.Index(title, " - "); idx > 0 {
				title = title[:idx]
			}

			results = append(results, SearchResult{
				Title:   title,
				Link:    topic.FirstURL,
				Snippet: topic.Text,
			})
		}
	}

	// 如果没有结果，返回一个基本的搜索建议
	if len(results) == 0 {
		results = append(results, SearchResult{
			Title:   "搜索建议",
			Link:    fmt.Sprintf("https://duckduckgo.com/?q=%s", url.QueryEscape(query)),
			Snippet: fmt.Sprintf("没有找到关于 '%s' 的即时答案，建议访问DuckDuckGo进行更详细的搜索。", query),
		})
	}

	return results, nil
}

// NewDuckDuckGoSearchTool 创建DuckDuckGo搜索工具
func NewDuckDuckGoSearchTool() (tool.InvokableTool, error) {
	return utils.InferTool("duckduckgo_search", "一个注重隐私的搜索工具，当你需要搜索时事时可以使用该工具", duckduckgoSearchTool)
}
