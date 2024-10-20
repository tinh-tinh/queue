package throttler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Throttler(t *testing.T) {
	appController := func(module *core.DynamicModule) *core.DynamicController {
		ctrl := module.NewController("throttler")

		ctrl.Guard(Guard("throttler")).Get("", func(ctx core.Ctx) error {
			return ctx.JSON(core.Map{
				"data": "OK",
			})
		})

		return ctrl
	}

	appModule := func() *core.DynamicModule {
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Module{
				Register(&Options{
					Redis:    &redis.Options{Addr: "localhost:6379"},
					Name:     "throttler",
					Max:      10,
					Duration: time.Second,
				}),
			},
			Controllers: []core.Controller{
				appController,
			},
		})

		return module
	}

	app := core.CreateFactory(appModule)
	app.SetGlobalPrefix("api")

	testServer := httptest.NewServer(app.PrepareBeforeListen())
	defer testServer.Close()

	testClient := testServer.Client()

	req, err := http.NewRequest("GET", testServer.URL+"/api/throttler", nil)
	require.Nil(t, err)

	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	for i := 0; i < 20; i++ {
		resp, err := testClient.Do(req)
		require.Nil(t, err)
		if i <= 10 {
			require.Equal(t, http.StatusOK, resp.StatusCode)
		} else {
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
		}
	}
}
