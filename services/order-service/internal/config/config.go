package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppPort             string
	PostgresDSN         string
	RabbitMQURL         string
	RabbitMQQueue       string
	RabbitMQExchange    string
	InventoryServiceURL string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, falling back to system environment variables")
	}

	cfg := &Config{
		AppPort:             getEnv("APP_PORT", "8081"),
		PostgresDSN:         getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5433/orders?sslmode=disable"),
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQQueue:       getEnv("RABBITMQ_ORDER_QUEUE", "order.created"),
		RabbitMQExchange:    getEnv("RABBITMQ_EXCHANGE", ""),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8082"),
	}

	log.Info().
		Str("app_port", cfg.AppPort).
		Str("postgres_dsn", cfg.PostgresDSN).
		Str("rabbitmq_url", cfg.RabbitMQURL).
		Str("rabbitmq_queue", cfg.RabbitMQQueue).
		Str("rabbitmq_exchange", cfg.RabbitMQExchange).
		Str("inventory_service_url", cfg.InventoryServiceURL).
		Msg("Loaded configuration")

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Warn().Str("key", key).Msg("Using fallback for missing environment variable")
	return fallback
}
