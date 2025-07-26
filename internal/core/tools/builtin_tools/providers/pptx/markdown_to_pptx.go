package pptx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MarkdownToPptxTool represents a tool for converting Markdown to PPTX
type MarkdownToPptxTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Slide represents a single slide in the presentation
type Slide struct {
	Title   string
	Content []string
}

// NewMarkdownToPptxTool creates a new MarkdownToPptxTool instance
func NewMarkdownToPptxTool() *MarkdownToPptxTool {
	return &MarkdownToPptxTool{
		Name:        "markdown_to_pptx",
		Description: "这是一款能将Markdown文档内容转换成本地PPT文件的工具",
	}
}

// Run executes the Markdown to PPTX conversion
func (t *MarkdownToPptxTool) Run(args map[string]interface{}) (interface{}, error) {
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required and must be a string")
	}

	// Parse markdown content into slides
	slides := t.parseMarkdownToSlides(content)
	if len(slides) == 0 {
		return "没有找到可转换的内容", nil
	}

	// Generate output filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("presentation_%s.txt", timestamp)
	
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename)

	// Generate presentation content (simplified text format)
	presentationContent := t.generatePresentationContent(slides)

	// Write to file
	if err := os.WriteFile(outputPath, []byte(presentationContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write presentation file: %w", err)
	}

	result := map[string]interface{}{
		"success":     true,
		"output_path": outputPath,
		"slides_count": len(slides),
		"message":     fmt.Sprintf("成功生成演示文稿，包含 %d 张幻灯片", len(slides)),
	}

	return result, nil
}

// parseMarkdownToSlides parses markdown content into slides
func (t *MarkdownToPptxTool) parseMarkdownToSlides(content string) []Slide {
	var slides []Slide
	lines := strings.Split(content, "\n")
	
	var currentSlide *Slide
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for slide title (# or ##)
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			// Save previous slide if exists
			if currentSlide != nil {
				slides = append(slides, *currentSlide)
			}
			
			// Start new slide
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimPrefix(title, "## ")
			currentSlide = &Slide{
				Title:   title,
				Content: []string{},
			}
		} else if currentSlide != nil && line != "" {
			// Add content to current slide
			// Clean up markdown formatting
			cleanLine := t.cleanMarkdownFormatting(line)
			if cleanLine != "" {
				currentSlide.Content = append(currentSlide.Content, cleanLine)
			}
		}
	}
	
	// Add the last slide
	if currentSlide != nil {
		slides = append(slides, *currentSlide)
	}
	
	return slides
}

// cleanMarkdownFormatting removes basic markdown formatting
func (t *MarkdownToPptxTool) cleanMarkdownFormatting(text string) string {
	// Remove bold and italic formatting
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)
	text = re.ReplaceAllString(text, "$1")
	
	re = regexp.MustCompile(`\*(.*?)\*`)
	text = re.ReplaceAllString(text, "$1")
	
	// Remove code formatting
	re = regexp.MustCompile("`(.*?)`")
	text = re.ReplaceAllString(text, "$1")
	
	// Remove list markers
	text = strings.TrimPrefix(text, "- ")
	text = strings.TrimPrefix(text, "* ")
	
	// Remove numbered list markers
	re = regexp.MustCompile(`^\d+\.\s`)
	text = re.ReplaceAllString(text, "")
	
	return strings.TrimSpace(text)
}

// generatePresentationContent generates the presentation content in text format
func (t *MarkdownToPptxTool) generatePresentationContent(slides []Slide) string {
	var content strings.Builder
	
	content.WriteString("=== 演示文稿 ===\n")
	content.WriteString(fmt.Sprintf("生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("幻灯片数量: %d\n\n", len(slides)))
	
	for i, slide := range slides {
		content.WriteString(fmt.Sprintf("=== 幻灯片 %d ===\n", i+1))
		content.WriteString(fmt.Sprintf("标题: %s\n\n", slide.Title))
		
		if len(slide.Content) > 0 {
			content.WriteString("内容:\n")
			for _, item := range slide.Content {
				content.WriteString(fmt.Sprintf("• %s\n", item))
			}
		}
		
		content.WriteString("\n" + strings.Repeat("-", 50) + "\n\n")
	}
	
	content.WriteString("注意: 这是一个简化的文本格式输出。\n")
	content.WriteString("在实际应用中，可以使用专门的PPTX库（如unioffice）来生成真正的PowerPoint文件。\n")
	
	return content.String()
}

// MarkdownToPptx is the exported function for dynamic loading
func MarkdownToPptx(args map[string]interface{}) (interface{}, error) {
	tool := NewMarkdownToPptxTool()
	return tool.Run(args)
}
