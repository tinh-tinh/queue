package ioredis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/tinh-tinh/ioredis/common"
)

type Redis struct {
	ctx    context.Context
	client *redis.Client
}

func New(opt *redis.Options) common.Redis {
	return &Redis{
		ctx:    context.Background(),
		client: redis.NewClient(opt),
	}
}

func (r *Redis) GetCtx() context.Context {
	return r.ctx
}

func (r *Redis) SetCtx(ctx context.Context) {
	r.ctx = ctx
}

func (r *Redis) GetClient() *redis.Client {
	return r.client
}
