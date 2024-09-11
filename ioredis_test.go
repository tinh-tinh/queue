package ioredis

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func Test_New(t *testing.T) {
	t.Run("Test_Conenct", func(t *testing.T) {
		rdb := New(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})

		client := rdb.GetClient()
		if client == nil {
			t.Error("not connect redis")
			return
		}

		ctx := rdb.GetCtx()
		if ctx == nil {
			t.Error("not context")
			return
		}

		err := client.Ping(ctx).Err()
		if err != nil {
			t.Error("cannot ping")
		}
	})
}
