package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go-order-api/internal/domain/logger"
)

type Config struct {
	Port          string
	MongoURI      string
	DatabaseName  string
	RabbitMQURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	OrderCacheTTL time.Duration
}

var AppConfig *Config

func LoadConfig(l logger.Logger) {
	if err := godotenv.Load(); err != nil {
		l.Warning("Warning: .env file not found, using environmental variables.")
	}

	AppConfig = &Config{
		Port:          getEnv("PORT", "3333"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:  getEnv("DB_NAME", "order-api"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0, l),
		OrderCacheTTL: getEnvAsDuration("ORDER_CACHE_TTL", 5*time.Minute, l),
	}

	l.Info("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int, l logger.Logger) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		l.Warning("Invalid integer env var, using default", logger.Any("key", key), logger.Any("value", value), logger.Any("default", defaultValue))
		return defaultValue
	}

	return parsed
}

func getEnvAsDuration(key string, defaultValue time.Duration, l logger.Logger) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		l.Warning("Invalid duration env var, using default", logger.Any("key", key), logger.Any("value", value), logger.Any("default", defaultValue.String()))
		return defaultValue
	}

	return parsed
}
