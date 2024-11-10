package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadConfig carrega as variáveis de ambiente do arquivo .env.
func LoadConfig() map[string]string {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Não foi possível carregar o arquivo .env")
	}

	return map[string]string{
		"REDIS_ADDR":      getEnv("REDIS_ADDR", "localhost:6379"),
		"REDIS_PASSWORD":  getEnv("REDIS_PASSWORD", ""),
		"RATE_LIMIT_IP":   getEnv("RATE_LIMIT_IP", "5"),
		"RATE_LIMIT_TOKEN": getEnv("RATE_LIMIT_TOKEN", "10"),
		"BLOCK_DURATION":  getEnv("BLOCK_DURATION", "300"),
		"USE_MEMORY":      getEnv("USE_MEMORY", "false"),
	}
}

// getEnv retorna o valor de uma variável de ambiente ou o valor padrão.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetInt retorna o valor de uma variável de ambiente como inteiro.
func GetInt(config map[string]string, key string) int {
	value, err := strconv.Atoi(config[key])
	if err != nil {
		log.Fatalf("Erro ao converter %s para inteiro: %v", key, err)
	}
	return value
}
