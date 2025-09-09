package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

const pattern = "(?m)```json\\s*\\n([\\s\\S]+?)```"

// ReactAgent represents an agent that uses ReACT (Reasoning and Acting) methodology
// It inherits from FunctionCallAgent but overrides key methods for models without native tool calling
type ReactAgent struct {
	*FunctionCallAgent
}

// NewReactAgent creates a new ReACT agent
func NewReactAgent(llm model.BaseChatModel, config *entities.AgentConfig, queueFactory *AgentQueueManagerFactory) BaseAgent {
	functionCallAgent := NewFunctionCallAgent(llm, config, queueFactory).(*FunctionCallAgent)

	return &ReactAgent{FunctionCallAgent: functionCallAgent}
}

// Stream implements the BaseAgent interface for ReactAgent
func (r *ReactAgent) Stream(ctx context.Context, input entities.AgentState) (<-chan *entities.AgentThought, error) {
	// Ensure task ID is set
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}

	// Initialize other fields if not set
	if input.History == nil {
		input.History = make([]*schema.Message, 0)
	}

	// Create queue manager for this task
	queueManager := r.queueFactory.CreateManager(r.agentConfig.UserID, r.agentConfig.InvokeFrom)
	defer queueManager.Close()

	// Create queue for this task
	thoughtChan, err := queueManager.Listen(ctx, input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// Start processing in background
	go func() {
		defer func() {
			queueManager.Publish(input.TaskID, &entities.AgentThought{
				ID:     uuid.New(),
				TaskID: input.TaskID,
				Event:  entities.EventAgentEnd,
			})
		}()

		// Execute the agent processing pipeline
		if err := r.processAgentPipeline(ctx, input, queueManager); err != nil {
			queueManager.PublishError(input.TaskID, fmt.Errorf("Agent processing failed: %v", err))
		}
	}()

	return thoughtChan, nil
}

// processAgentPipeline overrides the pipeline to use ReACT-specific logic
func (r *ReactAgent) processAgentPipeline(ctx context.Context, state entities.AgentState, queueManager *AgentQueueManager) error {
	// 1. Preset operation node
	if shouldEnd, err := r.presetOperationNode(ctx, state, queueManager); err != nil {
		return err
	} else if shouldEnd {
		return nil
	}

	// 2. ReACT-specific long term memory recall node
	if err := r.reactLongTermMemoryRecallNode(ctx, &state, queueManager); err != nil {
		return err
	}

	// 3. Main processing loop with ReACT logic
	for state.IterationCount <= r.agentConfig.MaxIterationCount {
		// ReACT LLM node
		shouldEnd, err := r.reactLLMNode(ctx, &state, queueManager)
		if err != nil {
			return err
		}
		if shouldEnd {
			return nil
		}

		// Tools node (if needed)
		if len(state.Messages) > 0 {
			lastMsg := state.Messages[len(state.Messages)-1]
			if len(lastMsg.ToolCalls) > 0 {
				if err := r.reactToolsNode(ctx, &state, queueManager); err != nil {
					return err
				}
			} else {
				// No tool calls, we're done
				break
			}
		}

		state.IterationCount++
	}

	// Max iteration reached
	if state.IterationCount > r.agentConfig.MaxIterationCount {
		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:      uuid.New(),
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: entities.MaxIterationResponse,
			Answer:  entities.MaxIterationResponse,
			Latency: 0,
		})
	}

	return nil
}

// reactLongTermMemoryRecallNode handles long-term memory recall with ReACT-specific prompt
func (r *ReactAgent) reactLongTermMemoryRecallNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) error {
	// Check if model supports tool calling, if so use parent implementation
	if toolCallingModel, ok := r.llm.(model.ToolCallingChatModel); ok && len(r.agentConfig.Tools) > 0 {
		_ = toolCallingModel // Use parent implementation
		return r.longTermMemoryRecallNode(ctx, state, queueManager)
	}

	// Handle long-term memory
	longTermMemory := ""
	if r.agentConfig.EnableLongTermMemory && state.LongTermMemory != "" {
		longTermMemory = state.LongTermMemory
		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:          uuid.New(),
			TaskID:      state.TaskID,
			Event:       entities.EventLongTermMemoryRecall,
			Observation: longTermMemory,
		})
	}

	// Build tool description for ReACT prompt
	toolDescription := r.buildToolDescription(ctx)

	// Use ReACT-specific system prompt with tool descriptions
	systemPrompt := strings.ReplaceAll(entities.ReactAgentSystemPromptTemplate, "{preset_prompt}", r.agentConfig.PresetPrompt)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{long_term_memory}", longTermMemory)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{tool_description}", toolDescription)

	// Build message list
	messages := []*schema.Message{schema.SystemMessage(systemPrompt)}

	// Add history messages
	if len(state.History) > 0 {
		// Validate history format (should be alternating human/ai messages)
		if len(state.History)%2 != 0 {
			return fmt.Errorf("invalid history message format")
		}
		messages = append(messages, state.History...)
	}

	// Add current user message
	if len(state.Messages) > 0 {
		lastMsg := state.Messages[len(state.Messages)-1]
		messages = append(messages, schema.UserMessage(lastMsg.Content))
	}

	// Update state with prepared messages
	state.Messages = messages
	return nil
}

