package ioredis

import (
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

func Test_New(t *testing.T) {
	t.Run("Test_Conenct", func(t *testing.T) {
		addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
		rdb := New(&redis.Options{
			Addr:     addr,
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
