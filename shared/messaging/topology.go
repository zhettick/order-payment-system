package messaging

import amqp "github.com/rabbitmq/amqp091-go"

const (
	PaymentExchange = "payments.events"
	PaymentDLX      = "payments.dlx"

	PaymentCompletedRoutingKey = "payment.completed"
	PaymentCompletedQueue      = "payment.completed"

	PaymentCompletedDLQ        = "payment.completed.dlq"
	PaymentCompletedDLQRouting = "payment.completed.dlq"
)

func DeclarePaymentTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		PaymentExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(
		PaymentDLX,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(
		PaymentCompletedDLQ,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := ch.QueueBind(
		PaymentCompletedDLQ,
		PaymentCompletedDLQRouting,
		PaymentDLX,
		false,
		nil,
	); err != nil {
		return err
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    PaymentDLX,
		"x-dead-letter-routing-key": PaymentCompletedDLQRouting,
	}

	if _, err := ch.QueueDeclare(
		PaymentCompletedQueue,
		true,
		false,
		false,
		false,
		args,
	); err != nil {
		return err
	}

	return ch.QueueBind(
		PaymentCompletedQueue,
		PaymentCompletedRoutingKey,
		PaymentExchange,
		false,
		nil,
	)
}
