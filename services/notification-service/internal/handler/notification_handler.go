package handler

import (
	"github.com/cemrezr/ecommerce-system/notification-service/internal/model"
	"github.com/rs/zerolog"
)

type NotificationHandler struct {
	log zerolog.Logger
}

func NewNotificationHandler(log zerolog.Logger) *NotificationHandler {
	return &NotificationHandler{log: log}
}

func (h *NotificationHandler) SendOrderCreatedEmail(event model.OrderCreatedEvent) error {
	h.log.Info().
		Int("user_id", event.UserID).
		Int("order_id", event.ID).
		Msg("Order creation email sent to user")
	return nil
}

func (h *NotificationHandler) SendOrderCancelledEmail(event model.OrderCancelledEvent) error {
	h.log.Info().
		Int("user_id", event.UserID).
		Int("order_id", event.ID).
		Msg("Order cancellation email sent to user")
	return nil
}
