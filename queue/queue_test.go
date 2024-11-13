package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func Test_Queue(t *testing.T) {
	addr := "localhost:6379"
	userQueue := New("user", &QueueOption{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       6,
		RetryFailures: 3,
		Limiter: &RateLimiter{
			Max:      3,
			Duration: time.Millisecond,
		},
	})

	t.Parallel()

	time.Sleep(time.Millisecond)
	userQueue.AddJob("1", "value 1")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("2", "value 2")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("3", "value 3")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("4", "value 4")
	time.Sleep(time.Second)
	userQueue.AddJob("5", "value 5")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("6", "value 6")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("7", "value 7")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("8", "value 8")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("9", "value 9")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("10", "value 10")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("11", "value 11")

	userQueue.Process(func(job *Job) {
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
}

func Test_SchedulerQueue(t *testing.T) {
	addr := "localhost:6379"
	userQueue := New("user", &QueueOption{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       6,
		RetryFailures: 3,
		Limiter: &RateLimiter{
			Max:      3,
			Duration: time.Second,
		},
		Pattern: "@every 0h0m1s",
	})

	t.Parallel()

	time.Sleep(time.Millisecond)
	userQueue.AddJob("1", "value 1")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("2", "value 2")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("3", "value 3")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("4", "value 4")
	time.Sleep(time.Second)
	userQueue.AddJob("5", "value 5")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("6", "value 6")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("7", "value 7")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("8", "value 8")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("9", "value 9")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("10", "value 10")
	time.Sleep(time.Millisecond)
	userQueue.AddJob("11", "value 11")

	userQueue.Process(func(job *Job) {
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
}

func HeaveTask(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	return string(bytes), err
}
