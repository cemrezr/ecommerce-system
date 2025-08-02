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
		false,
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
					_ = msg.Nack(false, false)
					continue
				}

				if c.repo.HasAlreadyProcessed(order.ID, order.ProductID, "order.created") {
					c.log.Warn().Int64("order_id", order.ID).Msg("💡 Duplicate order detected — skipping")
					_ = msg.Ack(false)
					continue
				}

				c.log.Info().
					Str("event", msg.Type).
					Int64("product_id", order.ProductID).
					Int("quantity", order.Quantity).
					Msg("Decreasing stock for order.created")

				if err := c.repo.DecreaseStock(ctx, order.ProductID, order.Quantity); err != nil {
					c.log.Error().Err(err).Msg("Failed to decrease stock — NACKing for retry")
					_ = msg.Nack(false, true)
					continue
				}

				if err := c.repo.LogStockChange(ctx, order.ProductID, -order.Quantity, "order.created", &order.ID); err != nil {
					c.log.Warn().Err(err).Msg("Failed to log stock change — proceeding anyway")
				}

				c.log.Info().
					Int64("product_id", order.ProductID).
					Int("quantity", order.Quantity).
					Msg("Stock decreased successfully")

				_ = msg.Ack(false)

			case "order.cancelled":
				var payload struct {
					OrderID   int64 `json:"id"`
					ProductID int64 `json:"product_id"`
					Quantity  int   `json:"quantity"`
				}

				if err := json.Unmarshal(msg.Body, &payload); err != nil {
					c.log.Error().Err(err).Msg("Failed to parse order.cancelled payload")
					_ = msg.Nack(false, false)
					continue
				}

				if !c.repo.HasOrderCreatedLog(payload.OrderID, payload.ProductID) {
					c.log.Warn().
						Int64("order_id", payload.OrderID).
						Int64("product_id", payload.ProductID).
						Msg("Cancelled event received without a matching order.created log — skipping")
					_ = msg.Ack(false)
					continue
				}

				if c.repo.HasAlreadyProcessed(payload.OrderID, payload.ProductID, "order.cancelled") {
					c.log.Warn().Int64("order_id", payload.OrderID).Msg("💡 Duplicate cancelled order detected — skipping")
					_ = msg.Ack(false)
					continue
				}

				c.log.Info().
					Str("event", msg.Type).
					Int64("order_id", payload.OrderID).
					Int64("product_id", payload.ProductID).
					Int("quantity", payload.Quantity).
					Msg("Restoring stock for cancelled order")

				if err := c.repo.IncreaseStock(ctx, payload.ProductID, payload.Quantity); err != nil {
					c.log.Error().Err(err).Msg("Failed to increase stock — NACKing for retry")
					_ = msg.Nack(false, true)
					continue
				}

				if err := c.repo.LogStockChange(ctx, payload.ProductID, payload.Quantity, "order.cancelled", &payload.OrderID); err != nil {
					c.log.Warn().Err(err).Msg("⚠ Failed to log stock change for order.cancelled")
				}

				c.log.Info().
					Int64("product_id", payload.ProductID).
					Int("quantity", payload.Quantity).
					Msg("Stock restored successfully")

				_ = msg.Ack(false)

			default:
				c.log.Warn().Str("type", msg.Type).Msg("Unknown event type received")
				_ = msg.Ack(false)
			}
		}
	}()

	<-ctx.Done()
	c.log.Info().Msg("Consumer shutdown initiated")
	return nil
}
