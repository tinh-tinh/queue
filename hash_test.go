package ioredis

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type User struct {
	Name  string
	Email string
}

func Test_Hash(t *testing.T) {
	rdb := New(&redis.Options{
		Addr:     "localhost:6379",
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

	time.Sleep(20 * time.Second)
	t.Run("TesT_Expire", func(t *testing.T) {
		data, err := userHash.FindMany()
		if err != nil {
			t.Error(err)
		}
		if len(data) != 0 {
			fmt.Println(data)
			t.Error("data should out of redis")
		}
	})
}
