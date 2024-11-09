package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type User struct {
	Name  string
	Email string
}

func Test_Hash(t *testing.T) {
	addr := "localhost:6379"
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
		key := fmt.Sprint(i + 1)
		err := userHash.Upsert(key, &testcase[i])
		require.Nil(t, err)
	}

	_, err := userHash.FindMany()
	require.Nil(t, err)

	_, err = userHash.FindByKey("1")
	require.Nil(t, err)

	err = userHash.Delete("1")
	require.Nil(t, err)
}
