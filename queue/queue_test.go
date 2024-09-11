package queue

import (
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func Test_Queue(t *testing.T) {
	userQueue := New("user", &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	t.Parallel()

	userQueue.AddJob("1", "value 1")
	userQueue.AddJob("2", "value 2")
	userQueue.AddJob("3", "value 3")
	userQueue.AddJob("4", "value 4")
	userQueue.AddJob("5", "value 5")
	userQueue.AddJob("6", "value 6")
	userQueue.AddJob("7", "value 7")
	userQueue.AddJob("8", "value 8")
	userQueue.AddJob("9", "value 9")
	userQueue.AddJob("10", "value 10")
	userQueue.AddJob("11", "value 11")

	t.Run("TestCase", func(t *testing.T) {
		userQueue.Process(func(job *Job) {
			job.Process(func() error {
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
