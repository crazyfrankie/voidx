package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
	"github.com/crazyfrankie/voidx/types/consts"
)

// AgentQueueManagerFactory 智能体队列管理器工厂
type AgentQueueManagerFactory struct {
	redisClient redis.Cmdable
	managers    sync.Map // key: string(userID-invokeFrom), value: *AgentQueueManager
}

// NewAgentQueueManagerFactory 创建队列管理器工厂
func NewAgentQueueManagerFactory(redisClient redis.Cmdable) *AgentQueueManagerFactory {
	return &AgentQueueManagerFactory{
		redisClient: redisClient,
	}
}

// GetOrCreateManager 获取或创建队列管理器（推荐用于长期任务）
func (f *AgentQueueManagerFactory) GetOrCreateManager(userID uuid.UUID, invokeFrom consts.InvokeFrom) *AgentQueueManager {
	key := fmt.Sprintf("%s-%s", userID.String(), invokeFrom)

	if manager, exists := f.managers.Load(key); exists {
		return manager.(*AgentQueueManager)
	}

	// 创建新的管理器
	manager := &AgentQueueManager{
		userID:      userID,
		invokeFrom:  invokeFrom,
		redisClient: f.redisClient,
		queues:      make(map[string]chan *entities.AgentThought),
		factory:     f, // 反向引用，用于清理
		managerKey:  key,
	}

	// 存储到 map 中
	f.managers.Store(key, manager)
	return manager
}

// CreateManager 为特定用户和调用来源创建临时队列管理器（用于一次性任务）
func (f *AgentQueueManagerFactory) CreateManager(userID uuid.UUID, invokeFrom consts.InvokeFrom) *AgentQueueManager {
	return &AgentQueueManager{
		userID:      userID,
		invokeFrom:  invokeFrom,
		redisClient: f.redisClient,
		queues:      make(map[string]chan *entities.AgentThought),
	}
}

// RemoveManager 移除管理器（通常在用户会话结束时调用）
func (f *AgentQueueManagerFactory) RemoveManager(userID uuid.UUID, invokeFrom consts.InvokeFrom) {
	key := fmt.Sprintf("%s-%s", userID.String(), invokeFrom)
	if manager, exists := f.managers.LoadAndDelete(key); exists {
		manager.(*AgentQueueManager).Close()
	}
}

// StopTask 停止特定任务（用于 StopDebug 等场景）
func (f *AgentQueueManagerFactory) StopTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, invokeFrom consts.InvokeFrom) error {
	return SetStopFlag(ctx, f.redisClient, taskID, invokeFrom, userID)
}

// AgentQueueManager 智能体队列管理器
type AgentQueueManager struct {
	userID      uuid.UUID
	invokeFrom  consts.InvokeFrom
	redisClient redis.Cmdable
	queues      map[string]chan *entities.AgentThought
	mu          sync.RWMutex

	// 用于持久化管理器的字段
	factory    *AgentQueueManagerFactory // 反向引用工厂
	managerKey string                    // 在工厂中的键
}

// Listen 监听队列返回的生成式数据
func (aqm *AgentQueueManager) Listen(ctx context.Context, taskID uuid.UUID) (<-chan *entities.AgentThought, error) {
	// 创建输出通道
	outputChan := make(chan *entities.AgentThought, 100)

	// 获取或创建任务队列
	queue := aqm.getQueue(taskID)

	// 启动监听协程
	go func() {
		defer close(outputChan)

		// 定义基础数据记录超时时间、开始时间、最后一次ping通时间
		listenTimeout := 600 * time.Second
		startTime := time.Now()
		lastPingTime := int64(0)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		pingTicker := time.NewTicker(10 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case item := <-queue:
				if item == nil {
					return // 队列已关闭
				}
				select {
				case outputChan <- item:
				case <-ctx.Done():
					return
				}
			case <-pingTicker.C:
				// 每10秒发起一个ping请求
				elapsed := time.Since(startTime)
				currentPingTime := int64(elapsed.Seconds()) / 10
				if currentPingTime > lastPingTime {
					pingThought := &entities.AgentThought{
						ID:        uuid.New(),
						TaskID:    taskID,
						Event:     entities.EventPing,
						CreatedAt: time.Now(),
					}
					aqm.Publish(taskID, pingThought)
					lastPingTime = currentPingTime
				}

				// 判断总耗时是否超时
				if elapsed >= listenTimeout {
					timeoutThought := &entities.AgentThought{
						ID:        uuid.New(),
						TaskID:    taskID,
						Event:     entities.EventTimeout,
						CreatedAt: time.Now(),
					}
					aqm.Publish(taskID, timeoutThought)
				}

				// 检测是否停止
				if aqm.isStopped(ctx, taskID) {
					stopThought := &entities.AgentThought{
						ID:        uuid.New(),
						TaskID:    taskID,
						Event:     entities.EventStop,
						CreatedAt: time.Now(),
					}
					aqm.Publish(taskID, stopThought)
				}
			}
		}
	}()

	return outputChan, nil
}

