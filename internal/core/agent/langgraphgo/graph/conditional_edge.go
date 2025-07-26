package graph

import (
	"context"
	"fmt"
	
	"github.com/tmc/langchaingo/llms"
)

// ConditionFunc is a function that determines the next node based on the state
type ConditionFunc func(ctx context.Context, state []llms.MessageContent) (string, error)

// ConditionalEdge represents an edge with a condition function
type ConditionalEdge struct {
	From      string
	Condition ConditionFunc
}

// AddConditionalEdges adds conditional edges from a node
func (g *MessageGraph) AddConditionalEdges(from string, conditionFn ConditionFunc) {
	g.conditionalEdges = append(g.conditionalEdges, ConditionalEdge{
		From:      from,
		Condition: conditionFn,
	})
}

// findNextNode finds the next node based on conditional edges
func (g *MessageGraph) findNextNode(ctx context.Context, currentNode string, state []llms.MessageContent) (string, error) {
	// First check conditional edges
	for _, edge := range g.conditionalEdges {
		if edge.From == currentNode {
			nextNode, err := edge.Condition(ctx, state)
			if err != nil {
				return "", fmt.Errorf("error in condition function for node %s: %w", currentNode, err)
			}
			return nextNode, nil
		}
	}

	// If no conditional edge matched, use regular edges
	for _, edge := range g.edges {
		if edge.From == currentNode {
			return edge.To, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrNoOutgoingEdge, currentNode)
}
