package task

import (
	"context"
	"sync"

	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/index"
	"github.com/crazyfrankie/voidx/internal/task/consumer"
	"github.com/crazyfrankie/voidx/pkg/logs"
)

// TaskManager 任务管理器
type TaskManager struct {
	documentConsumer *consumer.DocumentConsumer
	appConsumer      *consumer.AppConsumer
	datasetConsumer  *consumer.DatasetConsumer
	wg               sync.WaitGroup
}

// NewTaskManager 创建任务管理器
func NewTaskManager(brokers []string, indexingService *index.Service, appService *app.Service) (*TaskManager, error) {
	// 创建文档消费者
	documentConsumer, err := consumer.NewDocumentConsumer(brokers, "document-consumer-group", indexingService)
	if err != nil {
		return nil, err
	}

	// 创建应用消费者
	appConsumer, err := consumer.NewAppConsumer(brokers, "app-consumer-group", appService)
	if err != nil {
		return nil, err
	}

	// 创建数据集消费者
	datasetConsumer, err := consumer.NewDatasetConsumer(brokers, "dataset-consumer-group", indexingService)
	if err != nil {
		return nil, err
	}

	return &TaskManager{
		documentConsumer: documentConsumer,
		appConsumer:      appConsumer,
		datasetConsumer:  datasetConsumer,
	}, nil
}

// Start 启动所有消费者
func (m *TaskManager) Start(ctx context.Context) error {
	// 启动文档消费者
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.documentConsumer.Start(ctx); err != nil {
			logs.Errorf("Document consumer error: %v", err)
		}
	}()

	// 启动应用消费者
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.appConsumer.Start(ctx); err != nil {
			logs.Errorf("App consumer error: %v", err)
		}
	}()

	// 启动数据集消费者
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.datasetConsumer.Start(ctx); err != nil {
			logs.Errorf("Dataset consumer error: %v", err)
		}
	}()

	logs.Info("All task consumers started successfully")
	return nil
}

// Stop 停止所有消费者
func (m *TaskManager) Stop() error {
	if m.documentConsumer != nil {
		if err := m.documentConsumer.Close(); err != nil {
			logs.Errorf("Failed to close document consumer: %v", err)
		}
	}

	if m.appConsumer != nil {
		if err := m.appConsumer.Close(); err != nil {
			logs.Errorf("Failed to close app consumer: %v", err)
		}
	}

	if m.datasetConsumer != nil {
		if err := m.datasetConsumer.Close(); err != nil {
			logs.Errorf("Failed to close dataset consumer: %v", err)
		}
	}

	// 等待所有消费者停止
	m.wg.Wait()
	logs.Info("All task consumers stopped")
	return nil
}
