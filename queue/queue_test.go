package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func Test_Queue(t *testing.T) {
	addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	userQueue := New("user", &QueueOption{
		Workers:       6,
		RetryFailures: 3,
		Limiter: &RateLimiter{
			Max:      3,
			Duration: time.Millisecond,
		},
	}, rdb)

	fmt.Println(userQueue)
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

	t.Run("TestCase", func(t *testing.T) {
		userQueue.Process(func(job *Job) {
			job.Process(func() error {
				num, err := strconv.Atoi(job.Id)
				if err != nil {
					return err
				}
				if num%3 == 0 {
					return errors.New("error by test")
				}

				key, err := json.Marshal(job.Data)
				if err != nil {
					return err
				}
				_, err = HeaveTask(string(key))
				if err != nil {
					return err
				}
				return nil
			})
		})
	})
}

func Test_SchedulerQueue(t *testing.T) {
	addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	userQueue := New("user", &QueueOption{
		Workers:       6,
		RetryFailures: 3,
		Limiter: &RateLimiter{
			Max:      3,
			Duration: time.Second,
		},
		Pattern: "@every 0h0m1s",
	}, rdb)

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
	t.Run("TestCase", func(t *testing.T) {
		userQueue.Process(func(job *Job) {
			job.Process(func() error {
				num, err := strconv.Atoi(job.Id)
				if err != nil {
					return err
				}
				if num%3 == 0 {
					return errors.New("error by test")
				}

				key, err := json.Marshal(job.Data)
				if err != nil {
					return err
				}
				_, err = HeaveTask(string(key))
				if err != nil {
					return err
				}
				return nil
			})
		})
	})
}

func HeaveTask(key string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	return string(bytes), err
}
