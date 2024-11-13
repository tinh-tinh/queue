package throttler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Throttler struct {
	redis      *redis.Client
	Max        int
	Duration   time.Duration
	Name       string
	skipRoutes []string
}

func New(opt *Options) *Throttler {
	throttler := &Throttler{
		Name:       opt.Name,
		redis:      redis.NewClient(opt.Redis),
		Max:        opt.Max,
		Duration:   opt.Duration,
		skipRoutes: opt.SkipRoutes,
	}

	return throttler
}

func (t *Throttler) generateKey(id string) string {
	return fmt.Sprintf("%s-%s", t.Name, id)
}

func (t *Throttler) Get(id string) int {
	key := t.generateKey(id)
	val, _ := t.redis.Get(context.Background(), key).Result()
	valNum, _ := strconv.Atoi(val)
	return valNum
}

func (t *Throttler) Incr(id string) {
	key := t.generateKey(id)
	val, _ := t.redis.Incr(context.Background(), key).Result()
	if val == 1 {
		t.redis.Expire(context.Background(), key, t.Duration)
	}
}
