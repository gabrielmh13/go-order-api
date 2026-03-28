package config

import (
	"os"

	"go-order-api/internal/domain/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MongoURI    string
	DatabaseName string
	RabbitMQURL string
}

var AppConfig *Config

func LoadConfig(l logger.Logger) {
	if err := godotenv.Load(); err != nil {
		l.Warning("Warning: .env file not found, using environmental variables.")
	}

	AppConfig = &Config{
		Port:         getEnv("PORT", "3333"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DB_NAME", "order-api"),
		RabbitMQURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}

	l.Info("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
