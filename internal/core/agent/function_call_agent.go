package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

// FunctionCallAgent represents an agent that uses function/tool calling capabilities
type FunctionCallAgent struct {
	*baseAgentImpl
}

// NewFunctionCallAgent creates a new function call agent
func NewFunctionCallAgent(llm model.BaseChatModel, config *entities.AgentConfig, queueFactory *AgentQueueManagerFactory) BaseAgent {
	base := NewBaseAgent(llm, config, queueFactory).(*baseAgentImpl)

	return &FunctionCallAgent{baseAgentImpl: base}
}

// Stream implements the BaseAgent interface for FunctionCallAgent
func (f *FunctionCallAgent) Stream(ctx context.Context, input entities.AgentState) (<-chan *entities.AgentThought, error) {
	// Ensure task ID is set
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}

	// Initialize other fields if not set
	if input.History == nil {
		input.History = make([]*schema.Message, 0)
	}

	// Create queue manager for this task
	queueManager := f.queueFactory.CreateManager(f.agentConfig.UserID, f.agentConfig.InvokeFrom)
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
		if err := f.processAgentPipeline(ctx, input, queueManager); err != nil {
			queueManager.PublishError(input.TaskID, fmt.Errorf("Agent processing failed: %v", err))
		}
	}()

	return thoughtChan, nil
}