// reactLLMNode handles LLM processing with ReACT-specific logic
func (r *ReactAgent) reactLLMNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) (bool, error) {
	// Check if model supports tool calling, if so use parent implementation
	if toolCallingModel, ok := r.llm.(model.ToolCallingChatModel); ok && len(r.agentConfig.Tools) > 0 {
		_ = toolCallingModel // Use parent implementation
		return r.llmNode(ctx, state, queueManager)
	}

	// Check iteration count
	if state.IterationCount > r.agentConfig.MaxIterationCount {
		queueManager.Publish(state.TaskID, &entities.AgentThought{
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

	// Use streaming to detect tool calls vs regular messages
	streamReader, err := r.llm.Stream(ctx, state.Messages)
	if err != nil {
		queueManager.PublishError(state.TaskID, fmt.Errorf("LLM streaming failed: %v", err))
		return false, err
	}

	var gatheredContent strings.Builder
	var generationType string // "thought" for tool calls, "message" for regular response
	isFirstChunk := true

	// Process streaming chunks
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			break // End of stream
		}

		gatheredContent.WriteString(chunk.Content)

		// Determine generation type based on content
		if generationType == "" && gatheredContent.Len() >= 7 {
			content := strings.TrimSpace(gatheredContent.String())
			if strings.HasPrefix(content, "```json") {
				generationType = "thought"
			} else {
				generationType = "message"
				// Publish the initial content to avoid missing first characters
				r.publishMessageChunk(state.TaskID, id, content, startTime, queueManager)
			}
		}

		// If it's a regular message, publish streaming chunks
		if generationType == "message" && !isFirstChunk {
			content := r.applyOutputReview(chunk.Content)
			r.publishMessageChunk(state.TaskID, id, content, startTime, queueManager)
		}

		isFirstChunk = false
	}

	finalContent := gatheredContent.String()

	// Handle tool calling (thought) generation
	if generationType == "thought" {
		toolCalls, err := r.parseToolCallFromContent(finalContent)
		if err != nil {
			// If parsing fails, treat as regular message
			generationType = "message"
			content := r.applyOutputReview(finalContent)
			r.publishMessageChunk(state.TaskID, id, content, startTime, queueManager)
		} else {
			// Publish thought event
			queueManager.Publish(state.TaskID, &entities.AgentThought{
				ID:      id,
				TaskID:  state.TaskID,
				Event:   entities.EventAgentThought,
				Thought: finalContent,
				Latency: time.Since(startTime).Seconds(),
			})

			// Create AI message with tool calls
			aiMessage := &schema.Message{
				Role:      schema.Assistant,
				Content:   "",
				ToolCalls: toolCalls,
			}
			state.Messages = append(state.Messages, aiMessage)
			state.IterationCount++
			return false, nil // Continue processing (tools need to be executed)
		}
	}

	// Handle regular message generation
	if generationType == "message" {
		// Publish final statistics
		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:      id,
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: "",
			Answer:  "",
			Latency: time.Since(startTime).Seconds(),
		})

		// Create AI message
		aiMessage := &schema.Message{
			Role:    schema.Assistant,
			Content: r.applyOutputReview(finalContent),
		}
		state.Messages = append(state.Messages, aiMessage)
		state.IterationCount++
		return true, nil // End processing
	}

	return true, nil
}

