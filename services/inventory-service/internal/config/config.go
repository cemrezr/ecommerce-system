package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppPort          string
	DBDSN            string
	RabbitMQURL      string
	RabbitMQQueue    string
	RabbitMQExchange string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, falling back to system environment variables")
	}

	cfg := &Config{
		AppPort:          getEnv("APP_PORT", "8082"),
		DBDSN:            getEnv("DB_DSN", "postgres://user:pass@localhost:5433/inventory?sslmode=disable"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQQueue:    getEnv("RABBITMQ_ORDER_QUEUE", "inventory.order.queue"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "order.events"),
	}

	log.Info().
		Str("app_port", cfg.AppPort).
		Str("db_dsn", cfg.DBDSN).
		Str("rabbitmq_url", cfg.RabbitMQURL).
		Str("rabbitmq_queue", cfg.RabbitMQQueue).
		Str("rabbitmq_exchange", cfg.RabbitMQExchange).
		Msg("Loaded inventory-service config")

	return cfg
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	log.Warn().Str("key", key).Msg("Using fallback for missing environment variable")
	return fallback
}
