# Agent Module

This module provides agent implementations for the LLMOps platform, including:

## Components

- Base Agent: Core agent functionality and interfaces
- Function Call Agent: Agent implementation using function/tool calling
- React Agent: Agent implementation using the ReAct pattern
- Agent Queue Manager: Manages agent execution queues and events
- Agent Entities: Data structures and configurations for agents

## Features

- Tool-based reasoning and execution
- Streaming support for real-time responses
- Queue-based event management
- Configurable agent behaviors
- Long-term memory support
- Graph-based execution flow using langgraphgo

## Implementation Details

The agent module is structured as follows:

1. **Base Agent**: Provides the core agent interface and base implementation
   - Invoke: Executes the agent and returns a complete result
   - Stream: Executes the agent and returns a channel of thoughts

2. **Function Call Agent**: Implements an agent that uses function/tool calling
   - Uses a graph-based execution flow
   - Supports tool execution and reasoning

3. **React Agent**: Implements an agent that uses the ReAct pattern
   - Similar to Function Call Agent but with different prompting

4. **Agent Queue Manager**: Manages event queues for agent execution
   - Publishes events to task-specific queues
   - Listens for events from specific tasks

5. **Agent Entities**: Defines data structures for agent configuration and state
   - AgentConfig: Configuration for agent behavior
   - AgentState: Current state of agent execution
   - AgentThought: Individual thought or action by the agent
   - AgentResult: Final result of agent execution

## Usage

```go
// Create agent configuration
config := entities.AgentConfig{
    UserID:              uuid.New(),
    InvokeFrom:          "web",
    MaxIterationCount:   5,
    SystemPrompt:        "You are a helpful assistant.",
    PresetPrompt:        "",
    EnableLongTermMemory: false,
    Tools:               []tools.Tool{},
    ReviewConfig:        map[string]any{
        "enable": false,
    },
}

// Create LLM instance
llm := // your LLM implementation

// Create function call agent
agent := NewFunctionCallAgent(llm, config)

// Create input state
input := entities.AgentState{
    TaskID:        uuid.New(),
    IterationCount: 0,
    History:       []llms.ChatMessage{},
    LongTermMemory: "",
    Messages: []llms.ChatMessage{
        &llms.HumanMessage{
            Content: "Hello, how are you?",
        },
    },
}

// For streaming response
thoughtChan, err := agent.Stream(context.Background(), input)
if err != nil {
    // Handle error
}
for thought := range thoughtChan {
    // Process each thought
}

// For blocking response
result, err := agent.Invoke(context.Background(), input)
if err != nil {
    // Handle error
}
// Use result.Answer
```
