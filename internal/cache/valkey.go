package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

type valkeyCache[T any] struct {
	client valkey.Client
}

func New[T any](client valkey.Client) Cache[T] {
	return &valkeyCache[T]{client: client}
}

func (c *valkeyCache[T]) Set(ctx context.Context, key string, value T) error {
	err := c.client.Do(ctx, c.client.B().Set().Key(key).Value(valueToString(value)).Build()).Error()
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}

func (c *valkeyCache[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error {
	err := c.client.Do(ctx, c.client.B().Set().Key(key).Value(valueToString(value)).Ex(ttl).Build()).Error()
	if err != nil {
		return fmt.Errorf("failed to set value with TTL: %w", err)
	}

	return nil
}

func (c *valkeyCache[T]) Get(ctx context.Context, key string) (T, error) {
	var value T
	result, err := c.client.Do(ctx, c.client.B().Get().Key(key).Build()).AsBytes()
	if err != nil {
		return value, fmt.Errorf("failed to get value: %w", err)
	}

	if result == nil {
		return value, fmt.Errorf("key not found: %s", key)
	}

	value, err = stringToValue[T](string(result))
	if err != nil {
		return value, fmt.Errorf("failed to convert value: %w", err)
	}

	return value, nil
}

func valueToString[T any](value T) string {
	if str, ok := any(value).(string); ok {
		return str
	}

	return fmt.Sprintf("%v", value)
}

func stringToValue[T any](str string) (T, error) {
	i, err := strconv.Atoi(str)
	if err == nil {
		return any(i).(T), nil
	}

	return any(str).(T), nil
}
