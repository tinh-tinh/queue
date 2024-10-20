package queue

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/ioredis"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Module(t *testing.T) {
	addr := "localhost:6379"
	module := core.NewModule(core.NewModuleOptions{
		Imports: []core.Module{
			ioredis.ForRoot(&redis.Options{
				Addr:     addr,
				DB:       0,
				Password: "",
			}),
			Register("jobs", &QueueOption{
				Workers:       6,
				RetryFailures: 3,
				Limiter: &RateLimiter{
					Max:      3,
					Duration: time.Millisecond,
				},
			}),
		},
	})

	queue := InjectQueue(module, "jobs")
	require.NotNil(t, queue)
}
