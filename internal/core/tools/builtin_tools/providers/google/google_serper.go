package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GoogleSerperRequest Google Serper搜索请求参数
type GoogleSerperRequest struct {
	Query string `json:"query" jsonschema:"description=需要检索查询的语句"`
}

// GoogleSerperResponse Google Serper搜索响应
type GoogleSerperResponse struct {
	Success bool   `json:"success"`
	Results string `json:"results,omitempty"`
	Message string `json:"message"`
}

// SerperResponse Serper API响应
type SerperResponse struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
	AnswerBox struct {
		Answer string `json:"answer"`
		Title  string `json:"title"`
		Link   string `json:"link"`
	} `json:"answerBox"`
}

// googleSerperTool Google Serper搜索工具实现
func googleSerperTool(ctx context.Context, req GoogleSerperRequest) (GoogleSerperResponse, error) {
	apiKey := os.Getenv("SERPER_API_KEY")
	if apiKey == "" {
		return GoogleSerperResponse{
			Success: false,
			Message: "SERPER_API_KEY环境变量未配置",
		}, nil
	}

	if req.Query == "" {
		return GoogleSerperResponse{
			Success: false,
			Message: "查询参数不能为空",
		}, nil
	}

	results, err := search(ctx, req.Query, apiKey)
	if err != nil {
		return GoogleSerperResponse{
			Success: false,
			Message: fmt.Sprintf("搜索失败: %v", err),
		}, nil
	}

	return GoogleSerperResponse{
		Success: true,
		Results: results,
		Message: "搜索完成",
	}, nil
}

// search 执行实际的搜索
func search(ctx context.Context, query, apiKey string) (string, error) {
	requestBody := map[string]interface{}{
		"q": query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

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

	var serperResp SerperResponse
	if err := json.Unmarshal(body, &serperResp); err != nil {
		return "", err
	}

	// 格式化结果
	var resultStrings []string

	// 添加答案框（如果有）
	if serperResp.AnswerBox.Answer != "" {
		resultStrings = append(resultStrings, fmt.Sprintf("答案: %s\n来源: %s (%s)",
			serperResp.AnswerBox.Answer, serperResp.AnswerBox.Title, serperResp.AnswerBox.Link))
	}

	// 添加搜索结果
	for i, result := range serperResp.Organic {
		if i >= 5 { // 限制结果数量
			break
		}
		resultStrings = append(resultStrings, fmt.Sprintf("标题: %s\n链接: %s\n摘要: %s",
			result.Title, result.Link, result.Snippet))
	}

	if len(resultStrings) == 0 {
		return fmt.Sprintf("没有找到关于 '%s' 的搜索结果", query), nil
	}

	return strings.Join(resultStrings, "\n\n---\n\n"), nil
}

// NewGoogleSerperTool 创建Google Serper搜索工具
func NewGoogleSerperTool() (tool.InvokableTool, error) {
	return utils.InferTool("google_serper", "这是一个低成本的谷歌搜索API。当你需要搜索时事的时候，可以使用该工具", googleSerperTool)
}
