package config

import (
	"os"
)

type Config struct {
	RabbitMQURL      string
	RabbitMQQueue    string
	RabbitMQExchange string
}

func Load() *Config {
	return &Config{
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQQueue:    getEnv("RABBITMQ_QUEUE", "notification.order.queue"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "order.events"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
