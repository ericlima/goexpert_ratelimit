package main

import (
	_ "context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	redisClient  	*redis.Client
	config         	Config
)

const (
	MESSAGE_BLOCK = "you have reached the maximum number of requests or actions allowed within a certain time frame"	
)

type Config struct {
	Strategy				string
	IpMaxRequests   		int
	IpBlockDuration 		int
	TokenKey				string
	TokenRequests			int
	TokenMaxBlockDuration	int
	RedisHost 				string
	RedisPort 				string
	ServerPort				string
	RequestCounts           map[string]int
}

func loadConfig() *Config {
	return &Config{
		Strategy: 				os.Getenv("STRATEGY"),
		IpMaxRequests:			readIntValueFromConfig("IP_MAX_REQUESTS"),
		IpBlockDuration:		readIntValueFromConfig("IP_BLOCK_DURATION"),
		TokenKey: 				os.Getenv("TOKEN_KEY"),
		TokenRequests:			readIntValueFromConfig("TOKEN_REQUESTS"),
		TokenMaxBlockDuration:	readIntValueFromConfig("TOKEN_MAX_BLOCK_DURATION"),
		RedisHost:     			os.Getenv("REDIS_HOST"),
		RedisPort:     			os.Getenv("REDIS_PORT"),
		ServerPort:				os.Getenv("SERVER_PORT"),	
		RequestCounts:          make(map[string]int),	
	}
}

func readIntValueFromConfig(key string) int {
	valueStr := os.Getenv(key)
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("Erro ao converter %s para int: %v", key, err)
	}
	return valueInt
}

func init() {
	// Le configs em .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	config = *loadConfig()
}

func main() {

	// Conectar ao Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort), // Endereço do Redis
		Password: "",               // Senha (se necessário)
		DB:       0,                // Banco de dados
	})

	// Verificar a conexão com o Redis
	_, err := redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		panic(err)
	}

	// Inicializar o roteador do Gorilla Mux
	r := mux.NewRouter()

	// Rota de exemplo
	r.HandleFunc("/api/example", exampleHandler).Methods("GET")

	// Middleware para controlar o número de requisições
	if config.Strategy == "REDIS" {
		r.Use(rateLimitMiddlewareRedis)
	} else {
		r.Use(rateLimitMiddlewareMemory)
	}
	
	// Iniciar o servidor
	http.ListenAndServe(fmt.Sprintf(":%s", config.ServerPort), r)
}

func exampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Exemplo de Rate Limiter"}`))
}

func rateLimitMiddlewareRedis(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr
		
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			// Lidar com o erro, se houver
			fmt.Println("Erro ao dividir o endereço:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Verificar se o IP está bloqueado
		blocked, err := redisClient.Get(redisClient.Context(), fmt.Sprintf("blocked:%s", ip)).Result()
		if err != nil && err != redis.Nil {
			fmt.Println("Erro ao verificar o bloqueio do IP:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if blocked == "true" {
			http.Error(w, MESSAGE_BLOCK , http.StatusTooManyRequests)
			return
		}

		// Definir o limite padrão de requisições por segundo
		limit := config.IpMaxRequests
		blockTime := config.IpBlockDuration

		// Verificar se o cabeçalho de autenticação é TOKEN_KEY
		if authHeader := r.Header.Get("API_KEY"); authHeader == config.TokenKey {
			// Aumentar o limite para 20 requisições por segundo se o cabeçalho de autenticação for "eureka"
			limit = config.TokenRequests
			blockTime = config.TokenMaxBlockDuration
		}

		// Obter ou inicializar o contador de requisições para este IP
		countKey := fmt.Sprintf("requests:%s", ip)
		_, err = redisClient.Get(redisClient.Context(), countKey).Result()
		if err != nil && err == redis.Nil {
			// Inicializar o contador para este IP
			err := redisClient.Set(redisClient.Context(), countKey, 0, 1*time.Second).Err()
			if err != nil {
				fmt.Println("Erro ao inicializar contador:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Incrementar o contador de requisições para este IP
		_, err = redisClient.Incr(redisClient.Context(), countKey).Result()
		if err != nil {
			fmt.Println("Erro ao incrementar contador:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Obter o número de requisições feitas por este IP nas últimas 10 segundos
		count, err := redisClient.Get(redisClient.Context(), countKey).Int()
		if err != nil {
			fmt.Println("Erro ao obter contador:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if count > limit {
			// Bloquear o IP por X segundos
			err := redisClient.Set(redisClient.Context(), fmt.Sprintf("blocked:%s", ip), "true", time.Duration(blockTime)*time.Second).Err()
			if err != nil {
				fmt.Println("Erro ao bloquear IP:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.Error(w, MESSAGE_BLOCK, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddlewareMemory(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr

		ip, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			// Lidar com o erro, se houver
			fmt.Println("Erro ao dividir o endereço:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Verificar se o IP está bloqueado
		blocked := false
		if count, ok := config.RequestCounts[ip]; ok && count > 10 {
			blocked = true
		}

		if blocked {
			http.Error(w, MESSAGE_BLOCK, http.StatusTooManyRequests)
			time.AfterFunc(10*time.Second, func() {
				delete(config.RequestCounts, ip)
			})
			return
		}

		// Definir o limite padrão de requisições por segundo
		limit := config.IpMaxRequests
		blockTime := config.IpBlockDuration

		// Verificar se o cabeçalho de autenticação é "eureka"
		if authHeader := r.Header.Get("API_KEY"); authHeader == config.TokenKey {
			// Aumentar o limite para 20 requisições por segundo se o cabeçalho de autenticação for "eureka"
			limit = config.TokenRequests
			blockTime = config.TokenMaxBlockDuration
		}

		// Incrementar o contador de requisições para este IP
		config.RequestCounts[ip]++

		// Verificar se o limite foi excedido
		if config.RequestCounts[ip] > limit {
			// Bloquear o IP por 10 segundos
			time.AfterFunc(time.Duration(blockTime)*time.Second, func() {
				delete(config.RequestCounts, ip)
			})

			http.Error(w, MESSAGE_BLOCK, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

