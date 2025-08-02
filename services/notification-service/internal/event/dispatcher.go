package event

import (
	"encoding/json"
	"fmt"

	"github.com/cemrezr/ecommerce-system/notification-service/internal/handler"
	"github.com/cemrezr/ecommerce-system/notification-service/internal/model"
	"github.com/rs/zerolog"
)

type Dispatcher struct {
	log     zerolog.Logger
	handler *handler.NotificationHandler
}

func NewDispatcher(log zerolog.Logger) *Dispatcher {
	return &Dispatcher{
		log:     log,
		handler: handler.NewNotificationHandler(log),
	}
}

func (d *Dispatcher) Dispatch(eventType string, body []byte) error {
	switch eventType {
	case "order.created":
		var event model.OrderCreatedEvent
		if err := json.Unmarshal(body, &event); err != nil {
			d.log.Error().
				Err(err).
				Str("event_type", eventType).
				Msg("Failed to unmarshal order.created event")
			return fmt.Errorf("failed to unmarshal order.created: %w", err)
		}
		return d.handler.SendOrderCreatedEmail(event)

	case "order.cancelled":
		var event model.OrderCancelledEvent
		if err := json.Unmarshal(body, &event); err != nil {
			d.log.Error().
				Err(err).
				Str("event_type", eventType).
				Msg("Failed to unmarshal order.cancelled event")
			return fmt.Errorf("failed to unmarshal order.cancelled: %w", err)
		}
		return d.handler.SendOrderCancelledEmail(event)

	default:
		d.log.Warn().
			Str("event_type", eventType).
			Msg("Unknown event type received")
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}
