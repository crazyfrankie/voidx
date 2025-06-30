package langchainx

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestFileChatHistory(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file_chat_history_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件路径
	filePath := filepath.Join(tempDir, "chat_history.txt")

	// 创建 FileChatHistory 实例
	history, err := NewFileChatHistory(filePath)
	require.NoError(t, err)

	// 测试添加消息
	ctx := context.Background()
	err = history.AddUserMessage(ctx, "你好")
	require.NoError(t, err)
	err = history.AddAIMessage(ctx, "你好！有什么我可以帮助你的吗？")
	require.NoError(t, err)

	// 测试获取消息
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, "你好", messages[0].GetContent())
	assert.Equal(t, "你好！有什么我可以帮助你的吗？", messages[1].GetContent())

	// 测试添加自定义消息
	err = history.AddMessage(ctx, llms.SystemChatMessage{Content: "这是一条系统消息"})
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(messages))
	assert.Equal(t, "这是一条系统消息", messages[2].GetContent())

	// 测试设置消息
	newMessages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "新消息1"},
		llms.AIChatMessage{Content: "新消息2"},
	}
	err = history.SetMessages(ctx, newMessages)
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, "新消息1", messages[0].GetContent())
	assert.Equal(t, "新消息2", messages[1].GetContent())

	// 测试清空消息
	err = history.Clear(ctx)
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(messages))

	// 测试文件持久化
	err = history.AddUserMessage(ctx, "测试持久化")
	require.NoError(t, err)

	// 创建新的实例，应该能读取到之前保存的消息
	newHistory, err := NewFileChatHistory(filePath)
	require.NoError(t, err)
	messages, err = newHistory.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "测试持久化", messages[0].GetContent())
}

func TestFileChatHistoryWithNestedDirectory(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file_chat_history_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建嵌套目录路径
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	filePath := filepath.Join(nestedDir, "chat_history.json")

	// 创建 FileChatHistory 实例，应该自动创建目录
	history, err := NewFileChatHistory(filePath)
	require.NoError(t, err)

	// 验证目录是否被创建
	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)

	// 测试添加消息
	ctx := context.Background()
	err = history.AddUserMessage(ctx, "测试嵌套目录")
	require.NoError(t, err)

	// 验证消息是否被保存
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "测试嵌套目录", messages[0].GetContent())
}
