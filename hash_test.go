package ioredis

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type User struct {
	Name  string
	Email string
}

func Test_Hash(t *testing.T) {
	addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	rdb := New(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	userHash := NewHash[User]("users", rdb).Expire(20 * time.Second)

	testcase := []User{
		{Name: "John", Email: "john@gmail.com"},
		{Name: "Kafka", Email: "kafka@gmail.com"},
		{Name: "Babadook", Email: "babadook@gmail.com"},
		{Name: "Mui", Email: "mui@gmail.com"},
		{Name: "Ren", Email: "ren@gmail.com"},
		{Name: "Griffith", Email: "griffith@gmail.com"},
		{Name: "Anne", Email: "anne@gmail.com"},
		{Name: "Karennine", Email: "karennnine@gmail.com"},
	}

	for i := 0; i < len(testcase); i++ {
		t.Run("Test_Upsert", func(t *testing.T) {
			key := fmt.Sprint(i + 1)
			err := userHash.Upsert(key, &testcase[i])
			if err != nil {
				t.Error(err)
			}
		})
	}

	t.Run("Test_FindMany", func(t *testing.T) {
		data, err := userHash.FindMany()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(data)
	})

	t.Run("Test_FindONe", func(t *testing.T) {
		data, err := userHash.FindByKey("1")
		if err != nil {
			t.Error(err)
		}
		fmt.Println(*data)
	})
}
