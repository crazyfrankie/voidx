package ioc

import (
	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/index"
	"github.com/crazyfrankie/voidx/internal/task"
)

func InitTask(indexService *index.Service, appService *app.Service) *task.TaskManager {
	manager, err := task.NewTaskManager(conf.GetConf().Kafka.Brokers, indexService, appService)
	if err != nil {
		panic(err)
	}

	return manager
}
