package limiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient estrutura o cliente Redis, que implementa a StorageStrategy.
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient cria uma nova instância de cliente Redis.
func NewRedisClient(addr, password string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &RedisClient{Client: client}
}

// IncrementKey incrementa o contador para a chave especificada.
func (r *RedisClient) IncrementKey(ctx context.Context, key string, expiration time.Duration) (int, error) {
	count, err := r.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		r.Client.Expire(ctx, key, expiration)
	}

	return int(count), nil
}

// BlockKey bloqueia uma chave por um tempo específico.
func (r *RedisClient) BlockKey(ctx context.Context, key string, duration time.Duration) error {
	return r.Client.Set(ctx, key, "BLOCKED", duration).Err()
}

// IsBlocked verifica se a chave está bloqueada.
func (r *RedisClient) IsBlocked(ctx context.Context, key string) (bool, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	return val == "BLOCKED", err
}
