package ioredis

import "encoding/json"

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

func (h *Hash[M]) Upsert(key string, data *M) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return h.redis.client.HSet(h.redis.ctx, h.Name, key, string(val)).Err()
}

func (h *Hash[M]) FindMany() ([]M, error) {
	var data []M
	val, err := h.redis.client.HGet(h.redis.ctx, h.Name, "*").Result()
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(val), &data)
	return data, nil
}

func (h *Hash[M]) FindByKey(key string) (*M, error) {
	var data M
	val, err := h.redis.client.HGet(h.redis.ctx, h.Name, key).Result()
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(val), &data)
	return &data, nil
}

func (h *Hash[M]) Delete(key string) error {
	return h.redis.client.HDel(h.redis.ctx, h.Name, key).Err()
}
