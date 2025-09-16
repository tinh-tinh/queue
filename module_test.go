package queue_test

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

func Test_Module(t *testing.T) {
	addr := "localhost:6379"
	module := core.NewModule(core.NewModuleOptions{
		Imports: []core.Modules{
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

func Test_Panic(t *testing.T) {
	require.Panics(t, func() {
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Modules{
				queue.Register("jobs"),
			},
		})

		require.NotNil(t, module)
	})
}

func Test_DefaultOptions(t *testing.T) {
	module := core.NewModule(core.NewModuleOptions{
		Imports: []core.Modules{
			queue.ForRoot(&queue.Options{
				Connect: &redis.Options{
					Addr:     "localhost:6379",
					DB:       0,
					Password: "",
				},
				Workers:       6,
				RetryFailures: 3,
			}),
			queue.Register("video"),
			queue.Register("media", &queue.Options{
				Workers: 3,
			}),
		},
	})

	videoQueue := queue.Inject(module, "video")
	require.NotNil(t, videoQueue)

	mediaQueue := queue.Inject(module, "media")
	require.NotNil(t, mediaQueue)
}

func Test_Factory(t *testing.T) {
	module := core.NewModule(core.NewModuleOptions{
		Imports: []core.Modules{
			queue.ForRootFactory(func(ref core.RefProvider) *queue.Options {
				return &queue.Options{
					Connect: &redis.Options{
						Addr:     "localhost:6379",
						DB:       0,
						Password: "",
					},
					Workers:       6,
					RetryFailures: 3,
				}
			}),
			queue.Register("livestream"),
		},
	})

	videoQueue := queue.Inject(module, "livestream")
	require.NotNil(t, videoQueue)
}
