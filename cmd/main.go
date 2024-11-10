package main

import (
	"log"
	"net/http"
	"time"

	"rate_limiter/config"
	"rate_limiter/limiter"
	"rate_limiter/middleware"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	var storage limiter.StorageStrategy
	if cfg["USE_MEMORY"] == "true" {
		log.Println("Usando armazenamento em memória")
		storage = limiter.NewMemoryClient()
	} else {
		log.Println("Usando armazenamento Redis")
		storage = limiter.NewRedisClient(cfg["REDIS_ADDR"], cfg["REDIS_PASSWORD"])
	}

	rateLimiter := limiter.NewRateLimiter(
		storage,
		config.GetInt(cfg, "RATE_LIMIT_IP"),
		config.GetInt(cfg, "RATE_LIMIT_TOKEN"),
		time.Duration(config.GetInt(cfg, "BLOCK_DURATION"))*time.Second,
	)

	router := mux.NewRouter()
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bem-vindo ao serviço!"))
	})

	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", router)
}
