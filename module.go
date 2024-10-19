package ioredis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

func ForRoot(opt *redis.Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		redisModule := module.New(core.NewModuleOptions{})
		redisModule.NewProvider(core.ProviderOptions{
			Name:  IO_REDIS,
			Value: New(opt),
		})
		redisModule.Export(IO_REDIS)

		return redisModule
	}
}

func InjectRedis(module *core.DynamicModule) *Redis {
	return module.Ref(IO_REDIS).(*Redis)
}

func InjectHash[M any](module *core.DynamicModule, name string) *Hash[M] {
	hash := module.Ref(core.Provide(getHashName(name)))
	if hash == nil {
		redis := module.Ref(IO_REDIS).(*Redis)
		hash = NewHash[M](name, redis)
		module.NewProvider(core.ProviderOptions{
			Name:  getHashName(name),
			Value: hash,
		})
	}
	return hash.(*Hash[M])
}

func getHashName(name string) core.Provide {
	hashName := fmt.Sprintf("%sHash", name)
	return core.Provide(hashName)
}