// reactToolsNode handles tool execution with ReACT-specific message conversion
func (r *ReactAgent) reactToolsNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) error {
	// First execute tools using parent implementation
	if err := r.toolsNode(ctx, state, queueManager); err != nil {
		return err
	}

	// Find the AI message with tool calls and tool response messages
	var toolCallMessage *schema.Message
	var toolMessages []*schema.Message

	// Find the last AI message with tool calls
	for i := len(state.Messages) - 1; i >= 0; i-- {
		msg := state.Messages[i]
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			toolCallMessage = msg
			break
		}
	}

	// Find tool response messages after the tool call message
	if toolCallMessage != nil {
		for i := len(state.Messages) - 1; i >= 0; i-- {
			msg := state.Messages[i]
			if msg.Role == schema.Tool {
				toolMessages = append([]*schema.Message{msg}, toolMessages...) // Prepend to maintain order
			} else if msg == toolCallMessage {
				break
			}
		}
	}

	if toolCallMessage != nil && len(toolMessages) > 0 {
		// Remove the original AI tool call message from state
		newMessages := make([]*schema.Message, 0, len(state.Messages))
		for _, msg := range state.Messages {
			if msg != toolCallMessage {
				newMessages = append(newMessages, msg)
			}
		}

		// Remove tool messages from state
		for _, toolMsg := range toolMessages {
			for i, msg := range newMessages {
				if msg == toolMsg {
					newMessages = append(newMessages[:i], newMessages[i+1:]...)
					break
				}
			}
		}

		// Create new AI message with JSON format (ReACT style)
		if len(toolCallMessage.ToolCalls) > 0 {
			toolCall := toolCallMessage.ToolCalls[0]
			toolCallJSON := map[string]interface{}{
				"name": toolCall.Function.Name,
				"args": json.RawMessage(toolCall.Function.Arguments),
			}
			jsonBytes, _ := json.Marshal([]map[string]interface{}{toolCallJSON})
			aiMessage := &schema.Message{
				Role:    schema.Assistant,
				Content: fmt.Sprintf("```json\n%s\n```", string(jsonBytes)),
			}
			newMessages = append(newMessages, aiMessage)
		}

		// Convert tool messages to human message
		var contentBuilder strings.Builder
		for _, toolMsg := range toolMessages {
			contentBuilder.WriteString(fmt.Sprintf("工具: %s\n执行结果: %s\n==========\n\n",
				toolMsg.ToolName, toolMsg.Content))
		}

		humanMessage := &schema.Message{
			Role:    schema.User,
			Content: contentBuilder.String(),
		}
		newMessages = append(newMessages, humanMessage)

		// Update state messages
		state.Messages = newMessages
	}

	return nil
}

// buildToolDescription builds tool description text for ReACT prompt
func (r *ReactAgent) buildToolDescription(ctx context.Context) string {
	if len(r.agentConfig.Tools) == 0 {
		return ""
	}

	var descriptions []string
	for _, toolInterface := range r.agentConfig.Tools {
		if tool, ok := toolInterface.(interface {
			Info(context.Context) (*schema.ToolInfo, error)
		}); ok {
			toolInfo, err := tool.Info(ctx)
			if err != nil {
				continue
			}

			// Format: tool_name - description, args: {parameters}
			argsJSON, _ := json.Marshal(toolInfo.ParamsOneOf)
			description := fmt.Sprintf("%s - %s, args: %s",
				toolInfo.Name, toolInfo.Desc, string(argsJSON))
			descriptions = append(descriptions, description)
		}
	}

	return strings.Join(descriptions, "\n")
}

// parseToolCallFromContent parses tool call information from ReACT-style content
func (r *ReactAgent) parseToolCallFromContent(content string) ([]schema.ToolCall, error) {
	// Use regex to extract JSON from ```json...``` blocks
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(strings.TrimSpace(content))

	if len(matches) < 2 {
		return nil, fmt.Errorf("no JSON block found in content")
	}

	// Parse the JSON
	var toolCallData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(matches[1])), &toolCallData); err != nil {
		return nil, fmt.Errorf("failed to parse tool call JSON: %w", err)
	}

	// Extract tool name and args
	name, ok := toolCallData["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name not found or not a string")
	}

	args, ok := toolCallData["args"]
	if !ok {
		args = map[string]interface{}{}
	}

	// Convert args to JSON string
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool args: %w", err)
	}

	// Create tool call
	toolCall := schema.ToolCall{
		ID:   uuid.New().String(),
		Type: "function",
		Function: schema.FunctionCall{
			Name:      name,
			Arguments: string(argsJSON),
		},
	}

	return []schema.ToolCall{toolCall}, nil
}

// publishMessageChunk publishes a message chunk for streaming
func (r *ReactAgent) publishMessageChunk(taskID uuid.UUID, id uuid.UUID, content string, startTime time.Time, queueManager *AgentQueueManager) {
	reviewedContent := r.applyOutputReview(content)
	queueManager.Publish(taskID, &entities.AgentThought{
		ID:      id,
		TaskID:  taskID,
		Event:   entities.EventAgentMessage,
		Thought: reviewedContent,
		Answer:  reviewedContent,
		Latency: time.Since(startTime).Seconds(),
	})
}
