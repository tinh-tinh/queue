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

func Register(opt *Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		throttlerModule := module.New(core.NewModuleOptions{})
		provideName := core.Provide(opt.Name)
		throttlerModule.NewProvider(core.ProviderOptions{
			Name:  provideName,
			Value: New(opt),
		})
		throttlerModule.Export(provideName)

		return throttlerModule
	}
}
