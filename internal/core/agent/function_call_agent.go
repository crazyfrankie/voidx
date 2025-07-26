package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

// FunctionCallAgent represents an agent that uses function/tool calling capabilities
type FunctionCallAgent struct {
	*baseAgentImpl
}

// NewFunctionCallAgent creates a new function call agent
func NewFunctionCallAgent(llm llms.Model, config entities.AgentConfig, agentQueueManager *AgentQueueManager) BaseAgent {
	base := NewBaseAgent(llm, config, agentQueueManager).(*baseAgentImpl)
	return &FunctionCallAgent{baseAgentImpl: base}
}

// Stream implements the BaseAgent interface for FunctionCallAgent
func (f *FunctionCallAgent) Stream(ctx context.Context, input entities.AgentState) (<-chan entities.AgentThought, error) {
	// Ensure task ID is set
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}

	// Initialize other fields if not set
	if input.History == nil {
		input.History = make([]llms.ChatMessage, 0)
	}

	// Create queue for this task
	thoughtChan, err := f.queueManager.Listen(input.TaskID, f.agentConfig.UserID, f.agentConfig.InvokeFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// Start processing in background
	go func() {
		defer func() {
			f.queueManager.Publish(input.TaskID, entities.AgentThought{
				ID:     uuid.New(),
				TaskID: input.TaskID,
				Event:  entities.EventAgentEnd,
			})
		}()

		// Process through agent graph
		if err := f.processAgentGraph(ctx, input); err != nil {
			f.queueManager.PublishError(input.TaskID, fmt.Sprintf("Agent processing failed: %v", err))
		}
	}()

	return thoughtChan, nil
}

// processAgentGraph processes the input through the agent's execution graph
func (f *FunctionCallAgent) processAgentGraph(ctx context.Context, state entities.AgentState) error {
	// 1. Preset operation node
	if shouldEnd, err := f.presetOperationNode(state); err != nil {
		return err
	} else if shouldEnd {
		return nil
	}

	// 2. Long term memory recall node
	if err := f.longTermMemoryRecallNode(state); err != nil {
		return err
	}

	// 3. Main processing loop
	for state.IterationCount <= f.agentConfig.MaxIterationCount {
		// LLM node
		shouldEnd, err := f.llmNode(ctx, state)
		if err != nil {
			return err
		}
		if shouldEnd {
			return nil
		}

		state.IterationCount++
	}

	// Max iteration reached
	f.queueManager.Publish(state.TaskID, entities.AgentThought{
		ID:      uuid.New(),
		TaskID:  state.TaskID,
		Event:   entities.EventAgentMessage,
		Thought: entities.MaxIterationResponse,
		Answer:  entities.MaxIterationResponse,
		Latency: 0,
	})

	return nil
}

// presetOperationNode handles preset operations including input review
func (f *FunctionCallAgent) presetOperationNode(state entities.AgentState) (bool, error) {
	reviewConfig := f.agentConfig.ReviewConfig
	if !reviewConfig.Enable || !reviewConfig.InputsConfig.Enable {
		return false, nil
	}

	// Get query from last message
	if len(state.Messages) == 0 {
		return false, nil
	}

	query := extractQueryFromMessage(state.Messages[len(state.Messages)-1])

	// Check for sensitive keywords
	for _, keyword := range reviewConfig.Keywords {
		if strings.Contains(strings.ToLower(query), strings.ToLower(keyword)) {
			presetResponse := reviewConfig.InputsConfig.PresetResponse

			f.queueManager.Publish(state.TaskID, entities.AgentThought{
				ID:      uuid.New(),
				TaskID:  state.TaskID,
				Event:   entities.EventAgentMessage,
				Thought: presetResponse,
				Answer:  presetResponse,
				Latency: 0,
			})

			return true, nil
		}
	}

	return false, nil
}

// longTermMemoryRecallNode handles long-term memory recall
func (f *FunctionCallAgent) longTermMemoryRecallNode(state entities.AgentState) error {
	if !f.agentConfig.EnableLongTermMemory || state.LongTermMemory == "" {
		return nil
	}

	f.queueManager.Publish(state.TaskID, entities.AgentThought{
		ID:          uuid.New(),
		TaskID:      state.TaskID,
		Event:       entities.EventLongTermMemoryRecall,
		Observation: state.LongTermMemory,
	})

	return nil
}

