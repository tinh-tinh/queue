package trigger

import (
	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

const TRIGGER core.Provide = "TRIGGER_REDIS"

func ForRoot(opt *redis.Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		triggerModule := module.New(core.NewModuleOptions{})

		return triggerModule
	}
}
