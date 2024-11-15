package queue_test

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Module(t *testing.T) {
	addr := "localhost:6379"
	module := core.NewModule(core.NewModuleOptions{
		Imports: []core.Module{
			queue.Register("jobs", &queue.Options{
				Connect: &redis.Options{
					Addr:     addr,
					DB:       0,
					Password: "",
				},
				Workers:       6,
				RetryFailures: 3,
			}),
		},
	})

	queue := queue.Inject(module, "jobs")
	require.NotNil(t, queue)
}
