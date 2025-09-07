package service

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// DatasetTaskType 数据集任务类型
type DatasetTaskType string

const (
	TaskTypeDelete DatasetTaskType = "delete"
)

// DatasetTask 数据集任务结构
type DatasetTask struct {
	TaskType  DatasetTaskType `json:"task_type"`
	DatasetID uuid.UUID       `json:"dataset_id"`
}

// DatasetProducer 数据集任务生产者
type DatasetProducer struct {
	producer sarama.SyncProducer
}

// NewDatasetProducer 创建数据集任务生产者
func NewDatasetProducer(brokers []string) (*DatasetProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &DatasetProducer{
		producer: producer,
	}, nil
}

// Close 关闭生产者
func (p *DatasetProducer) Close() error {
	return p.producer.Close()
}

// PublishDeleteDatasetTask 发布删除数据集任务
func (p *DatasetProducer) PublishDeleteDatasetTask(ctx context.Context, datasetID uuid.UUID) error {
	task := DatasetTask{
		TaskType:  TaskTypeDelete,
		DatasetID: datasetID,
	}

	return p.publishTask(ctx, "dataset.delete", task)
}

// publishTask 发布任务到Kafka
func (p *DatasetProducer) publishTask(ctx context.Context, topic string, task DatasetTask) error {
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
