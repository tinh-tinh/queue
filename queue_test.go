package queue_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
	"golang.org/x/crypto/bcrypt"
)

func Test_Queue(t *testing.T) {
	userQueue := queue.New("user", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
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

	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})

	count := userQueue.CountJobs(queue.CompletedStatus)
	assert.Equal(t, 1, count)

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

	count = userQueue.CountJobs(queue.CompletedStatus)
	assert.Equal(t, 7, count)

	count = userQueue.CountJobs(queue.FailedStatus)
	assert.Equal(t, 4, count)

	userQueue.Pause()
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "2",
		Data: "value 2",
	})

	count = userQueue.CountJobs(queue.WaitStatus)
	assert.Equal(t, 1, count)

	userQueue.Resume()

	count = userQueue.CountJobs(queue.CompletedStatus)
	assert.Equal(t, 8, count)
}

func Test_SchedulerQueue(t *testing.T) {
	userQueue := queue.New("user_schedule", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       6,
		RetryFailures: 3,
		Pattern:       "@every 0h0m1s",
	})

	userQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
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
	userQueue := queue.New("crash", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
	})

	userQueue.Process(func(job *queue.Job) {
		if job.Id == "2" {
			panic("error by test")
		}
		job.Process(func() error {
			panic("error by test")
		})
	})

	require.NotPanics(t, func() {
		userQueue.AddJob(queue.AddJobOptions{
			Id:   "1",
			Data: "value 1",
		})
		userQueue.AddJob(queue.AddJobOptions{
			Id:   "2",
			Data: "value 2",
		})
	})
}

func TestDisableLog(t *testing.T) {
	userQueue := queue.New("disabled", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
		DisableLog:    true,
	})

	userQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			panic("error by test")
		})
	})

	// t.Run("test", func(t *testing.T) {
	userQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})
}

func Test_LoggerInfo(t *testing.T) {
	userQueue := queue.New("info", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
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

func Test_CountJob(t *testing.T) {
	postQueue := queue.New("post", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
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

	postQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			key, err := json.Marshal(job.Data)
			require.Nil(t, err)
			_, err = HeaveTask(string(key))
			require.Nil(t, err)
			return nil
		})
	})

	postQueue.Pause()

	postQueue.AddJob(queue.AddJobOptions{
		Id:   "1",
		Data: "value 1",
	})
	postQueue.AddJob(queue.AddJobOptions{
		Id:   "2",
		Data: "value 2",
	})
	postQueue.AddJob(queue.AddJobOptions{
		Id:   "3",
		Data: "value 3",
	})
	postQueue.AddJob(queue.AddJobOptions{
		Id:   "4",
		Data: "value 4",
	})
	postQueue.AddJob(queue.AddJobOptions{
		Id:   "5",
		Data: "value 5",
	})
	postQueue.AddJob(queue.AddJobOptions{
		Id:   "6",
		Data: "value 6",
	})

	count := postQueue.CountJobs(queue.WaitStatus)
	assert.Equal(t, 3, count)

	count = postQueue.CountJobs(queue.DelayedStatus)
	assert.Equal(t, 3, count)

	postQueue.Resume()

	count = postQueue.CountJobs(queue.CompletedStatus)
	assert.Equal(t, 6, count)
}

func Test_RemoveOnComplateAndFail(t *testing.T) {
	docQueue := queue.New("document", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RemoveOnComplete: true,
		RemoveOnFail:     true,
	})

	docQueue.Process(func(job *queue.Job) {
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

	docQueue.BulkAddJob([]queue.AddJobOptions{
		{
			Id:   "2",
			Data: "value 2",
		},
		{
			Id:   "3",
			Data: "value 3",
		},
		{
			Id:   "4",
			Data: "value 4",
		},
		{
			Id:   "5",
			Data: "value 5",
		},
		{
			Id:   "6",
			Data: "value 6",
		},
		{
			Id:   "7",
			Data: "value 7",
		},
		{
			Id:   "8",
			Data: "value 8",
		},
		{
			Id:   "9",
			Data: "value 9",
		},
		{
			Id:   "10",
			Data: "value 20",
		},
		{
			Id:   "1",
			Data: "value 21",
		},
	})

	count := docQueue.CountJobs(queue.CompletedStatus)
	assert.Equal(t, 0, count)

	count = docQueue.CountJobs(queue.FailedStatus)
	assert.Equal(t, 0, count)
}

func Test_Delay(t *testing.T) {
	userDelayQueue := queue.New("user_delay", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0,
		Delay:         3 * time.Second,
	})

	userDelayQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			key, err := json.Marshal(job.Data)
			require.Nil(t, err)

			if job.Id == "1" {
				return errors.New("ac")
			}
			_, err = HeaveTask(string(key))
			require.Nil(t, err)

			return nil
		})
	})

	userDelayQueue.AddJob(queue.AddJobOptions{
		Id:       "1",
		Data:     "value 1",
		Priority: 1,
	})

	userDelayQueue.AddJob(queue.AddJobOptions{
		Id:       "2",
		Data:     "value 2",
		Priority: 1,
	})
}

func Test_Timeout(t *testing.T) {
	userTimeoutQueue := queue.New("user_timeout", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 3,
		Timeout:       1 * time.Second,
	})

	userTimeoutQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			key, err := json.Marshal(job.Data)
			require.Nil(t, err)

			if job.Id == "1" {
				time.Sleep(2 * time.Second)
			}
			_, err = HeaveTask(string(key))
			require.Nil(t, err)

			return nil
		})
	})

	userTimeoutQueue.AddJob(queue.AddJobOptions{
		Id:       "1",
		Data:     "value 1",
		Priority: 1,
	})

	failedJob := userTimeoutQueue.CountJobs(queue.FailedStatus)
	require.Equal(t, 1, failedJob)
}

func TestPanic(t *testing.T) {
	require.Panics(t, func() {
		abc := queue.New("Abc", nil)
		abc.Pause()
	})
}
