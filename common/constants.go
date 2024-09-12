package common

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/tinhtinh/core"
)

const IOREDIS core.Provide = "IO_REDIS"

type Redis interface {
	GetCtx() context.Context
	SetCtx(context.Context)
	GetClient() *redis.Client
}
