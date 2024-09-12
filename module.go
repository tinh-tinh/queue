package ioredis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/ioredis/common"
	"github.com/tinh-tinh/tinhtinh/core"
)

func ForRoot(opt *redis.Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		redisModule := module.New(core.NewModuleOptions{})
		redisModule.NewProvider(New(opt), common.IOREDIS)
		redisModule.Export(common.IOREDIS)

		return redisModule
	}
}

func InjectRedis(module *core.DynamicModule) *Redis {
	return module.Ref(common.IOREDIS).(*Redis)
}

func InjectHash[M any](module *core.DynamicModule, name string) *Hash[M] {
	hash := module.Ref(core.Provide(getHashName(name)))
	if hash == nil {
		redis := module.Ref(common.IOREDIS).(*Redis)
		hash = NewHash[M](name, redis)
		module.NewProvider(hash, core.Provide(getHashName(name)))
	}
	return hash.(*Hash[M])
}

func getHashName(name string) string {
	return fmt.Sprintf("%sHash", name)
}
