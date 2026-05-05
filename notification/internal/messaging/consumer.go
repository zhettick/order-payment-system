package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"notification/events"
	"notification/internal/idempotency"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	store *idempotency.Store
}

func NewConsumer(amqpURL string, store *idempotency.Store) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}

	if err := DeclarePaymentTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare RabbitMQ topology: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set consumer QoS: %w", err)
	}

	return &Consumer{
		conn:  conn,
		ch:    ch,
		store: store,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	const consumerTag = "notification-service"

	deliveries, err := c.ch.Consume(
		PaymentCompletedQueue,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("start consuming messages: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			_ = c.ch.Cancel(consumerTag, false)
			return ctx.Err()

		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			c.handleDelivery(delivery)
		}
	}
}

func (c *Consumer) handleDelivery(delivery amqp.Delivery) {
	var event events.PaymentCompletedEvent

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.Printf("[Notification] Invalid message moved to DLQ: %v", err)
		_ = delivery.Nack(false, false)
		return
	}

	if event.EventID == "" {
		log.Println("[Notification] Message without event_id moved to DLQ")
		_ = delivery.Nack(false, false)
		return
	}

	if !c.store.TryMarkProcessed(event.EventID) {
		log.Printf("[Notification] Duplicate event skipped: %s", event.EventID)
		_ = delivery.Ack(false)
		return
	}

	log.Printf(
		"[Notification] Sent email to %s for Order #%s. Amount: $%d",
		event.CustomerEmail,
		event.OrderID,
		event.Amount,
	)

	_ = delivery.Ack(false)
}

func (c *Consumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
