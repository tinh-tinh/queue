package throttler

import (
	"github.com/tinh-tinh/tinhtinh/core"
)

func Guard(name string) core.Guard {
	return func(ctrl *core.DynamicController, ctx *core.Ctx) bool {
		throttler := ctrl.Inject(core.Provide(name)).(*Throttler)
		ip := ctx.Headers("X-Real-Ip")
		if ip == "" {
			ip = ctx.Headers("X-Forwarded-For")
		}

		if ip == "" {
			ip = ctx.Req().RemoteAddr
		}

		if ip == "" {
			ip = ctx.Req().RemoteAddr
		}

		hits := throttler.Get(ip)
		if hits > throttler.Max {
			return false
		}
		throttler.Incr(ip)

		return true
	}
}