// processAgentPipeline processes the input through the agent's execution pipeline
func (f *FunctionCallAgent) processAgentPipeline(ctx context.Context, state entities.AgentState, queueManager *AgentQueueManager) error {
	// 1. Preset operation node
	if shouldEnd, err := f.presetOperationNode(ctx, state, queueManager); err != nil {
		return err
	} else if shouldEnd {
		return nil
	}

	// 2. Long term memory recall node
	if err := f.longTermMemoryRecallNode(ctx, &state, queueManager); err != nil {
		return err
	}

	// 3. Main processing loop
	for state.IterationCount <= f.agentConfig.MaxIterationCount {
		// LLM node
		shouldEnd, err := f.llmNode(ctx, &state, queueManager)
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
				if err := f.toolsNode(ctx, &state, queueManager); err != nil {
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
	if state.IterationCount > f.agentConfig.MaxIterationCount {
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

// presetOperationNode handles preset operations including input review
func (f *FunctionCallAgent) presetOperationNode(ctx context.Context, state entities.AgentState, queueManager *AgentQueueManager) (bool, error) {
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

			queueManager.Publish(state.TaskID, &entities.AgentThought{
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
func (f *FunctionCallAgent) longTermMemoryRecallNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) error {
	if !f.agentConfig.EnableLongTermMemory || state.LongTermMemory == "" {
		return nil
	}

	queueManager.Publish(state.TaskID, &entities.AgentThought{
		ID:          uuid.New(),
		TaskID:      state.TaskID,
		Event:       entities.EventLongTermMemoryRecall,
		Observation: state.LongTermMemory,
	})

	// Prepare system message with preset prompt and long-term memory
	systemPrompt := strings.ReplaceAll(entities.AgentSystemPromptTemplate, "{preset_prompt}", f.agentConfig.PresetPrompt)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{long_term_memory}", state.LongTermMemory)

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

// llmNode handles LLM processing
func (f *FunctionCallAgent) llmNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) (bool, error) {
	// Check iteration count
	if state.IterationCount > f.agentConfig.MaxIterationCount {
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

	// Use the LLM to generate response
	var llmModel model.BaseChatModel = f.llm

	// If the model supports tool calling and we have tools, bind them
	if toolCallingModel, ok := f.llm.(model.ToolCallingChatModel); ok && len(f.agentConfig.Tools) > 0 {
		// Convert tools to schema.ToolInfo
		toolInfos := make([]*schema.ToolInfo, 0, len(f.agentConfig.Tools))
		for _, toolInterface := range f.agentConfig.Tools {
			if tool, ok := toolInterface.(interface {
				Info(context.Context) (*schema.ToolInfo, error)
			}); ok {
				toolInfo, err := tool.Info(ctx)
				if err != nil {
					return false, fmt.Errorf("failed to get tool info: %w", err)
				}
				toolInfos = append(toolInfos, toolInfo)
			}
		}

		var err error
		llmModel, err = toolCallingModel.WithTools(toolInfos)
		if err != nil {
			return false, fmt.Errorf("failed to bind tools: %w", err)
		}
	}

	// Generate response
	response, err := llmModel.Generate(ctx, state.Messages)
	if err != nil {
		queueManager.PublishError(state.TaskID, fmt.Errorf("LLM generation failed: %v", err))
		return false, err
	}

	// Apply output review if enabled
	content := f.applyOutputReview(response.Content)

	// Check if response has tool calls
	if len(response.ToolCalls) > 0 {
		// This is a thought (tool calling)
		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:      id,
			TaskID:  state.TaskID,
			Event:   entities.EventAgentThought,
			Thought: fmt.Sprintf("Tool calls: %v", response.ToolCalls),
			Latency: time.Since(startTime).Seconds(),
		})
	} else {
		// This is a final message
		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:      id,
			TaskID:  state.TaskID,
			Event:   entities.EventAgentMessage,
			Thought: content,
			Answer:  content,
			Latency: time.Since(startTime).Seconds(),
		})
	}

	// Update state
	state.Messages = append(state.Messages, response)
	state.IterationCount++

	return len(response.ToolCalls) == 0, nil // Return true if no tool calls (end processing)
}

// toolsNode handles tool execution
func (f *FunctionCallAgent) toolsNode(ctx context.Context, state *entities.AgentState, queueManager *AgentQueueManager) error {
	if len(state.Messages) == 0 {
		return nil
	}

	lastMsg := state.Messages[len(state.Messages)-1]
	if len(lastMsg.ToolCalls) == 0 {
		return nil
	}

	// Convert tools to map for easy lookup
	toolsMap := make(map[string]interface{})
	for _, toolInterface := range f.agentConfig.Tools {
		if tool, ok := toolInterface.(interface {
			Info(context.Context) (*schema.ToolInfo, error)
		}); ok {
			toolInfo, err := tool.Info(ctx)
			if err != nil {
				continue
			}
			toolsMap[toolInfo.Name] = toolInterface
		}
	}

	var toolMessages []*schema.Message

	// Execute each tool call
	for _, toolCall := range lastMsg.ToolCalls {
		id := uuid.New()
		startTime := time.Now()

		var toolResult string
		var toolName string

		if toolCall.Function.Name != "" {
			toolName = toolCall.Function.Name
			if toolInterface, exists := toolsMap[toolName]; exists {
				// Try to cast to eino's InvokableTool interface
				if tool, ok := toolInterface.(interface {
					InvokableRun(context.Context, string, ...interface{}) (string, error)
				}); ok {
					// Execute tool with correct signature (no options for now)
					if result, err := tool.InvokableRun(ctx, toolCall.Function.Arguments); err == nil {
						toolResult = result
					} else {
						toolResult = fmt.Sprintf("工具执行出错: %s", err.Error())
					}
				} else {
					toolResult = fmt.Sprintf("工具类型不匹配: %s", toolName)
				}
			} else {
				toolResult = fmt.Sprintf("工具不存在: %s", toolName)
			}
		}

		// Create tool message
		toolMsg := schema.ToolMessage(toolResult, toolCall.ID, schema.WithToolName(toolName))
		toolMessages = append(toolMessages, toolMsg)

		// Publish tool execution event
		event := entities.EventAgentAction
		if toolName == entities.DatasetRetrievalToolName {
			event = entities.EventDatasetRetrieval
		}

		queueManager.Publish(state.TaskID, &entities.AgentThought{
			ID:          id,
			TaskID:      state.TaskID,
			Event:       event,
			Observation: toolResult,
			Tool:        toolName,
			ToolInput:   map[string]interface{}{"args": toolCall.Function.Arguments},
			Latency:     time.Since(startTime).Seconds(),
		})
	}

	// Add tool messages to state
	state.Messages = append(state.Messages, toolMessages...)
	return nil
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
