package task

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// AppTaskType 应用任务类型
type AppTaskType string

const (
	TaskTypeAutoCreate AppTaskType = "auto_create"
)

// AppTask 应用任务结构
type AppTask struct {
	TaskType    AppTaskType `json:"task_type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	AccountID   uuid.UUID   `json:"account_id"`
}

// AppProducer 应用任务生产者
type AppProducer struct {
	producer sarama.SyncProducer
}

// NewAppProducer 创建应用任务生产者
func NewAppProducer(brokers []string) (*AppProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &AppProducer{
		producer: producer,
	}, nil
}

// Close 关闭生产者
func (p *AppProducer) Close() error {
	return p.producer.Close()
}

// PublishAutoCreateAppTask 发布自动创建应用任务
func (p *AppProducer) PublishAutoCreateAppTask(ctx context.Context, name, description string, accountID uuid.UUID) error {
	task := AppTask{
		TaskType:    TaskTypeAutoCreate,
		Name:        name,
		Description: description,
		AccountID:   accountID,
	}

	return p.publishTask(ctx, "app.auto_create", task)
}

// publishTask 发布任务到Kafka
func (p *AppProducer) publishTask(ctx context.Context, topic string, task AppTask) error {
	data, err := sonic.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s: %w", topic, err)
	}

	return nil
}
