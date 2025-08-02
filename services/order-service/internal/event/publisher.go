package event

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cemrezr/ecommerce-system/order-service/internal/model"
	"github.com/cemrezr/ecommerce-system/order-service/internal/repository"
	"github.com/cemrezr/ecommerce-system/order-service/internal/utils"
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
	"github.com/streadway/amqp"
)

type Publisher struct {
	channel     *amqp.Channel
	exchange    string
	breaker     *gobreaker.CircuitBreaker
	eventLogger repository.EventLogger
	log         zerolog.Logger
}

func NewPublisher(
	ch *amqp.Channel,
	exchange string,
	breaker *gobreaker.CircuitBreaker,
	eventLogger repository.EventLogger,
	log zerolog.Logger,
) *Publisher {
	return &Publisher{
		channel:     ch,
		exchange:    exchange,
		breaker:     breaker,
		eventLogger: eventLogger,
		log:         log,
	}
}

func (p *Publisher) PublishOrderCreated(order *model.Order) error {
	payload, _ := json.Marshal(order)

	logEntry := &repository.EventLog{
		EventType:    "order.created",
		EventVersion: "v1",
		Payload:      string(payload),
		Status:       "publishing",
		RetryCount:   0,
		OrderID:      &order.ID,
	}

	ctx := context.Background()
	if err := p.eventLogger.Insert(ctx, logEntry); err != nil {
		p.log.Error().Err(err).Msg("Failed to insert event log (order.created)")
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
		Type:        "order.created",
	}

	return p.publishWithRetries(ctx, msg, logEntry)
}

func (p *Publisher) PublishOrderCancelled(order *model.Order) error {
	payloadMap := map[string]interface{}{
		"order_id":   order.ID,
		"product_id": order.ProductID,
		"quantity":   order.Quantity,
	}
	payload, _ := json.Marshal(payloadMap)

	logEntry := &repository.EventLog{
		EventType:    "order.cancelled",
		EventVersion: "v1",
		Payload:      string(payload),
		Status:       "publishing",
		RetryCount:   0,
		OrderID:      &order.ID,
	}

	ctx := context.Background()
	if err := p.eventLogger.Insert(ctx, logEntry); err != nil {
		p.log.Error().Err(err).Msg("Failed to insert event log (order.cancelled)")
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
		Type:        "order.cancelled",
	}

	return p.publishWithRetries(ctx, msg, logEntry)
}

func (p *Publisher) RepublishEvent(order *model.Order, logID int64, currentRetry *int) error {
	ctx := context.Background()
	payload, _ := json.Marshal(order)

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
		Type:        "order.created",
	}

	retryCount, err := utils.RetryWithBreaker(p.breaker, func() error {
		return p.channel.Publish(p.exchange, msg.Type, false, false, msg)
	})

	*currentRetry += retryCount
	if err := p.eventLogger.UpdateStatus(ctx, logID, "publishing", *currentRetry); err != nil {
		p.log.Error().Err(err).Int64("log_id", logID).Msg("Failed to update event log status after retry")
	}

	if err != nil {
		p.log.Error().Err(err).Msg("Republish failed, sending to DLQ")

		dlqErr := p.channel.Publish("order.dlx", "order.failed", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
			Type:        "order.failed",
		})
		if dlqErr != nil {
			p.log.Error().Err(dlqErr).Msg("Failed to publish to DLQ")
		}

		_ = p.eventLogger.UpdateStatus(ctx, logID, "failed", *currentRetry)
		return errors.New("replay failed and sent to DLQ")
	}

	p.log.Info().Str("event", msg.Type).Int("retry", *currentRetry).Msg("Event replay successful")
	return nil
}

func (p *Publisher) publishWithRetries(ctx context.Context, msg amqp.Publishing, logEntry *repository.EventLog) error {
	for i := 1; i <= 3; i++ {
		err := p.channel.Publish(p.exchange, msg.Type, false, false, msg)
		logEntry.RetryCount = i

		if err == nil {
			p.log.Info().Str("event", msg.Type).Int("retry", i).Msg("Published successfully")
			_ = p.eventLogger.UpdateStatus(ctx, logEntry.ID, "published", logEntry.RetryCount)
			return nil
		}

		p.log.Warn().Err(err).Int("retry", i).Msg("⏳ Publish failed, retrying...")
		_ = p.eventLogger.UpdateStatus(ctx, logEntry.ID, "publishing", logEntry.RetryCount)
		time.Sleep(2 * time.Second)
	}

	p.log.Error().Str("event", msg.Type).Msg("Event lost after retries")

	dlqErr := p.channel.Publish("order.dlx", "order.failed", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msg.Body,
		Type:        "order.failed",
	})
	if dlqErr != nil {
		p.log.Error().Err(dlqErr).Msg("Failed to publish to DLQ")
	}

	_ = p.eventLogger.UpdateStatus(ctx, logEntry.ID, "failed", logEntry.RetryCount)
	return errors.New("event lost after retries")
}
