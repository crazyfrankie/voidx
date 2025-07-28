package consumer

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"

	datasetService "github.com/crazyfrankie/voidx/internal/dataset/service"
	"github.com/crazyfrankie/voidx/internal/index/service"
)

// DatasetConsumer 数据集任务消费者
type DatasetConsumer struct {
	consumerGroup   sarama.ConsumerGroup
	indexingService *service.IndexingService
	topics          []string
}

// NewDatasetConsumer 创建数据集任务消费者
func NewDatasetConsumer(brokers []string, groupID string, indexingService *service.IndexingService) (*DatasetConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &DatasetConsumer{
		consumerGroup:   consumerGroup,
		indexingService: indexingService,
		topics:          []string{"dataset.delete"},
	}, nil
}

// Start 启动消费者
func (c *DatasetConsumer) Start(ctx context.Context) error {
	handler := &datasetConsumerGroupHandler{
		indexingService: c.indexingService,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				return err
			}
		}
	}
}

// Close 关闭消费者
func (c *DatasetConsumer) Close() error {
	return c.consumerGroup.Close()
}

// datasetConsumerGroupHandler 消费者组处理器
type datasetConsumerGroupHandler struct {
	indexingService *service.IndexingService
}

// Setup 设置消费者组
func (h *datasetConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理消费者组
func (h *datasetConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *datasetConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			err := h.handleMessage(message)
			if err != nil {
				log.Printf("Failed to handle message: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage 处理消息
func (h *datasetConsumerGroupHandler) handleMessage(message *sarama.ConsumerMessage) error {
	var datasetTask datasetService.DatasetTask
	if err := sonic.Unmarshal(message.Value, &datasetTask); err != nil {
		log.Printf("Failed to unmarshal dataset task: %v", err)
		return err
	}

	ctx := context.Background()

	switch message.Topic {
	case "dataset.delete":
		return h.handleDeleteDatasetTask(ctx, datasetTask)
	default:
		log.Printf("Unknown topic: %s", message.Topic)
		return nil
	}
}

// handleDeleteDatasetTask 处理删除数据集任务
func (h *datasetConsumerGroupHandler) handleDeleteDatasetTask(ctx context.Context, datasetTask datasetService.DatasetTask) error {
	if datasetTask.TaskType != datasetService.TaskTypeDelete {
		return nil
	}

	err := h.indexingService.DeleteDataset(ctx, datasetTask.DatasetID)
	if err != nil {
		log.Printf("Failed to delete dataset %s: %v", datasetTask.DatasetID, err)
		return err
	}

	log.Printf("Successfully deleted dataset: %s", datasetTask.DatasetID)
	return nil
}