// StopListen 停止监听队列信息
func (aqm *AgentQueueManager) StopListen(taskID uuid.UUID) {
	aqm.mu.Lock()
	defer aqm.mu.Unlock()

	taskIDStr := taskID.String()
	if queue, exists := aqm.queues[taskIDStr]; exists {
		close(queue)
		delete(aqm.queues, taskIDStr)
	}
}

// Publish 发布事件信息到队列
func (aqm *AgentQueueManager) Publish(taskID uuid.UUID, agentThought *entities.AgentThought) {
	// 将事件添加到队列中
	queue := aqm.getQueue(taskID)
	agentThought.CreatedAt = time.Now()

	select {
	case queue <- agentThought:
	default:
		// 队列已满，跳过此消息以防止阻塞
	}

	// 检测事件类型是否为需要停止的类型
	if agentThought.Event == entities.EventStop ||
		agentThought.Event == entities.EventError ||
		agentThought.Event == entities.EventTimeout ||
		agentThought.Event == entities.EventAgentEnd {
		aqm.StopListen(taskID)
	}
}

// PublishError 发布错误信息到队列
func (aqm *AgentQueueManager) PublishError(taskID uuid.UUID, err error) {
	agentThought := &entities.AgentThought{
		ID:          uuid.New(),
		TaskID:      taskID,
		Event:       entities.EventError,
		Observation: err.Error(),
		CreatedAt:   time.Now(),
	}
	aqm.Publish(taskID, agentThought)
}

// isStopped 检测任务是否停止
func (aqm *AgentQueueManager) isStopped(ctx context.Context, taskID uuid.UUID) bool {
	taskStoppedCacheKey := aqm.generateTaskStoppedCacheKey(taskID)
	result := aqm.redisClient.Get(ctx, taskStoppedCacheKey)
	return result.Err() == nil
}

// getQueue 根据传递的taskID获取对应的任务队列信息
func (aqm *AgentQueueManager) getQueue(taskID uuid.UUID) chan *entities.AgentThought {
	aqm.mu.Lock()
	defer aqm.mu.Unlock()

	taskIDStr := taskID.String()
	queue, exists := aqm.queues[taskIDStr]

	if !exists {
		// 添加缓存键标识
		userPrefix := "account"
		if aqm.invokeFrom == consts.InvokeFromEndUser {
			userPrefix = "end-user"
		}

		// 设置任务对应的缓存键，代表这次任务已经开始了
		ctx := context.Background()
		aqm.redisClient.SetEx(ctx,
			aqm.generateTaskBelongCacheKey(taskID),
			fmt.Sprintf("%s-%s", userPrefix, aqm.userID.String()),
			30*time.Minute,
		)

		// 创建新队列
		queue = make(chan *entities.AgentThought, 1000)
		aqm.queues[taskIDStr] = queue
	}

	return queue
}

// SetStopFlag 根据传递的任务id+调用来源停止某次会话（静态方法，保持向后兼容）
func SetStopFlag(ctx context.Context, redisClient redis.Cmdable, taskID uuid.UUID, invokeFrom consts.InvokeFrom, userID uuid.UUID) error {
	aqm := &AgentQueueManager{redisClient: redisClient}

	// 获取当前任务的缓存键，如果任务没执行，则不需要停止
	result := redisClient.Get(ctx, aqm.generateTaskBelongCacheKey(taskID))
	if result.Err() != nil {
		return nil // 任务不存在，无需停止
	}

	// 计算对应缓存键的结果
	userPrefix := "account"
	if invokeFrom == consts.InvokeFromEndUser {
		userPrefix = "end-user"
	}

	expectedValue := fmt.Sprintf("%s-%s", userPrefix, userID.String())
	if result.Val() != expectedValue {
		return fmt.Errorf("unauthorized to stop task %s", taskID)
	}

	// 生成停止键标识
	stoppedCacheKey := aqm.generateTaskStoppedCacheKey(taskID)
	return redisClient.SetEx(ctx, stoppedCacheKey, "1", 10*time.Minute).Err()
}

// generateTaskBelongCacheKey 生成任务专属的缓存键
func (aqm *AgentQueueManager) generateTaskBelongCacheKey(taskID uuid.UUID) string {
	return fmt.Sprintf("generate_task_belong:%s", taskID.String())
}

// generateTaskStoppedCacheKey 生成任务已停止的缓存键
func (aqm *AgentQueueManager) generateTaskStoppedCacheKey(taskID uuid.UUID) string {
	return fmt.Sprintf("generate_task_stopped:%s", taskID.String())
}

// Close 关闭所有队列并清理资源
func (aqm *AgentQueueManager) Close() {
	aqm.mu.Lock()
	defer aqm.mu.Unlock()

	for taskID, queue := range aqm.queues {
		close(queue)
		delete(aqm.queues, taskID)
	}
}
