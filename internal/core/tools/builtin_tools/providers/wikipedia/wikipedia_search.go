package wikipedia

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

// WikipediaSearchRequest 维基百科搜索请求参数
type WikipediaSearchRequest struct {
	Query string `json:"query" jsonschema:"description=需要搜索的查询语句"`
}

// WikipediaSearchResponse 维基百科搜索响应
type WikipediaSearchResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Message string `json:"message"`
}

// WikipediaSearchAPIResponse 维基百科搜索API响应
type WikipediaSearchAPIResponse struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
			Size    int    `json:"size"`
		} `json:"search"`
	} `json:"query"`
}

// WikipediaPageResponse 维基百科页面响应
type WikipediaPageResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

// wikipediaSearchTool 维基百科搜索工具实现
func wikipediaSearchTool(ctx context.Context, req WikipediaSearchRequest) (WikipediaSearchResponse, error) {
	if req.Query == "" {
		return WikipediaSearchResponse{
			Success: false,
			Message: "查询参数不能为空",
		}, nil
	}

	results, err := search(ctx, req.Query)
	if err != nil {
		return WikipediaSearchResponse{
			Success: false,
			Message: fmt.Sprintf("搜索失败: %v", err),
		}, nil
	}

	return WikipediaSearchResponse{
		Success: true,
		Content: results,
		Message: "维基百科搜索完成",
	}, nil
}

// search 执行实际的搜索
func search(ctx context.Context, query string) (string, error) {
	// 首先搜索相关页面
	searchURL := fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&srlimit=3",
		url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "LLMOps/1.0 (https://example.com/contact)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var searchResp WikipediaSearchAPIResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return "", err
	}

	if len(searchResp.Query.Search) == 0 {
		return fmt.Sprintf("没有找到关于 '%s' 的维基百科条目", query), nil
	}

	// 获取第一个搜索结果的详细内容
	firstResult := searchResp.Query.Search[0]
	pageContent, err := getPageContent(ctx, firstResult.Title)
	if err != nil {
		// 如果获取页面内容失败，返回搜索结果摘要
		var resultStrings []string
		for _, result := range searchResp.Query.Search {
			snippet := strings.ReplaceAll(result.Snippet, "<span class=\"searchmatch\">", "")
			snippet = strings.ReplaceAll(snippet, "</span>", "")
			resultStrings = append(resultStrings, fmt.Sprintf("标题: %s\n摘要: %s",
				result.Title, snippet))
		}
		return strings.Join(resultStrings, "\n\n---\n\n"), nil
	}

	return pageContent, nil
}

// getPageContent 获取页面详细内容
func getPageContent(ctx context.Context, title string) (string, error) {
	pageURL := fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&prop=extracts&exintro&explaintext&titles=%s&format=json",
		url.QueryEscape(title))

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "LLMOps/1.0 (https://example.com/contact)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var pageResp WikipediaPageResponse
	if err := json.Unmarshal(body, &pageResp); err != nil {
		return "", err
	}

	for _, page := range pageResp.Query.Pages {
		if page.Extract != "" {
			// 限制内容长度
			extract := page.Extract
			if len(extract) > 1000 {
				extract = extract[:1000] + "..."
			}
			return fmt.Sprintf("标题: %s\n\n内容: %s\n\n来源: https://zh.wikipedia.org/wiki/%s",
				page.Title, extract, url.QueryEscape(strings.ReplaceAll(page.Title, " ", "_"))), nil
		}
	}

	return fmt.Sprintf("找到页面 '%s' 但无法获取内容", title), nil
}

// NewWikipediaSearchTool 创建维基百科搜索工具
func NewWikipediaSearchTool() (tool.InvokableTool, error) {
	return utils.InferTool("wikipedia_search", "维基百科搜索工具，可以搜索维基百科上的文章内容", wikipediaSearchTool)
}
