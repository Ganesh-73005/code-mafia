package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MongoDBURI      string
	MongoDBDatabase string
	RedisAddress    string
	RedisPassword   string
	SecretKey       string
	JudgeZeroAPI    string
	XAuthToken      string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		MongoDBURI:      getEnv("MONGODB_URI", ""),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "code_mafia"),
		RedisAddress:    getEnv("REDIS_ADDRESS", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		SecretKey:       getEnv("SECRET_KEY", ""),
		JudgeZeroAPI:    getEnv("JUDGE_ZERO_API", "http://localhost:2358"),
		XAuthToken:      getEnv("X_AUTH_TOKEN", ""),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
