package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

const pattern = "(?m)```json\\s*\\n([\\s\\S]+?)```"

// ReACTAgent 基于ReACT推理的智能体，继承FunctionCallAgent，并重写long_term_memory_node和llm_node两个节点
type ReACTAgent struct {
	*FunctionCallAgent
}

// NewReACTAgent creates a new ReACT agent
func NewReACTAgent(llm llms.Model, config entities.AgentConfig, agentQueueManager *AgentQueueManager) BaseAgent {
	functionCallAgent := NewFunctionCallAgent(llm, config, agentQueueManager).(*FunctionCallAgent)
	return &ReACTAgent{FunctionCallAgent: functionCallAgent}
}

// Stream implements the BaseAgent interface for ReACT Agent
func (r *ReACTAgent) Stream(ctx context.Context, input entities.AgentState) (<-chan entities.AgentThought, error) {
	// Ensure task ID is set
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}

	// Initialize other fields if not set
	if input.History == nil {
		input.History = make([]llms.ChatMessage, 0)
	}

	// Create queue for this task
	thoughtChan, err := r.queueManager.Listen(input.TaskID, r.agentConfig.UserID, r.agentConfig.InvokeFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// Start processing in background
	go func() {
		defer func() {
			r.queueManager.Publish(input.TaskID, entities.AgentThought{
				ID:     uuid.New(),
				TaskID: input.TaskID,
				Event:  entities.EventAgentEnd,
			})
		}()

		// Process through ReACT agent graph
		if err := r.processReACTAgentGraph(ctx, input); err != nil {
			r.queueManager.PublishError(input.TaskID, fmt.Sprintf("ReACT Agent processing failed: %v", err))
		}
	}()

	return thoughtChan, nil
}

// processReACTAgentGraph processes the input through the ReACT agent's execution graph
func (r *ReACTAgent) processReACTAgentGraph(ctx context.Context, state entities.AgentState) error {
	// 1. Preset operation node
	if shouldEnd, err := r.presetOperationNode(state); err != nil {
		return err
	} else if shouldEnd {
		return nil
	}

	// 2. Long term memory recall node (ReACT specific implementation)
	if err := r.longTermMemoryRecallNodeReACT(&state); err != nil {
		return err
	}

	// 3. Main processing loop
	for state.IterationCount <= r.agentConfig.MaxIterationCount {
		// LLM node (ReACT specific implementation)
		shouldEnd, err := r.llmNodeReACT(ctx, &state)
		if err != nil {
			return err
		}
		if shouldEnd {
			return nil
		}

		// Tools node (ReACT specific implementation)
		if err := r.toolsNodeReACT(&state); err != nil {
			return err
		}

		state.IterationCount++
	}

	// Max iteration reached
	r.queueManager.Publish(state.TaskID, entities.AgentThought{
		ID:      uuid.New(),
		TaskID:  state.TaskID,
		Event:   entities.EventAgentMessage,
		Thought: entities.MaxIterationResponse,
		Answer:  entities.MaxIterationResponse,
		Latency: 0,
	})

	return nil
}

