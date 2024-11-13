package trigger

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Trigger struct {
	client *redis.Client
}

func New(opt *redis.Options) *Trigger {
	return &Trigger{
		client: redis.NewClient(opt),
	}
}

func (trigger *Trigger) Emit(name string, data interface{}) {
	trigger.client.Publish(context.Background(), name, data)
}

func (trigger *Trigger) OnEvent(name string, cb func(msg string)) {
	pubsub := trigger.client.Subscribe(context.Background(), name)

	defer pubsub.Close()
	ch := pubsub.Channel()

	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
		cb(msg.Payload)
	}
}
