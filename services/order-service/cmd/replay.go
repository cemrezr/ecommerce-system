package main

import (
	"time"

	"github.com/cemrezr/ecommerce-system/order-service/internal/config"
	"github.com/cemrezr/ecommerce-system/order-service/internal/event"
	"github.com/cemrezr/ecommerce-system/order-service/internal/repository"
	"github.com/cemrezr/ecommerce-system/pkg/database"
	"github.com/cemrezr/ecommerce-system/pkg/logger"
	"github.com/cemrezr/ecommerce-system/pkg/rabbitmq"

	"github.com/sony/gobreaker"
)

func main() {
	log := logger.NewLogger("order-replayer")
	log.Info().Msg("Starting order-replayer")

	cfg := config.LoadConfig()

	db := database.Connect(cfg.PostgresDSN, log)
	defer db.Close()

	eventLogger := repository.NewEventLogRepository(db)

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "replay-circuit",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})

	conn, ch, err := rabbitmq.Connect(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()
	defer ch.Close()

	if err := rabbitmq.SetupOrderQueues(ch, cfg.RabbitMQExchange, cfg.RabbitMQQueue, log); err != nil {
		log.Fatal().Err(err).Msg("Failed to declare queues")
	}

	publisher := event.NewPublisher(ch, cfg.RabbitMQExchange, cb, eventLogger, log)
	replayer := event.NewReplayer(eventLogger, publisher, cb, log)

	if err := replayer.ReplayFailedEvents(); err != nil {
		log.Fatal().Err(err).Msg("Replay process failed")
	}

	log.Info().Msg("Replay finished")
}
