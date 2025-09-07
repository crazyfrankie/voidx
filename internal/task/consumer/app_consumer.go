package consumer

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"

	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/pkg/logs"
)

// AppConsumer 应用任务消费者
type AppConsumer struct {
	consumerGroup sarama.ConsumerGroup
	appService    *service.AppService
	topics        []string
}

// NewAppConsumer 创建应用任务消费者
func NewAppConsumer(brokers []string, groupID string, appService *service.AppService) (*AppConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &AppConsumer{
		consumerGroup: consumerGroup,
		appService:    appService,
		topics:        []string{"app.auto_create"},
	}, nil
}

// Start 启动消费者
func (c *AppConsumer) Start(ctx context.Context) error {
	handler := &appConsumerGroupHandler{
		appService: c.appService,
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
func (c *AppConsumer) Close() error {
	return c.consumerGroup.Close()
}

// appConsumerGroupHandler 消费者组处理器
type appConsumerGroupHandler struct {
	appService *service.AppService
}

// Setup 设置消费者组
func (h *appConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理消费者组
func (h *appConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息
func (h *appConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
func (h *appConsumerGroupHandler) handleMessage(message *sarama.ConsumerMessage) error {
	var appTask task.AppTask
	if err := sonic.Unmarshal(message.Value, &appTask); err != nil {
		logs.Errorf("Failed to unmarshal app task: %v", err)
		return err
	}

	ctx := context.Background()

	switch message.Topic {
	case "app.auto_create":
		return h.handleAutoCreateAppTask(ctx, appTask)
	default:
		logs.Errorf("Unknown topic: %s", message.Topic)
		return nil
	}
}

// handleAutoCreateAppTask 处理自动创建应用任务
func (h *appConsumerGroupHandler) handleAutoCreateAppTask(ctx context.Context, appTask task.AppTask) error {
	if appTask.TaskType != task.TaskTypeAutoCreate {
		return nil
	}

	err := h.appService.AutoCreateApp(ctx, appTask.Name, appTask.Description, appTask.AccountID)
	if err != nil {
		logs.Errorf("Failed to auto create app %s: %v", appTask.Name, err)
		return err
	}

	logs.Errorf("Successfully auto created app: %s", appTask.Name)
	return nil
}
