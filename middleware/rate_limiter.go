package middleware

import (
	"context"
	"net/http"

	"rate_limiter/limiter"
)

// RateLimiterMiddleware cria um middleware para limitar requisições.
func RateLimiterMiddleware(rateLimiter *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			token := r.Header.Get("API_KEY")
			identifier := r.RemoteAddr

			if token != "" {
				identifier = token
			}

			allowed, err := rateLimiter.AllowRequest(ctx, identifier, token != "")
			if err != nil || !allowed {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
