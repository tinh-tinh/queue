package throttler

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

// type Throttler struct {
// 	Max      int
// 	Duration time.Duration
// }

type Options struct {
	Redis      *redis.Options
	Name       string
	Max        int
	Duration   time.Duration
	SkipRoutes []string
}

func ForRoot(opt *Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		throttlerModule := module.New(core.NewModuleOptions{})
		throttlerModule.NewProvider(New(opt), core.Provide(opt.Name))
		throttlerModule.Export(core.Provide(opt.Name))

		return throttlerModule
	}
}
