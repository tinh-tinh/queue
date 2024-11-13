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
		Workers:       3,
		RetryFailures: 3,
		Limiter: &RateLimiter{
			Max:      3,
			Duration: time.Second,
		},
	})

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

	t.Parallel()

	// t.Run("test", func(t *testing.T) {
	userQueue.AddJob("1", "value 1")
	userQueue.AddJob("2", "value 2")
	userQueue.AddJob("3", "value 3")
	userQueue.AddJob("4", "value 4")
	time.Sleep(time.Second)
	userQueue.AddJob("5", "value 5")
	userQueue.AddJob("6", "value 6")
	userQueue.AddJob("7", "value 7")
	userQueue.AddJob("8", "value 8")
	userQueue.AddJob("9", "value 9")
	userQueue.AddJob("10", "value 10")
	userQueue.AddJob("11", "value 11")
	// })

	userQueue.BulkAddJob([]AddJobOptions{
		{
			Id:   "12",
			Data: "value 12",
		},
		{
			Id:   "13",
			Data: "value 13",
		},
		{
			Id:   "14",
			Data: "value 14",
		},
		{
			Id:   "15",
			Data: "value 15",
		},
		{
			Id:   "16",
			Data: "value 16",
		},
		{
			Id:   "17",
			Data: "value 17",
		},
		{
			Id:   "18",
			Data: "value 18",
		},
		{
			Id:   "19",
			Data: "value 19",
		},
		{
			Id:   "20",
			Data: "value 20",
		},
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

	t.Parallel()

	userQueue.AddJob("1", "value 1")
	userQueue.AddJob("2", "value 2")
	userQueue.AddJob("3", "value 3")
	userQueue.AddJob("4", "value 4")
	time.Sleep(time.Second)
	userQueue.AddJob("5", "value 5")
	userQueue.AddJob("6", "value 6")
	userQueue.AddJob("7", "value 7")
	userQueue.AddJob("8", "value 8")
	userQueue.AddJob("9", "value 9")
	userQueue.AddJob("10", "value 10")
	userQueue.AddJob("11", "value 11")
}

func HeaveTask(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	return string(bytes), err
}
