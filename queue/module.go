package queue

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

const QUEUE core.Provide = "QUEUE"

func getQueueName(name string) core.Provide {
	return core.Provide(fmt.Sprintf("%sQueue", name))
}

type Redis struct {
	client *redis.Client
}

func Register(name string, opt *QueueOption) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		queueModule := module.New(core.NewModuleOptions{})
		redis := module.Ref("IO_REDIS").(*Redis)

		queueModule.NewProvider(New(name, opt, redis.client), getQueueName(name))
		queueModule.Export(getQueueName(name))

		return queueModule
	}
}
