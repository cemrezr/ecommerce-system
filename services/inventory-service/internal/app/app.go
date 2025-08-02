package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cemrezr/ecommerce-system/inventory-service/internal/config"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/event"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/repository"
	"github.com/cemrezr/ecommerce-system/pkg/database"
	"github.com/cemrezr/ecommerce-system/pkg/rabbitmq"
	"github.com/rs/zerolog"
)

func Run(cfg *config.Config, log zerolog.Logger) {
	// PostgreSQL
	db := database.Connect(cfg.DBDSN, log)
	defer db.Close()

	inventoryRepo := repository.NewPostgresInventoryRepository(db)
	productRepo := repository.NewPostgresProductRepository(db)

	// RabbitMQ
	conn, ch, err := rabbitmq.Connect(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()
	defer ch.Close()

	// Setup queues for inventory
	if err := rabbitmq.SetupBasicQueue(ch, cfg.RabbitMQExchange, cfg.RabbitMQQueue,
		[]string{"order.created", "order.cancelled"}, log); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup inventory queue")
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		consumer := event.NewConsumer(ch, cfg.RabbitMQQueue, inventoryRepo, log)
		if err := consumer.StartConsuming(ctx); err != nil {
			log.Fatal().Err(err).Msg("Inventory consumer failed")
		}
	}()

	// Start HTTP
	startHTTPServer(cfg.AppPort, productRepo, log)

	// Wait for shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
	log.Info().Msg("🧹 Graceful shutdown complete")
}
