# LangChainX

LangChainX 是对 LangChainGo 的扩展，提供了一些在官方库中尚未实现的功能。

## FileChatHistory

`FileChatHistory` 是一个基于文件的聊天历史记录实现，类似于 Python 版本 LangChain 中的 `FileChatMessageHistory`。它将聊天消息保存在 JSON 文件中，支持持久化存储和读取。

### 特性

- 将聊天消息保存到 JSON 文件
- 支持多种消息类型（人类、AI、系统、通用）
- 线程安全（使用互斥锁保护文件操作）
- 自动创建目录结构
- 实现了 LangChainGo 的 `schema.ChatMessageHistory` 接口

### 使用方法

```go
import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourusername/llmops/pkg/langchainx"
)

func main() {
    // 创建一个新的 FileChatHistory 实例
    chatHistory, err := langchainx.NewFileChatHistory("/path/to/chat_history.json")
    if err != nil {
        log.Fatalf("创建文件聊天历史记录失败: %v", err)
    }
    
    ctx := context.Background()
    
    // 添加消息
    if err := chatHistory.AddUserMessage(ctx, "你好"); err != nil {
        log.Fatalf("添加用户消息失败: %v", err)
    }
    
    if err := chatHistory.AddAIMessage(ctx, "你好！有什么我可以帮助你的吗？"); err != nil {
        log.Fatalf("添加AI消息失败: %v", err)
    }
    
    // 获取所有消息
    messages, err := chatHistory.Messages(ctx)
    if err != nil {
        log.Fatalf("获取消息失败: %v", err)
    }
    
    // 打印消息
    for i, msg := range messages {
        fmt.Printf("[%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
    }
    
    // 清空历史记录
    if err := chatHistory.Clear(ctx); err != nil {
        log.Fatalf("清空历史记录失败: %v", err)
    }
}
```

### 与 LangChainGo 记忆组件集成

`FileChatHistory` 可以与 LangChainGo 的记忆组件集成使用：

```go
import (
    "github.com/tmc/langchaingo/memory"
    "github.com/yourusername/llmops/pkg/langchainx"
)

// 创建文件聊天历史记录
chatHistory, err := langchainx.NewFileChatHistory("/path/to/chat_history.json")
if err != nil {
    // 处理错误
}

// 使用文件聊天历史记录创建对话缓冲记忆
memoryWithFileHistory := memory.NewConversationBufferMemory(memory.WithChatHistory(chatHistory))

// 现在可以在链或代理中使用这个记忆组件
```

### 文件格式

聊天历史记录保存为 JSON 数组，每个消息包含以下字段：

- `type`: 消息类型（"human", "ai", "system", "generic"）
- `content`: 消息内容
- `name`: 消息名称（仅用于通用消息类型）

示例：

```json
[
  {
    "type": "human",
    "content": "你好"
  },
  {
    "type": "ai",
    "content": "你好！有什么我可以帮助你的吗？"
  },
  {
    "type": "system",
    "content": "这是一条系统消息"
  }
]
```
