package data

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) (err error)
	Get(ctx context.Context, key string, value any) (found bool, err error)
}

const (
	dbPrefix = "no-time-to-explain:"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

func NewCache() Cache {
	host, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		// Default to a memory backend
		slog.Warn("Redis persistence disabled")
		return &MemoryStore{Data: map[string][]byte{}}
	}

	// Determine the address, whether it be HOST:PORT or HOST & PORT
	var port string
	if host, port, ok = strings.Cut(host, ":"); !ok {
		if port, ok = os.LookupEnv("REDIS_PORT"); !ok {
			port = "4646"
		}
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	opts := &redis.Options{Addr: addr}

	// Determine if a username & password is set
	if user, ok := os.LookupEnv("REDIS_USER"); ok {
		opts.TLSConfig = &tls.Config{}
		opts.Username = user
	}
	if pass, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		opts.TLSConfig = &tls.Config{}
		opts.Password = pass
	}

	// Set the singleton value to the Redis backend
	cache := &RedisStore{Client: redis.NewClient(opts)}
	if err := cache.Set(context.Background(), "client", "no-time-to-explain", time.Hour*24); err != nil {
		log.Fatalf("Unable to connect to Redis backend at %s: %s", addr, err)
	}
	slog.Info("Redis persistence configured", "addr", addr)
	return cache
}
