package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	Client *redis.Client
}

func (s *RedisStore) Set(ctx context.Context, key string, value any, expiration time.Duration) (err error) {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Client.Set(ctx, dbPrefix+key, data, expiration).Err()
}

func (s *RedisStore) Get(ctx context.Context, key string, value any) (bool, error) {
	cmd := s.Client.Get(ctx, dbPrefix+key)
	if cmd.Err() != nil {
		// TODO Test this
		switch {
		case strings.Contains(cmd.Err().Error(), "key not found"):
			fallthrough
		case strings.Contains(cmd.Err().Error(), "nil"):
			return false, nil
		default:
			slog.Warn("Unhandled error encountered", "err", cmd.Err().Error())
			return false, cmd.Err()
		}
	}

	data, err := cmd.Bytes()
	if err != nil {
		return false, err
	}

	return true, json.Unmarshal(data, value)
}
