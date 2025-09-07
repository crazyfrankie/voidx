package consumer

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/document/task"
	"github.com/crazyfrankie/voidx/internal/index/service"
	"github.com/crazyfrankie/voidx/pkg/logs"
)

// DocumentConsumer 文档任务消费者
type DocumentConsumer struct {
	consumerGroup   sarama.ConsumerGroup
	indexingService *service.IndexingService
	topics          []string
}

// NewDocumentConsumer 创建文档任务消费者
func NewDocumentConsumer(brokers []string, groupID string, indexingService *service.IndexingService) (*DocumentConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &DocumentConsumer{
		consumerGroup:   consumerGroup,
		indexingService: indexingService,
		topics:          []string{"document.build", "document.update_enabled", "document.delete"},
	}, nil
}

// Start 启动消费者
func (c *DocumentConsumer) Start(ctx context.Context) error {
	handler := &documentConsumerGroupHandler{
		indexingService: c.indexingService,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				logs.Errorf("Error from consumer: %v", err)
				return err
			}
		}
	}
}

// Close 关闭消费者
func (c *DocumentConsumer) Close() error {
	return c.consumerGroup.Close()
}

// documentConsumerGroupHandler 消费者组处理器
type documentConsumerGroupHandler struct {
	indexingService *service.IndexingService
}

// Setup 设置消费者组
func (h *documentConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理消费者组
func (h *documentConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *documentConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			err := h.handleMessage(message)
			if err != nil {
				logs.Errorf("Failed to handle message: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage 处理消息
func (h *documentConsumerGroupHandler) handleMessage(message *sarama.ConsumerMessage) error {
	var documentTask task.DocumentTask
	if err := sonic.Unmarshal(message.Value, &documentTask); err != nil {
		logs.Errorf("Failed to unmarshal document task: %v", err)
		return err
	}

	ctx := context.Background()

	switch message.Topic {
	case "document.build":
		return h.handleBuildDocumentTask(ctx, documentTask)
	case "document.update_enabled":
		return h.handleUpdateDocumentEnabledTask(ctx, documentTask)
	case "document.delete":
		return h.handleDeleteDocumentTask(ctx, documentTask)
	default:
		logs.Errorf("Unknown topic: %s", message.Topic)
		return nil
	}
}

// handleBuildDocumentTask 处理构建文档任务
func (h *documentConsumerGroupHandler) handleBuildDocumentTask(ctx context.Context, documentTask task.DocumentTask) error {
	if documentTask.TaskType != task.TaskTypeBuild {
		return nil
	}

	err := h.indexingService.BuildDocuments(ctx, []uuid.UUID{documentTask.DocumentID})
	if err != nil {
		logs.Errorf("Failed to build document %s: %v", documentTask.DocumentID, err)
		return err
	}

	logs.Errorf("Successfully built document: %s", documentTask.DocumentID)
	return nil
}

// handleUpdateDocumentEnabledTask 处理更新文档启用状态任务
func (h *documentConsumerGroupHandler) handleUpdateDocumentEnabledTask(ctx context.Context, documentTask task.DocumentTask) error {
	if documentTask.TaskType != task.TaskTypeUpdateEnabled {
		return nil
	}

	err := h.indexingService.UpdateDocumentEnabled(ctx, documentTask.DocumentID)
	if err != nil {
		logs.Errorf("Failed to update document enabled status %s: %v", documentTask.DocumentID, err)
		return err
	}

	logs.Errorf("Successfully updated document enabled status: %s", documentTask.DocumentID)
	return nil
}

// handleDeleteDocumentTask 处理删除文档任务
func (h *documentConsumerGroupHandler) handleDeleteDocumentTask(ctx context.Context, documentTask task.DocumentTask) error {
	if documentTask.TaskType != task.TaskTypeDelete {
		return nil
	}

	err := h.indexingService.DeleteDocument(ctx, documentTask.DatasetID, documentTask.DocumentID)
	if err != nil {
		logs.Errorf("Failed to delete document %s: %v", documentTask.DocumentID, err)
		return err
	}

	logs.Errorf("Successfully deleted document: %s", documentTask.DocumentID)
	return nil
}