// longTermMemoryRecallNodeReACT 重写长期记忆召回节点，使用prompt实现工具调用及规范数据生成
func (r *ReACTAgent) longTermMemoryRecallNodeReACT(state *entities.AgentState) error {
	// 1.判断是否支持工具调用，如果支持工具调用，则可以直接使用工具智能体的长期记忆召回节点
	if r.supportsToolCall() {
		return r.longTermMemoryRecallNode(*state)
	}

	// 2.根据传递的智能体配置判断是否需要召回长期记忆
	longTermMemory := ""
	if r.agentConfig.EnableLongTermMemory {
		longTermMemory = state.LongTermMemory
		r.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:          uuid.New(),
			TaskID:      state.TaskID,
			Event:       entities.EventLongTermMemoryRecall,
			Observation: longTermMemory,
		})
	}

	// 3.检测是否支持AGENT_THOUGHT，如果不支持，则使用没有工具描述的prompt
	var systemPrompt string
	if !r.supportsAgentThought() {
		systemPrompt = strings.ReplaceAll(entities.AgentSystemPromptTemplate, "{preset_prompt}", r.agentConfig.PresetPrompt)
		systemPrompt = strings.ReplaceAll(systemPrompt, "{long_term_memory}", longTermMemory)
	} else {
		// 4.支持智能体推理，则使用REACT_AGENT_SYSTEM_PROMPT_TEMPLATE并添加工具描述
		toolDescription := r.renderToolDescription()
		systemPrompt = strings.ReplaceAll(entities.ReactAgentSystemPromptTemplate, "{preset_prompt}", r.agentConfig.PresetPrompt)
		systemPrompt = strings.ReplaceAll(systemPrompt, "{long_term_memory}", longTermMemory)
		systemPrompt = strings.ReplaceAll(systemPrompt, "{tool_description}", toolDescription)
	}

	// 5.将短期历史消息添加到消息列表中
	var presetMessages []llms.ChatMessage
	presetMessages = append(presetMessages, llms.SystemChatMessage{Content: systemPrompt})

	// 6.校验历史消息是不是复数形式，也就是[人类消息, AI消息, 人类消息, AI消息, ...]
	if len(state.History)%2 != 0 {
		r.queueManager.PublishError(state.TaskID, "智能体历史消息列表格式错误")
		return fmt.Errorf("智能体历史消息列表格式错误, len(history)=%d", len(state.History))
	}

	// 7.拼接历史消息
	presetMessages = append(presetMessages, state.History...)

	// 8.拼接当前用户的提问消息
	if len(state.Messages) > 0 {
		humanMessage := state.Messages[len(state.Messages)-1]
		presetMessages = append(presetMessages, llms.HumanChatMessage{Content: humanMessage.GetContent()})
	}

	// 9.处理预设消息，将预设消息添加到用户消息前，先去删除用户的原始消息，然后补充一个新的代替
	// 在Go中，我们直接替换整个Messages数组
	state.Messages = presetMessages

	return nil
}

// llmNodeReACT 重写工具调用智能体的LLM节点
func (r *ReACTAgent) llmNodeReACT(ctx context.Context, state *entities.AgentState) (bool, error) {
	// 1.判断当前LLM是否支持tool_call，如果是则使用FunctionCallAgent的_llm_node
	if r.supportsToolCall() {
		return r.llmNode(ctx, *state)
	}

	// 2.检测当前Agent迭代次数是否符合需求
	if state.IterationCount > r.agentConfig.MaxIterationCount {
		r.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:      uuid.New(),
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: entities.MaxIterationResponse,
			Answer:  entities.MaxIterationResponse,
			Latency: 0,
		})
		return true, nil
	}

	// 3.从智能体配置中提取大语言模型
	id := uuid.New()
	startTime := time.Now()

	// 4.定义变量存储流式输出内容
	var gathered strings.Builder
	isFirstChunk := true
	generationType := ""

	// 5.转换ChatMessage到MessageContent
	messageContents := r.convertChatMessagesToMessageContents(state.Messages)

	// 6.流式输出调用LLM，并判断输出内容是否以"```json"为开头，用于区分工具调用和文本生成
	response, err := r.llm.GenerateContent(ctx, messageContents, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		content := string(chunk)

		// 6.处理流式输出内容块叠加
		if isFirstChunk {
			gathered.WriteString(content)
			isFirstChunk = false
		} else {
			gathered.WriteString(content)
		}

		// 7.如果生成的是消息则提交智能体消息事件
		if generationType == "message" {
			// 8.提取片段内容并检测是否开启输出审核
			reviewedContent := r.applyOutputReview(content)

			r.queueManager.Publish(state.TaskID, entities.AgentThought{
				ID:      id,
				TaskID:  state.TaskID,
				Event:   entities.EventAgentMessage,
				Thought: reviewedContent,
				Answer:  reviewedContent,
				Latency: time.Since(startTime).Seconds(),
			})
		}

		// 9.检测生成的类型是工具调用还是文本生成，同时赋值
		if generationType == "" {
			// 10.当生成内容的长度大于等于7(```json)长度时才可以判断出类型是什么
			currentContent := strings.TrimSpace(gathered.String())
			if len(currentContent) >= 7 {
				if strings.HasPrefix(currentContent, "```json") {
					generationType = "thought"
				} else {
					generationType = "message"
					// 11.添加发布事件，避免前几个字符遗漏
					r.queueManager.Publish(state.TaskID, entities.AgentThought{
						ID:      id,
						TaskID:  state.TaskID,
						Event:   entities.EventAgentMessage,
						Thought: currentContent,
						Answer:  currentContent,
						Latency: time.Since(startTime).Seconds(),
					})
				}
			}
		}

		return nil
	}))

	if err != nil {
		return false, fmt.Errorf("LLM generation failed: %w", err)
	}

	// 获取完整内容
	content := gathered.String()
	if response != nil && len(response.Choices) > 0 {
		content = response.Choices[0].Content
	}

	// 8.计算LLM的输入+输出token总数 (简化实现)
	inputTokenCount := r.estimateTokenCount(messageContents)
	outputTokenCount := len(content) / 4 // 简化的token计算

	// 9.获取输入/输出价格和单位 (简化实现)
	inputPrice, outputPrice, unit := 0.0, 0.0, 1.0

	// 10.计算总token+总成本
	totalTokenCount := inputTokenCount + outputTokenCount
	totalPrice := (float64(inputTokenCount)*inputPrice + float64(outputTokenCount)*outputPrice) * unit

	// 12.如果类型为推理则解析json，并添加智能体消息
	if generationType == "thought" {
		toolCalls, err := r.parseToolCallFromContent(content)
		if err != nil {
			// 13.解析失败，当作普通消息处理
			generationType = "message"
			r.queueManager.Publish(state.TaskID, entities.AgentThought{
				ID:      id,
				TaskID:  state.TaskID,
				Event:   entities.EventAgentMessage,
				Thought: content,
				Answer:  content,
				Latency: time.Since(startTime).Seconds(),
			})
		} else {
			r.queueManager.Publish(state.TaskID, entities.AgentThought{
				ID:                id,
				TaskID:            state.TaskID,
				Event:             entities.EventAgentThought,
				Thought:           content,
				MessageTokenCount: inputTokenCount,
				AnswerTokenCount:  outputTokenCount,
				TotalTokenCount:   totalTokenCount,
				TotalPrice:        totalPrice,
				Latency:           time.Since(startTime).Seconds(),
			})

			// 创建工具调用消息
			toolCallMsg := llms.AIChatMessage{
				Content:   "",
				ToolCalls: toolCalls,
			}
			state.Messages = append(state.Messages, toolCallMsg)
			state.IterationCount++
			return false, nil
		}
	}

	// 14.如果最终类型是message则表示已经拿到最终答案
	if generationType == "message" {
		r.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:                id,
			TaskID:            state.TaskID,
			Event:             entities.EventAgentMessage,
			Thought:           "",
			Answer:            "",
			MessageTokenCount: inputTokenCount,
			AnswerTokenCount:  outputTokenCount,
			TotalTokenCount:   totalTokenCount,
			TotalPrice:        totalPrice,
			Latency:           time.Since(startTime).Seconds(),
		})

		r.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:     uuid.New(),
			TaskID: state.TaskID,
			Event:  entities.EventAgentEnd,
		})

		return true, nil
	}

	// 添加生成的消息到状态
	aiMsg := llms.AIChatMessage{Content: content}
	state.Messages = append(state.Messages, aiMsg)
	state.IterationCount++

	return false, nil
}

