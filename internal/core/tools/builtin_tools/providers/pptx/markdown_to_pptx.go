package pptx

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// MarkdownToPPTXRequest Markdown转PPT请求参数
type MarkdownToPPTXRequest struct {
	Markdown string `json:"markdown" jsonschema:"description=要生成PPT内容的markdown文档字符串"`
}

// MarkdownToPPTXResponse Markdown转PPT响应
type MarkdownToPPTXResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Message string `json:"message"`
}

// markdownToPPTXTool Markdown转PPT工具实现
func markdownToPPTXTool(ctx context.Context, req MarkdownToPPTXRequest) (MarkdownToPPTXResponse, error) {
	if req.Markdown == "" {
		return MarkdownToPPTXResponse{
			Success: false,
			Message: "Markdown内容不能为空",
		}, nil
	}

	pptContent, err := convertMarkdownToPPT(req.Markdown)
	if err != nil {
		return MarkdownToPPTXResponse{
			Success: false,
			Message: fmt.Sprintf("PPT生成失败: %v", err),
		}, nil
	}

	return MarkdownToPPTXResponse{
		Success: true,
		Content: pptContent,
		Message: "PPT结构化内容生成成功",
	}, nil
}

// convertMarkdownToPPT 将Markdown转换为PPT结构化内容
func convertMarkdownToPPT(markdown string) (string, error) {
	lines := strings.Split(markdown, "\n")
	var pptSlides []string
	var currentSlide strings.Builder
	var slideTitle string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// 处理标题
		if strings.HasPrefix(line, "# ") {
			// H1作为PPT标题页
			if currentSlide.Len() > 0 {
				pptSlides = append(pptSlides, formatSlide(slideTitle, currentSlide.String()))
				currentSlide.Reset()
			}
			title := strings.TrimPrefix(line, "# ")
			pptSlides = append(pptSlides, formatTitleSlide(title))
		} else if strings.HasPrefix(line, "## ") {
			// H2作为新幻灯片标题
			if currentSlide.Len() > 0 {
				pptSlides = append(pptSlides, formatSlide(slideTitle, currentSlide.String()))
				currentSlide.Reset()
			}
			slideTitle = strings.TrimPrefix(line, "## ")
		} else if strings.HasPrefix(line, "### ") {
			// H3作为内容小标题
			if currentSlide.Len() > 0 {
				currentSlide.WriteString("\n")
			}
			currentSlide.WriteString("【" + strings.TrimPrefix(line, "### ") + "】\n")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// 列表项
			if currentSlide.Len() > 0 {
				currentSlide.WriteString("\n")
			}
			content := strings.TrimPrefix(line, "- ")
			content = strings.TrimPrefix(content, "* ")
			currentSlide.WriteString("• " + content)
		} else if strings.HasPrefix(line, "```") {
			// 代码块（简化处理）
			if currentSlide.Len() > 0 {
				currentSlide.WriteString("\n")
			}
			currentSlide.WriteString("[代码块]")
		} else if strings.HasPrefix(line, "![") {
			// 图片（简化处理）
			if currentSlide.Len() > 0 {
				currentSlide.WriteString("\n")
			}
			currentSlide.WriteString("[图片]")
		} else if line != "" {
			// 普通段落
			if currentSlide.Len() > 0 {
				currentSlide.WriteString("\n")
			}
			currentSlide.WriteString(line)
		}
	}

	// 添加最后一张幻灯片
	if currentSlide.Len() > 0 {
		pptSlides = append(pptSlides, formatSlide(slideTitle, currentSlide.String()))
	}

	if len(pptSlides) == 0 {
		return "无法从提供的Markdown内容生成PPT结构", nil
	}

	result := "PPT生成成功！以下是PPT的结构化内容：\n\n"
	result += strings.Join(pptSlides, "\n"+strings.Repeat("=", 50)+"\n\n")
	result += "\n注意：这是PPT的结构化文本表示。在实际应用中，您可以将此内容导入到PowerPoint或其他演示软件中创建正式的PPT文件。"

	return result, nil
}

// formatTitleSlide 格式化标题页
func formatTitleSlide(title string) string {
	return fmt.Sprintf("【封面页】\n标题: %s\n副标题: 由LLMOps平台生成", title)
}

// formatSlide 格式化内容页
func formatSlide(title, content string) string {
	if title == "" {
		title = "内容页"
	}
	return fmt.Sprintf("【幻灯片】\n标题: %s\n内容:\n%s", title, content)
}

// NewMarkdownToPPTXTool 创建Markdown转PPT工具
func NewMarkdownToPPTXTool() (tool.InvokableTool, error) {
	return utils.InferTool("markdown_to_pptx", "这是一个可以将markdown文本转换成PPT的工具，传递的参数是markdown对应的文本字符串，返回的数据是PPT的结构化内容", markdownToPPTXTool)
}
