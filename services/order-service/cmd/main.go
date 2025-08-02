package main

import (
	"net/http"
	"time"

	"github.com/cemrezr/ecommerce-system/order-service/internal/client"
	"github.com/cemrezr/ecommerce-system/order-service/internal/config"
	"github.com/cemrezr/ecommerce-system/order-service/internal/event"
	"github.com/cemrezr/ecommerce-system/order-service/internal/handler"
	"github.com/cemrezr/ecommerce-system/order-service/internal/repository"
	"github.com/cemrezr/ecommerce-system/pkg/logger"
	"github.com/cemrezr/ecommerce-system/pkg/rabbitmq"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
	"github.com/streadway/amqp"
)

func main() {
	log := logger.New("order-service")
	cfg := config.LoadConfig()

	db := setupDatabase(cfg, log)
	defer db.Close()

	conn, ch := setupRabbitMQ(cfg, log)
	defer conn.Close()
	defer ch.Close()

	orderRepo := repository.NewOrderRepository(db)
	eventLogger := repository.NewEventLogRepository(db)

	breaker := setupCircuitBreaker()
	publisher := event.NewPublisher(ch, cfg.RabbitMQExchange, breaker, eventLogger, log)

	invClient := client.NewInventoryClient(cfg.InventoryServiceURL, log)

	router := setupRouter(orderRepo, publisher, invClient, log)

	log.Info().Str("addr", ":"+cfg.AppPort).Msg("Starting HTTP server")
	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func setupDatabase(cfg *config.Config, log zerolog.Logger) *sqlx.DB {
	db := repository.ConnectPostgres(cfg.PostgresDSN)
	log.Info().Str("component", "postgres").Msg("Connected to PostgreSQL")
	return db
}

func setupRabbitMQ(cfg *config.Config, log zerolog.Logger) (*amqp.Connection, *amqp.Channel) {
	conn, ch, err := rabbitmq.Connect(cfg.RabbitMQURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("RabbitMQ setup failed")
	}
	if err := rabbitmq.SetupOrderQueues(ch, cfg.RabbitMQExchange, cfg.RabbitMQQueue, log); err != nil {
		log.Fatal().Err(err).Msg("Queue setup failed")
	}
	return conn, ch
}

func setupCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "rabbitmq-publisher",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})
}

func setupRouter(orderRepo repository.OrderRepository, publisher *event.Publisher, invClient *client.InventoryClient, log zerolog.Logger) *mux.Router {
	handler := handler.NewOrderHandler(orderRepo, publisher, invClient, log)

	router := mux.NewRouter()
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("order-service OK"))
	}).Methods("GET")

	router.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{order_id}/cancel", handler.CancelOrder).Methods("POST")

	return router
}
