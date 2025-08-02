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
	return &Consumer{ch: ch, queue: queue, log: log, dispatcher: dispatcher}
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
		c.log.Error().Err(err).Msg("Failed to start consuming messages")
		return err
	}

	c.log.Info().Str("queue", c.queue).Msg("Consumer started")

	go func() {
		for msg := range msgs {
			c.log.Debug().
				Str("type", msg.Type).
				Msg("Received message")

			err := c.dispatcher.Dispatch(msg.Type, msg.Body)
			if err != nil {
				c.log.Error().Err(err).Str("type", msg.Type).Msg("Failed to process event — NACKing")
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
			c.log.Info().Str("type", msg.Type).Msg("✅ Event processed and ACKed")
		}
	}()

	<-ctx.Done()
	c.log.Info().Msg("Consumer shutting down")
	return nil
}
