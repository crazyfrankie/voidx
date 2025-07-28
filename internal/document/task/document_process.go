package task

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

// DocumentTaskType 文档任务类型
type DocumentTaskType string

const (
	TaskTypeBuild         DocumentTaskType = "build"
	TaskTypeUpdateEnabled DocumentTaskType = "update_enabled"
	TaskTypeDelete        DocumentTaskType = "delete"
)

// DocumentTask 文档任务结构
type DocumentTask struct {
	TaskType   DocumentTaskType `json:"task_type"`
	DocumentID uuid.UUID        `json:"document_id"`
	DatasetID  uuid.UUID        `json:"dataset_id,omitempty"`
}

// DocumentProducer 文档任务生产者
type DocumentProducer struct {
	producer sarama.SyncProducer
}

// NewDocumentProducer 创建文档任务生产者
func NewDocumentProducer(brokers []string) (*DocumentProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &DocumentProducer{
		producer: producer,
	}, nil
}

// Close 关闭生产者
func (p *DocumentProducer) Close() error {
	return p.producer.Close()
}

// PublishBuildDocumentTask 发布构建文档任务
func (p *DocumentProducer) PublishBuildDocumentTask(ctx context.Context, documentID uuid.UUID) error {
	task := DocumentTask{
		TaskType:   TaskTypeBuild,
		DocumentID: documentID,
	}

	return p.publishTask(ctx, "document.build", task)
}

// PublishUpdateDocumentEnabledTask 发布更新文档启用状态任务
func (p *DocumentProducer) PublishUpdateDocumentEnabledTask(ctx context.Context, documentID uuid.UUID) error {
	task := DocumentTask{
		TaskType:   TaskTypeUpdateEnabled,
		DocumentID: documentID,
	}

	return p.publishTask(ctx, "document.update_enabled", task)
}

// PublishDeleteDocumentTask 发布删除文档任务
func (p *DocumentProducer) PublishDeleteDocumentTask(ctx context.Context, datasetID, documentID uuid.UUID) error {
	task := DocumentTask{
		TaskType:   TaskTypeDelete,
		DocumentID: documentID,
		DatasetID:  datasetID,
	}

	return p.publishTask(ctx, "document.delete", task)
}

// publishTask 发布任务到Kafka
func (p *DocumentProducer) publishTask(ctx context.Context, topic string, task DocumentTask) error {
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
