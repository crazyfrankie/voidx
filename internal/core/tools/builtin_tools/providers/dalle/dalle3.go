package dalle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// Dalle3Request DALLE-3图像生成请求参数
type Dalle3Request struct {
	Query string `json:"query" jsonschema:"description=输入应该是生成图像的文本提示(prompt)"`
}

// Dalle3Response DALLE-3图像生成响应
type Dalle3Response struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"image_url,omitempty"`
	Message  string `json:"message"`
}

// OpenAIImageRequest OpenAI图像生成请求
type OpenAIImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// OpenAIImageResponse OpenAI图像生成响应
type OpenAIImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// dalle3Tool DALLE-3图像生成工具实现
func dalle3Tool(ctx context.Context, req Dalle3Request) (Dalle3Response, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return Dalle3Response{
			Success: false,
			Message: "OpenAI API密钥未配置，无法使用DALLE-3图像生成功能",
		}, nil
	}

	if req.Query == "" {
		return Dalle3Response{
			Success: false,
			Message: "查询参数不能为空",
		}, nil
	}

	imageURL, err := generateImage(ctx, req.Query, apiKey)
	if err != nil {
		return Dalle3Response{
			Success: false,
			Message: fmt.Sprintf("图像生成失败: %v", err),
		}, nil
	}

	return Dalle3Response{
		Success:  true,
		ImageURL: imageURL,
		Message:  fmt.Sprintf("图像生成成功！\n提示词: %s\n图像URL: %s\n\n请注意：生成的图像URL有效期有限，建议及时保存。", req.Query, imageURL),
	}, nil
}

// generateImage 生成图像
func generateImage(ctx context.Context, prompt, apiKey string) (string, error) {
	requestBody := OpenAIImageRequest{
		Model:  "dall-e-3",
		Prompt: prompt,
		N:      1,
		Size:   "1024x1024",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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

	var imageResp OpenAIImageResponse
	if err := json.Unmarshal(body, &imageResp); err != nil {
		return "", err
	}

	if imageResp.Error.Message != "" {
		return "", fmt.Errorf("OpenAI API error: %s", imageResp.Error.Message)
	}

	if len(imageResp.Data) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	return imageResp.Data[0].URL, nil
}

// NewDalle3Tool 创建DALLE-3图像生成工具
func NewDalle3Tool() (tool.InvokableTool, error) {
	return utils.InferTool("dalle3", "DALLE-3图像生成工具，可以根据文本描述生成图像", dalle3Tool)
}
