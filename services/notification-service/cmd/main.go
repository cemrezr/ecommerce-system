package main

import (
	"context"

	"github.com/cemrezr/ecommerce-system/notification-service/internal/config"
	"github.com/cemrezr/ecommerce-system/notification-service/internal/event"
	"github.com/cemrezr/ecommerce-system/pkg/logger"
	"github.com/cemrezr/ecommerce-system/pkg/rabbitmq"
)

func main() {
	cfg := config.Load()
	log := logger.NewLogger("notification-service")

	log.Info().Msg("Starting notification-service")

	conn, ch, err := rabbitmq.Connect(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("RabbitMQ connection failed")
	}
	defer conn.Close()
	defer ch.Close()

	if err := rabbitmq.SetupBasicQueue(ch, cfg.RabbitMQExchange, cfg.RabbitMQQueue, []string{
		"order.created",
		"order.cancelled",
	}, log); err != nil {
		log.Fatal().Err(err).Msg("Failed to declare queue and bindings")
	}

	dispatcher := event.NewDispatcher(log)
	consumer := event.NewConsumer(ch, cfg.RabbitMQQueue, log, dispatcher)

	if err := consumer.StartConsuming(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Consumer startup failed")
	}
}
