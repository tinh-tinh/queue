package ioredis

import (
	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

const IOREDIS core.Provide = "IO_REDIS"

func Registry(opt *redis.Options) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		redisModule := module.New(core.NewModuleOptions{})
		redisModule.NewProvider(New(opt), IOREDIS)
		redisModule.Export(IOREDIS)

		return redisModule
	}
}
