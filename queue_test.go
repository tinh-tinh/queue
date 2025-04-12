package queue_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
	"golang.org/x/crypto/bcrypt"
)

func Test_Queue(t *testing.T) {
	addr := "localhost:6379"
	userQueue := queue.New("user", &queue.Options{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
		Limiter: &queue.RateLimiter{
			Max:      3,
			Duration: time.Second,
		},
	})

	userQueue.Process(func(job *queue.Job) {
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
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})

	userQueue.BulkAddJob([]queue.AddJobOptions{
		{
			Id:       "12",
			Data:     "value 12",
			Priority: 11,
		},
		{
			Id:       "13",
			Data:     "value 13",
			Priority: 12,
		},
		{
			Id:       "14",
			Data:     "value 14",
			Priority: 13,
		},
		{
			Id:       "15",
			Data:     "value 15",
			Priority: 14,
		},
		{
			Id:       "16",
			Data:     "value 16",
			Priority: 15,
		},
		{
			Id:       "17",
			Data:     "value 17",
			Priority: 16,
		},
		{
			Id:       "18",
			Data:     "value 18",
			Priority: 17,
		},
		{
			Id:       "19",
			Data:     "value 19",
			Priority: 18,
		},
		{
			Id:       "20",
			Data:     "value 20",
			Priority: 19,
		},
		{
			Id:       "21",
			Data:     "value 21",
			Priority: 20,
		},
	})

	userQueue.Pause()
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "2",
		Data: "value 2",
	})
	userQueue.Resume()
}

func Test_SchedulerQueue(t *testing.T) {
	addr := "localhost:6379"
	userQueue := queue.New("user_schedule", &queue.Options{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       6,
		RetryFailures: 3,
		Pattern:       "@every 0h0m1s",
	})

	userQueue.Process(func(job *queue.Job) {
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

	userQueue.AddJob(queue.AddJobOptions{
		Id:       "1",
		Data:     "value 1",
		Priority: 1,
	})
	time.Sleep(5 * time.Second)
}

func HeaveTask(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	return string(bytes), err
}

func Test_Crash(t *testing.T) {
	addr := "localhost:6379"
	userQueue := queue.New("crash", &queue.Options{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
	})

	userQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			panic("error by test")
		})
	})

	t.Parallel()

	// t.Run("test", func(t *testing.T) {
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})
}

func TestDisableLog(t *testing.T) {
	addr := "localhost:6379"
	userQueue := queue.New("disabled", &queue.Options{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
		Logger:        queue.LoggerDisabled,
	})

	userQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			panic("error by test")
		})
	})

	t.Parallel()

	// t.Run("test", func(t *testing.T) {
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})
}

func Test_LoggerInfo(t *testing.T) {
	addr := "localhost:6379"
	userQueue := queue.New("info", &queue.Options{
		Connect: &redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
		Logger:        queue.LoggerInfo,
	})

	userQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			panic("error by test")
		})
	})

	t.Parallel()

	// t.Run("test", func(t *testing.T) {
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})
}
