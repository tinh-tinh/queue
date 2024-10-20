package ioredis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	addr := "localhost:6379"
	rdb := New(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	rdb.SetCtx(context.TODO())

	client := rdb.GetClient()
	require.NotNil(t, client)

	ctx := rdb.GetCtx()
	require.NotNil(t, ctx)

	err := client.Ping(ctx).Err()
	require.Nil(t, err)
}
