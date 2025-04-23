package queue

import (
	"fmt"

	"github.com/tinh-tinh/tinhtinh/v2/core"
)

const QUEUE core.Provide = "QUEUE"

func ForRoot(opt *Options) core.Modules {
	return func(module core.Module) core.Module {
		queueModule := module.New(core.NewModuleOptions{})

		queueModule.NewProvider(core.ProviderOptions{
			Name:  QUEUE,
			Value: opt,
		})
		queueModule.Export(QUEUE)

		return queueModule
	}
}

func ForRootFactory(factory func(ref core.RefProvider) *Options) core.Modules {
	return func(module core.Module) core.Module {
		opt := factory(module)
		queueModule := module.New(core.NewModuleOptions{})

		queueModule.NewProvider(core.ProviderOptions{
			Name:  QUEUE,
			Value: opt,
		})
		queueModule.Export(QUEUE)

		return queueModule
	}
}

// getQueueName generates a unique name for a queue provider.
//
// The name is in the form "<name>Queue".
func getQueueName(name string) core.Provide {
	return core.Provide(fmt.Sprintf("%sQueue", name))
}

// Register registers a new queue module with the given name and options. The
// registered module creates a new queue with the given name and options, and
// exports the queue under the name "<name>Queue".
func Register(name string, opts ...*Options) core.Modules {
	var option *Options
	if len(opts) > 0 {
		option = opts[0]
	}
	return func(module core.Module) core.Module {
		if option == nil {
			defaultOptions, ok := module.Ref(QUEUE).(*Options)
			if !ok || defaultOptions == nil {
				panic("not config option for queue")
			}
			option = defaultOptions
		}
		queueModule := module.New(core.NewModuleOptions{})

		queueModule.NewProvider(core.ProviderOptions{
			Name:  getQueueName(name),
			Value: New(name, option),
		})
		queueModule.Export(getQueueName(name))

		return queueModule
	}
}

// InjectQueue injects a queue from the given module, using the given name. If the
// module does not contain a queue with the given name, or if the queue is not of
// type *Queue, InjectQueue returns nil.
func Inject(module core.RefProvider, name string) *Queue {
	queue, ok := module.Ref(getQueueName(name)).(*Queue)
	if !ok {
		return nil
	}

	return queue
}
