package limiter

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter gerencia as configurações de limitação de requisições.
type RateLimiter struct {
	Storage   StorageStrategy
	LimitIP   int
	LimitToken int
	BlockTime time.Duration
}

// NewRateLimiter cria uma nova instância do RateLimiter.
func NewRateLimiter(storage StorageStrategy, limitIP, limitToken int, blockTime time.Duration) *RateLimiter {
	return &RateLimiter{
		Storage:   storage,
		LimitIP:   limitIP,
		LimitToken: limitToken,
		BlockTime: blockTime,
	}
}

// AllowRequest verifica se a requisição é permitida.
func (r *RateLimiter) AllowRequest(ctx context.Context, identifier string, isToken bool) (bool, error) {
	key := fmt.Sprintf("limiter:%s", identifier)

	blocked, err := r.Storage.IsBlocked(ctx, key)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}

	limit := r.LimitIP
	if isToken {
		limit = r.LimitToken
	}

	count, err := r.Storage.IncrementKey(ctx, key, time.Second)
	if err != nil {
		return false, err
	}

	if count > limit {
		err := r.Storage.BlockKey(ctx, key, r.BlockTime)
		return false, err
	}

	return true, nil
}
