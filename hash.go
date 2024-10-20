package ioredis

import (
	"encoding/json"
	"fmt"
	"time"
)

type Hash[M any] struct {
	Name  string
	redis *Redis
}

func NewHash[M any](name string, r *Redis) *Hash[M] {
	return &Hash[M]{
		Name:  name,
		redis: r,
	}
}

func (h *Hash[M]) Expire(time time.Duration) *Hash[M] {
	result := h.redis.GetClient().Expire(h.redis.GetCtx(), h.Name, time)
	fmt.Println(result)
	return h
}

func (h *Hash[M]) Upsert(key string, data *M) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return h.redis.GetClient().HSet(h.redis.GetCtx(), h.Name, key, string(val)).Err()
}

func (h *Hash[M]) FindMany() ([]*M, error) {
	var data []*M
	mapper, err := h.redis.GetClient().HGetAll(h.redis.GetCtx(), h.Name).Result()
	if err != nil {
		return nil, err
	}

	for _, v := range mapper {
		var item *M
		err := json.Unmarshal([]byte(v), &item)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return data, nil
}

func (h *Hash[M]) FindByKey(key string) (*M, error) {
	var data M
	val, err := h.redis.GetClient().HGet(h.redis.GetCtx(), h.Name, key).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (h *Hash[M]) Delete(key string) error {
	return h.redis.GetClient().HDel(h.redis.GetCtx(), h.Name, key).Err()
}
