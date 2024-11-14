package queue

import (
	"fmt"

	"github.com/tinh-tinh/tinhtinh/core"
)

const QUEUE core.Provide = "QUEUE"

// getQueueName generates a unique name for a queue provider.
//
// The name is in the form "<name>Queue".
func getQueueName(name string) core.Provide {
	return core.Provide(fmt.Sprintf("%sQueue", name))
}

// Register registers a new queue module with the given name and options. The
// registered module creates a new queue with the given name and options, and
// exports the queue under the name "<name>Queue".
func Register(name string, opt *QueueOption) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		queueModule := module.New(core.NewModuleOptions{})

		queueModule.NewProvider(core.ProviderOptions{
			Name:  getQueueName(name),
			Value: New(name, opt),
		})
		queueModule.Export(getQueueName(name))

		return queueModule
	}
}

// InjectQueue injects a queue from the given module, using the given name. If the
// module does not contain a queue with the given name, or if the queue is not of
// type *Queue, InjectQueue returns nil.
func InjectQueue(module *core.DynamicModule, name string) *Queue {
	queue, ok := module.Ref(getQueueName(name)).(*Queue)
	if !ok {
		return nil
	}

	return queue
}
