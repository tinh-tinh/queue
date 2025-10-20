package queue_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

func Test_Proccessor(t *testing.T) {
	service := func(module core.Module) core.Provider {
		processor := queue.NewProcessor("test", module)

		processor.Process(func(job *queue.Job) {
			job.Process(func() error {
				num, err := strconv.Atoi(job.Id)
				require.Nil(t, err)
				if num%3 == 0 {
					return errors.New("error by test")
				}

				key, err := json.Marshal(job.Data)
				require.Nil(t, err)
				_, err = HeaveTask(string(key))
				require.Nil(t, err)
				return nil
			})
		})

		return processor
	}

	controller := func(module core.Module) core.Controller {
		ctrl := module.NewController("test")

		testQueue := queue.Inject(module, "test")
		ctrl.Post("", func(ctx core.Ctx) error {
			testQueue.AddJob(queue.AddJobOptions{
				Id:   "1",
				Data: "value 1",
			})
			return ctx.JSON(core.Map{
				"data": "ok",
			})
		})

		return ctrl
	}

	appModule := func() core.Module {
		return core.NewModule(core.NewModuleOptions{
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
				queue.Register("test"),
			},
			Controllers: []core.Controllers{controller},
			Providers:   []core.Providers{service},
		})
	}

	server := core.CreateFactory(appModule)
	server.SetGlobalPrefix("api")

	testServer := httptest.NewServer(server.PrepareBeforeListen())
	defer testServer.Close()

	testClient := testServer.Client()
	resp, err := testClient.Post(testServer.URL+"/api/test", "application/json", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	time.Sleep(1 * time.Second)
}

func TestNil(t *testing.T) {
	appModule := core.NewModule(core.NewModuleOptions{
		Imports: []core.Modules{},
	})

	require.Nil(t, queue.Inject(appModule, ""))
}
