package trigger

import (
	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

const TRIGGER core.Provide = "TRIGGER_REDIS"

func ForRoot(opt *redis.Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		triggerModule := module.New(core.NewModuleOptions{})

		triggerModule.NewProvider(core.ProviderOptions{
			Name:  TRIGGER,
			Value: New(opt),
		})
		triggerModule.Export(TRIGGER)

		return triggerModule
	}
}

func Inject(module *core.DynamicModule) *Trigger {
	trigger, ok := module.Ref(TRIGGER).(*Trigger)
	if !ok {
		return nil
	}
	return trigger
}