// toolsNodeReACT 重写工具节点，处理工具节点的`AI工具调用参数消息`与`工具消息转人类消息`
func (r *ReACTAgent) toolsNodeReACT(state *entities.AgentState) error {
	if len(state.Messages) == 0 {
		return nil
	}

	lastMsg := state.Messages[len(state.Messages)-1]
	aiMsg, ok := lastMsg.(llms.AIChatMessage)
	if !ok || len(aiMsg.ToolCalls) == 0 {
		return nil
	}

	// 1.调用工具节点执行并获取结果
	toolResults := r.executeTools(aiMsg.ToolCalls, state.TaskID)

	// 2.移除原始的AI工具调用参数消息，并创建新的ai消息
	state.Messages = state.Messages[:len(state.Messages)-1]

	// 3.提取工具调用的第1条消息还原原始AI消息(ReACTAgent一次最多只有一个工具调用)
	if len(aiMsg.ToolCalls) > 0 {
		toolCallJSON := map[string]any{
			"name": aiMsg.ToolCalls[0].FunctionCall.Name,
			"args": aiMsg.ToolCalls[0].FunctionCall.Arguments,
		}
		jsonBytes, _ := json.Marshal(toolCallJSON)
		aiMessage := llms.AIChatMessage{Content: fmt.Sprintf("```json\n%s\n```", string(jsonBytes))}
		state.Messages = append(state.Messages, aiMessage)
	}

	// 4.将ToolMessage转换成HumanMessage，提升LLM的兼容性
	var content strings.Builder
	for _, result := range toolResults {
		content.WriteString(fmt.Sprintf("工具: %s\n执行结果: %s\n==========\n\n", result.Tool, result.Content))
	}

	humanMessage := llms.HumanChatMessage{Content: content.String()}
	state.Messages = append(state.Messages, humanMessage)

	return nil
}

