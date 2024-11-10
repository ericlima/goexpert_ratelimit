package tests

import (
	"context"
	"testing"
	"time"

	"rate_limiter/limiter"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	redisClient := limiter.NewRedisClient("localhost:6379", "")
	rateLimiter := limiter.NewRateLimiter(redisClient, 5, 10, 5*time.Second)

	ctx := context.Background()
	ip := "192.168.0.1"

	for i := 0; i < 5; i++ {
		allowed, err := rateLimiter.AllowRequest(ctx, ip, false)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	// A sexta requisição deve ser bloqueada
	allowed, err := rateLimiter.AllowRequest(ctx, ip, false)
	assert.NoError(t, err)
	assert.False(t, allowed)
}
