package event

import (
	"context"
	"encoding/json"

	"github.com/cemrezr/ecommerce-system/order-service/internal/model"
	"github.com/cemrezr/ecommerce-system/order-service/internal/repository"
	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"
)

type Replayer struct {
	eventLogger repository.EventLogger
	publisher   *Publisher
	breaker     *gobreaker.CircuitBreaker
	log         zerolog.Logger
}

func NewReplayer(
	logger repository.EventLogger,
	pub *Publisher,
	cb *gobreaker.CircuitBreaker,
	log zerolog.Logger,
) *Replayer {
	return &Replayer{
		eventLogger: logger,
		publisher:   pub,
		breaker:     cb,
		log:         log,
	}
}

func (r *Replayer) ReplayFailedEvents() error {
	ctx := context.Background()

	events, err := r.eventLogger.ListFailed(ctx)
	if err != nil {
		r.log.Error().Err(err).Msg("Failed to list failed events")
		return err
	}

	r.log.Info().Int("count", len(events)).Msg("🔁 Starting event replay")

	for _, e := range events {
		r.log.Info().
			Int64("id", e.ID).
			Str("type", e.EventType).
			Interface("order_id", e.OrderID).
			Msg("🔄 Replaying event")

		var order model.Order
		if err := json.Unmarshal([]byte(e.Payload), &order); err != nil {
			r.log.Error().Err(err).Int64("id", e.ID).Msg("Invalid payload, skipping")
			continue
		}

		retryCount := e.RetryCount
		err := r.publisher.RepublishEvent(&order, e.ID, &retryCount)

		status := "published"
		if err != nil {
			status = "failed"
			r.log.Error().Err(err).Int64("id", e.ID).Msg("Replay failed")
		} else {
			r.log.Info().Int64("id", e.ID).Msg("Replay successful")
		}

		if updateErr := r.eventLogger.UpdateStatus(ctx, e.ID, status, retryCount); updateErr != nil {
			r.log.Error().Err(updateErr).Int64("id", e.ID).Msg("Failed to update event status")
		}
	}

	return nil
}