// Helper methods

// convertChatMessagesToMessageContents 转换ChatMessage到MessageContent
func (r *ReACTAgent) convertChatMessagesToMessageContents(messages []llms.ChatMessage) []llms.MessageContent {
	var messageContents []llms.MessageContent

	for _, msg := range messages {
		var role llms.ChatMessageType
		switch msg.GetType() {
		case llms.ChatMessageTypeHuman:
			role = llms.ChatMessageTypeHuman
		case llms.ChatMessageTypeAI:
			role = llms.ChatMessageTypeAI
		case llms.ChatMessageTypeSystem:
			role = llms.ChatMessageTypeSystem
		case llms.ChatMessageTypeTool:
			role = llms.ChatMessageTypeTool
		default:
			role = llms.ChatMessageTypeHuman
		}

		messageContent := llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
		}

		messageContents = append(messageContents, messageContent)
	}

	return messageContents
}

// supportsToolCall checks if the LLM supports tool calling
func (r *ReACTAgent) supportsToolCall() bool {
	// 这里需要根据你的LLM接口实现来判断
	// 暂时返回false以使用ReACT模式
	return false
}

// supportsAgentThought checks if the LLM supports agent thought
func (r *ReACTAgent) supportsAgentThought() bool {
	// 这里需要根据你的LLM接口实现来判断
	return true
}

// renderToolDescription renders tool descriptions for the prompt
func (r *ReACTAgent) renderToolDescription() string {
	var descriptions []string
	for _, tool := range r.agentConfig.Tools {
		desc := fmt.Sprintf("%s - %s", tool.Name(), tool.Description())
		descriptions = append(descriptions, desc)
	}
	return strings.Join(descriptions, "\n")
}

// parseToolCallFromContent parses tool call information from LLM content
func (r *ReACTAgent) parseToolCallFromContent(content string) ([]llms.ToolCall, error) {
	// 13.使用正则解析信息，如果失败则当成普通消息返回
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no JSON found in content")
	}

	var toolCallData map[string]any
	jsonContent := strings.TrimSpace(matches[1])
	if err := json.Unmarshal([]byte(jsonContent), &toolCallData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	name, ok := toolCallData["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name not found")
	}

	args, ok := toolCallData["args"].(map[string]any)
	if !ok {
		args = make(map[string]any)
	}

	argsJSON, _ := json.Marshal(args)

	toolCall := llms.ToolCall{
		ID:   uuid.New().String(),
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      name,
			Arguments: string(argsJSON),
		},
	}

	return []llms.ToolCall{toolCall}, nil
}

// estimateTokenCount provides a simple token count estimation
func (r *ReACTAgent) estimateTokenCount(messages []llms.MessageContent) int {
	totalChars := 0
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				totalChars += len(textPart.Text)
			}
		}
	}
	// Rough estimation: 4 characters per token
	return totalChars / 4
}

// executeTools executes the given tool calls
func (r *ReACTAgent) executeTools(toolCalls []llms.ToolCall, taskID uuid.UUID) []ToolResult {
	var results []ToolResult

	// Convert tools to map for easy lookup
	toolsMap := make(map[string]tools.Tool)
	for _, tool := range r.agentConfig.Tools {
		toolsMap[tool.Name()] = tool
	}

	for _, toolCall := range toolCalls {
		id := uuid.New()
		startTime := time.Now()

		var result string
		var toolName string

		if toolCall.FunctionCall != nil {
			toolName = toolCall.FunctionCall.Name
			if tool, exists := toolsMap[toolName]; exists {
				// Execute tool
				if toolResult, err := tool.Call(context.Background(), toolCall.FunctionCall.Arguments); err == nil {
					result = toolResult
				} else {
					result = fmt.Sprintf("工具执行出错: %s", err.Error())
				}
			} else {
				result = fmt.Sprintf("工具不存在: %s", toolName)
			}
		}

		// Publish tool execution event
		event := entities.EventAgentAction
		if toolName == entities.DatasetRetrievalToolName {
			event = entities.EventDatasetRetrieval
		}

		r.queueManager.Publish(taskID, entities.AgentThought{
			ID:          id,
			TaskID:      taskID,
			Event:       event,
			Observation: result,
			Tool:        toolName,
			ToolInput:   toolCall.FunctionCall.Arguments,
			Latency:     time.Since(startTime).Seconds(),
		})

		results = append(results, ToolResult{
			Tool:    toolName,
			Content: result,
		})
	}

	return results
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Tool    string
	Content string
}
