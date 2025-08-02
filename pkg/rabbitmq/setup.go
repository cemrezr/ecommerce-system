package rabbitmq

import (
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

func SetupOrderQueues(ch *amqp.Channel, exchange, queueName string, log zerolog.Logger) error {
	// DLX
	if err := ch.ExchangeDeclare("order.dlx", "direct", true, false, false, false, nil); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(
		"order.failed", true, false, false, false, nil,
	); err != nil {
		return err
	}

	if err := ch.QueueBind("order.failed", "order.failed", "order.dlx", false, nil); err != nil {
		return err
	}

	// Main exchange
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Queue with DLQ support
	args := amqp.Table{
		"x-dead-letter-exchange":    "order.dlx",
		"x-dead-letter-routing-key": "order.failed",
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
		return err
	}

	for _, key := range []string{"order.created", "order.cancelled"} {
		if err := ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
			return err
		}
	}

	log.Info().Str("queue", queueName).Str("exchange", exchange).Msg("✅ Queues, DLQ and bindings set up")
	return nil
}

func SetupBasicQueue(ch *amqp.Channel, exchange, queueName string, routingKeys []string, log zerolog.Logger) error {
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	for _, key := range routingKeys {
		if err := ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
			return err
		}
	}

	log.Info().Str("queue", queueName).Str("exchange", exchange).Msg("✅ Basic queue setup complete")
	return nil
}
