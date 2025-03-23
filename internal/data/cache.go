package data

import (
	"context"
	"errors"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (err error)
	Get(ctx context.Context, key string, value interface{}) error
}

const (
	dbPrefix = "no-time-to-explain:"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)
