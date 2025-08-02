package event

import (
	"context"
	"encoding/json"

	"github.com/cemrezr/ecommerce-system/inventory-service/internal/model"
	"github.com/cemrezr/ecommerce-system/inventory-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Consumer struct {
	ch    *amqp.Channel
	queue string
	repo  repository.InventoryRepository
	log   zerolog.Logger
}

func NewConsumer(ch *amqp.Channel, queue string, repo repository.InventoryRepository, log zerolog.Logger) *Consumer {
	return &Consumer{ch: ch, queue: queue, repo: repo, log: log}
}

func (c *Consumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.ch.Consume(
		c.queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to start consuming from queue")
		return err
	}

	c.log.Info().Str("queue", c.queue).Msg("Consumer started")

	go func() {
		for msg := range msgs {
			c.log.Debug().Str("type", msg.Type).Msg("Received message")

			switch msg.Type {

			case "order.created":
				var order model.Order
				if err := json.Unmarshal(msg.Body, &order); err != nil {
					c.log.Error().Err(err).Msg("Failed to parse order.created payload")
					continue
				}

				c.log.Info().
					Str("event", msg.Type).
					Int64("product_id", order.ProductID).
					Int("quantity", order.Quantity).
					Msg("Decreasing stock for order.created")

				if err := c.repo.DecreaseStock(ctx, order.ProductID, order.Quantity); err != nil {
					c.log.Error().Err(err).Msg("Failed to decrease stock")
					continue
				}

				if err := c.repo.LogStockChange(ctx, order.ProductID, -order.Quantity, "order.created"); err != nil {
					c.log.Warn().Err(err).Msg("⚠Failed to log stock change for order.created")
				}

				c.log.Info().
					Int64("product_id", order.ProductID).
					Int("quantity", order.Quantity).
					Msg("Stock decreased successfully")

			case "order.cancelled":
				var payload struct {
					OrderID   int64 `json:"order_id"`
					ProductID int64 `json:"product_id"`
					Quantity  int   `json:"quantity"`
				}

				if err := json.Unmarshal(msg.Body, &payload); err != nil {
					c.log.Error().Err(err).Msg("Failed to parse order.cancelled payload")
					continue
				}

				c.log.Info().
					Str("event", msg.Type).
					Int64("order_id", payload.OrderID).
					Int64("product_id", payload.ProductID).
					Int("quantity", payload.Quantity).
					Msg("Restoring stock for cancelled order")

				if err := c.repo.IncreaseStock(ctx, payload.ProductID, payload.Quantity); err != nil {
					c.log.Error().Err(err).Msg("Failed to increase stock")
					continue
				}

				if err := c.repo.LogStockChange(ctx, payload.ProductID, payload.Quantity, "order.cancelled"); err != nil {
					c.log.Warn().Err(err).Msg("Failed to log stock change for order.cancelled")
				}

				c.log.Info().
					Int64("product_id", payload.ProductID).
					Int("quantity", payload.Quantity).
					Msg("Stock restored successfully")

			default:
				c.log.Warn().Str("type", msg.Type).Msg("Unknown event type received")
			}
		}
	}()

	<-ctx.Done()
	c.log.Info().Msg("Consumer shutdown initiated")
	return nil
}
