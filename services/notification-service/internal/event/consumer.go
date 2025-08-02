package event

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Consumer struct {
	ch         *amqp.Channel
	queue      string
	log        zerolog.Logger
	dispatcher *Dispatcher
}

func NewConsumer(ch *amqp.Channel, queue string, log zerolog.Logger, dispatcher *Dispatcher) *Consumer {
	return &Consumer{
		ch:         ch,
		queue:      queue,
		log:        log,
		dispatcher: dispatcher,
	}
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
		c.log.Error().Err(err).Msg("Failed to register consumer")
		return err
	}

	c.log.Info().Str("queue", c.queue).Msg("📥 Notification consumer started")

	go func() {
		for msg := range msgs {
			c.log.Info().
				Str("event_type", msg.Type).
				RawJSON("body", msg.Body).
				Msg("📨 Received message")

			if err := c.dispatcher.Dispatch(msg.Type, msg.Body); err != nil {
				c.log.Error().
					Err(err).
					Str("event_type", msg.Type).
					Msg("Dispatch failed")

				if err := msg.Nack(false, false); err != nil {
					c.log.Error().Err(err).Msg("Failed to NACK message")
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				c.log.Error().Err(err).Msg("Failed to ACK message")
			}
		}
	}()

	<-ctx.Done()
	c.log.Info().Msg("Consumer shutdown signal received")
	return nil
}
