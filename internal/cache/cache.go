package cache

import (
	"context"
	"time"
)

var ErrKeyNotFound = "key not found"

type Cache[T any] interface {
	Set(ctx context.Context, key string, value T) error
	SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error
	Get(ctx context.Context, key string) (T, error)
}
