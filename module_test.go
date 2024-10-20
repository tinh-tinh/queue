package ioredis

import (
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/tinhtinh/common"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Module(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	userController := func(module *core.DynamicModule) *core.DynamicController {
		ctrl := module.NewController("users")

		ctrl.Get("connect", func(ctx core.Ctx) error {
			redisService := InjectRedis(module)

			return ctx.JSON(core.Map{
				"data": redisService.Ping(),
			})
		})

		ctrl.Post("", func(ctx core.Ctx) error {
			userService := InjectHash[User](module, "users")

			return ctx.JSON(core.Map{
				"data": userService.Upsert("tinh", &User{
					Name:  "tinh",
					Email: "tinh@tinh",
				}),
			})
		})

		ctrl.Get("", func(ctx core.Ctx) error {
			userService := InjectHash[User](module, "users")

			data, err := userService.FindByKey("tinh")
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}
			return ctx.JSON(core.Map{
				"data": data,
			})
		})

		return ctrl
	}

	userModule := func(module *core.DynamicModule) *core.DynamicModule {
		userModule := module.New(core.NewModuleOptions{
			Controllers: []core.Controller{userController},
		})

		return userModule
	}

	appModule := func() *core.DynamicModule {
		addr := "localhost:6379"
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Module{
				ForRoot(&redis.Options{
					Addr:     addr,
					DB:       0,
					Password: "",
				}),
				userModule,
			},
		})

		return module
	}

	app := core.CreateFactory(appModule)
	app.SetGlobalPrefix("api")

	testServer := httptest.NewServer(app.PrepareBeforeListen())
	defer testServer.Close()

	testClient := testServer.Client()

	resp, err := testClient.Get(testServer.URL + "/api/users/connect")
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	resp, err = testClient.Post(testServer.URL+"/api/users", "application/json", nil)
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	resp, err = testClient.Get(testServer.URL + "/api/users")
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)
}
