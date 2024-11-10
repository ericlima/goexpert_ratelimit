package limiter

import (
	"context"
	"time"
)

// StorageStrategy define os métodos necessários para persistência.
type StorageStrategy interface {
	IncrementKey(ctx context.Context, key string, expiration time.Duration) (int, error)
	BlockKey(ctx context.Context, key string, duration time.Duration) error
	IsBlocked(ctx context.Context, key string) (bool, error)
}