// llmNode handles LLM processing
func (f *FunctionCallAgent) llmNode(ctx context.Context, state entities.AgentState) (bool, error) {
	// Check iteration count
	if state.IterationCount > f.agentConfig.MaxIterationCount {
		f.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:      uuid.New(),
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: entities.MaxIterationResponse,
			Answer:  entities.MaxIterationResponse,
			Latency: 0,
		})
		return true, nil
	}

	id := uuid.New()
	startTime := time.Now()

	// Prepare messages
	messages := f.prepareMessages(state)

	// Generate content using LLM
	response, err := f.llm.GenerateContent(ctx, messages)
	if err != nil {
		return false, fmt.Errorf("LLM generation failed: %w", err)
	}

	if len(response.Choices) > 0 {
		choice := response.Choices[0]
		content := choice.Content

		// Apply output review if enabled
		reviewedContent := f.applyOutputReview(content)

		f.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:      id,
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: reviewedContent,
			Answer:  reviewedContent,
			Latency: time.Since(startTime).Seconds(),
		})

		return true, nil
	}

	return false, nil
}

// prepareMessages prepares the messages for LLM processing
func (f *FunctionCallAgent) prepareMessages(state entities.AgentState) []llms.MessageContent {
	var messages []llms.MessageContent

	// Add system message with preset prompt and long-term memory
	systemPrompt := strings.ReplaceAll(entities.AgentSystemPromptTemplate, "{preset_prompt}", f.agentConfig.PresetPrompt)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{long_term_memory}", state.LongTermMemory)

	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
	})

	// Add history messages
	for _, histMsg := range state.History {
		messages = append(messages, convertChatMessageToMessageContent(histMsg))
	}

	// Add current messages
	for _, msg := range state.Messages {
		messages = append(messages, convertChatMessageToMessageContent(msg))
	}

	return messages
}

// applyOutputReview applies output content review
func (f *FunctionCallAgent) applyOutputReview(content string) string {
	reviewConfig := f.agentConfig.ReviewConfig
	if !reviewConfig.Enable || !reviewConfig.OutputsConfig.Enable {
		return content
	}

	reviewedContent := content
	for _, keyword := range reviewConfig.Keywords {
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(keyword))
		reviewedContent = re.ReplaceAllString(reviewedContent, "**")
	}

	return reviewedContent
}

func (f *FunctionCallAgent) toolsNode(ctx context.Context, state entities.AgentState) error {
	if len(state.Messages) == 0 {
		return nil
	}

	lastMsg := state.Messages[len(state.Messages)-1]
	aiMsg, ok := lastMsg.(llms.AIChatMessage)
	if !ok || len(aiMsg.ToolCalls) == 0 {
		return nil
	}

	// Convert tools to map for easy lookup
	toolsMap := make(map[string]tools.Tool)
	for _, tool := range f.agentConfig.Tools {
		toolsMap[tool.Name()] = tool
	}

	var toolMessages []llms.ChatMessage

	// Execute each tool call
	for _, toolCall := range aiMsg.ToolCalls {
		id := uuid.New()
		startTime := time.Now()

		var toolResult string
		var toolName string

		if toolCall.FunctionCall != nil {
			toolName = toolCall.FunctionCall.Name
			if tool, exists := toolsMap[toolName]; exists {
				// Execute tool
				if result, err := tool.Call(ctx, toolCall.FunctionCall.Arguments); err == nil {
					toolResult = result
				} else {
					toolResult = fmt.Sprintf("工具执行出错: %s", err.Error())
				}
			} else {
				toolResult = fmt.Sprintf("工具不存在: %s", toolName)
			}
		}

		// Create tool message
		toolMsg := llms.ToolChatMessage{
			Content: toolResult,
			ID:      toolCall.ID,
		}
		toolMessages = append(toolMessages, toolMsg)

		// Publish tool execution event
		event := entities.EventAgentAction
		if toolName == entities.DatasetRetrievalToolName {
			event = entities.EventDatasetRetrieval
		}

		f.queueManager.Publish(state.TaskID, entities.AgentThought{
			ID:          id,
			TaskID:      state.TaskID,
			Event:       event,
			Observation: toolResult,
			Tool:        toolName,
			ToolInput:   map[string]any{"args": toolCall.FunctionCall.Arguments},
			Latency:     time.Since(startTime).Seconds(),
		})
	}

	// Add tool messages to state
	state.Messages = append(state.Messages, toolMessages...)
	return nil
}

// Helper functions

// convertChatMessageToMessageContent converts llms.ChatMessage to llms.MessageContent
func convertChatMessageToMessageContent(msg llms.ChatMessage) llms.MessageContent {
	var role llms.ChatMessageType
	switch msg.GetType() {
	case llms.ChatMessageTypeHuman:
		role = llms.ChatMessageTypeHuman
	case llms.ChatMessageTypeAI:
		role = llms.ChatMessageTypeAI
	case llms.ChatMessageTypeSystem:
		role = llms.ChatMessageTypeSystem
	default:
		role = llms.ChatMessageTypeHuman
	}

	return llms.MessageContent{
		Role:  role,
		Parts: []llms.ContentPart{llms.TextPart(msg.GetContent())},
	}
}
