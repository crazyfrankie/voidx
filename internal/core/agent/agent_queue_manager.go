package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

// AgentQueueManager manages the agent's event queue
type AgentQueueManager struct {
	redisClient redis.Cmdable
	mu          sync.RWMutex
	queues      map[uuid.UUID]chan entities.AgentThought
}

// NewAgentQueueManager creates a new AgentQueueManager
func NewAgentQueueManager(redisClient redis.Cmdable) *AgentQueueManager {
	return &AgentQueueManager{
		redisClient: redisClient,
		queues:      make(map[uuid.UUID]chan entities.AgentThought),
	}
}

// Listen creates a new queue for the specified task and returns a channel for receiving thoughts
func (m *AgentQueueManager) Listen(taskID uuid.UUID, userID uuid.UUID, invokeFrom consts.InvokeFrom) (<-chan entities.AgentThought, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.queues[taskID]; exists {
		return nil, fmt.Errorf("queue already exists for task %s", taskID)
	}

	// Create queue channel
	queue := make(chan entities.AgentThought, 100)
	m.queues[taskID] = queue

	// Set task belong cache key
	userPrefix := "account"
	if invokeFrom == consts.InvokeFromEndUser {
		userPrefix = "end-user"
	}

	ctx := context.Background()
	belongKey := generateTaskBelongCacheKey(taskID)
	err := m.redisClient.SetEx(ctx, belongKey, fmt.Sprintf("%s-%s", userPrefix, userID.String()), 30*time.Minute).Err()
	if err != nil {
		close(queue)
		delete(m.queues, taskID)
		return nil, fmt.Errorf("failed to set task belong cache: %w", err)
	}

	// Start listening goroutine
	go m.listenLoop(taskID, queue)

	return queue, nil
}

// listenLoop handles the listening loop with timeout and ping logic
func (m *AgentQueueManager) listenLoop(taskID uuid.UUID, queue chan entities.AgentThought) {
	defer func() {
		m.mu.Lock()
		close(queue)
		delete(m.queues, taskID)
		m.mu.Unlock()
	}()

	listenTimeout := 10 * time.Minute
	startTime := time.Now()
	lastPingTime := int64(0)
	pingTicker := time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			elapsed := time.Since(startTime)
			currentPingTime := int64(elapsed.Seconds()) / 10

			// Send ping every 10 seconds
			if currentPingTime > lastPingTime {
				m.Publish(taskID, entities.AgentThought{
					ID:     uuid.New(),
					TaskID: taskID,
					Event:  entities.EventPing,
				})
				lastPingTime = currentPingTime
			}

			// Check timeout
			if elapsed >= listenTimeout {
				m.Publish(taskID, entities.AgentThought{
					ID:     uuid.New(),
					TaskID: taskID,
					Event:  entities.EventTimeout,
				})
				return
			}

			// Check if stopped
			if m.isStopped(taskID) {
				m.Publish(taskID, entities.AgentThought{
					ID:     uuid.New(),
					TaskID: taskID,
					Event:  entities.EventStop,
				})
				return
			}

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Publish publishes a thought to the specified task queue
func (m *AgentQueueManager) Publish(taskID uuid.UUID, thought entities.AgentThought) error {
	m.mu.RLock()
	queue, exists := m.queues[taskID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("queue not found for task %s", taskID)
	}

	select {
	case queue <- thought:
		// Check if this is a terminating event
		if thought.Event == entities.EventStop ||
			thought.Event == entities.EventError ||
			thought.Event == entities.EventTimeout ||
			thought.Event == entities.EventAgentEnd {
			m.stopListen(taskID)
		}
		return nil
	default:
		return fmt.Errorf("queue is full for task %s", taskID)
	}
}

// PublishError publishes an error event to the specified task queue
func (m *AgentQueueManager) PublishError(taskID uuid.UUID, errMsg string) error {
	return m.Publish(taskID, entities.AgentThought{
		ID:          uuid.New(),
		TaskID:      taskID,
		Event:       entities.EventError,
		Observation: errMsg,
	})
}

// stopListen stops listening for the specified task
func (m *AgentQueueManager) stopListen(taskID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if queue, exists := m.queues[taskID]; exists {
		close(queue)
		delete(m.queues, taskID)
	}
}

// isStopped checks if the task has been stopped
func (m *AgentQueueManager) isStopped(taskID uuid.UUID) bool {
	ctx := context.Background()
	stoppedKey := generateTaskStoppedCacheKey(taskID)
	result := m.redisClient.Get(ctx, stoppedKey)
	return result.Err() == nil
}

// SetStopFlag sets the stop flag for a task
func (m *AgentQueueManager) SetStopFlag(taskID uuid.UUID, invokeFrom consts.InvokeFrom, userID uuid.UUID) error {
	ctx := context.Background()

	// Check if task belongs to user
	belongKey := generateTaskBelongCacheKey(taskID)
	result := m.redisClient.Get(ctx, belongKey)
	if result.Err() != nil {
		return nil // Task doesn't exist, no need to stop
	}

	userPrefix := "account"
	if invokeFrom == consts.InvokeFromEndUser {
		userPrefix = "end-user"
	}

	expectedValue := fmt.Sprintf("%s-%s", userPrefix, userID.String())
	if result.Val() != expectedValue {
		return nil // Task doesn't belong to this user
	}

	// Set stop flag
	stoppedKey := generateTaskStoppedCacheKey(taskID)
	return m.redisClient.SetEx(ctx, stoppedKey, "1", 10*time.Minute).Err()
}

// generateTaskBelongCacheKey generates the cache key for task ownership
func generateTaskBelongCacheKey(taskID uuid.UUID) string {
	return fmt.Sprintf("generate_task_belong:%s", taskID.String())
}

// generateTaskStoppedCacheKey generates the cache key for task stop status
func generateTaskStoppedCacheKey(taskID uuid.UUID) string {
	return fmt.Sprintf("generate_task_stopped:%s", taskID.String())
}
