package rabbitmq

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

func Connect(url string, log zerolog.Logger) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to RabbitMQ")
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open RabbitMQ channel")
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	log.Info().Str("component", "rabbitmq").Msg("onnected to RabbitMQ")
	return conn, ch, nil
}
