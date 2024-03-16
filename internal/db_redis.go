package internal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisStore struct {
	client *redis.Client
}

func (s *redisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (err error) {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, dbPrefix+key, data, expiration).Err()
}

func (s *redisStore) Get(ctx context.Context, key string, value interface{}) error {
	cmd := s.client.Get(ctx, dbPrefix+key)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	data, err := cmd.Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}
