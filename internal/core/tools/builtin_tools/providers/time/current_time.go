package time

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CurrentTimeRequest 获取当前时间请求参数（空结构体，因为不需要参数）
type CurrentTimeRequest struct{}

// CurrentTimeResponse 获取当前时间响应
type CurrentTimeResponse struct {
	Success     bool   `json:"success"`
	CurrentTime string `json:"current_time,omitempty"`
	Message     string `json:"message"`
}

// currentTimeTool 获取当前时间工具实现
func currentTimeTool(ctx context.Context, req CurrentTimeRequest) (CurrentTimeResponse, error) {
	currentTime := time.Now().Format("2006-01-02 15:04:05 MST")

	return CurrentTimeResponse{
		Success:     true,
		CurrentTime: currentTime,
		Message:     fmt.Sprintf("当前时间: %s", currentTime),
	}, nil
}

// NewCurrentTimeTool 创建获取当前时间的工具
func NewCurrentTimeTool() (tool.InvokableTool, error) {
	return utils.InferTool("current_time", "一个用于获取当前时间的工具", currentTimeTool)
}
