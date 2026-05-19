package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification/events"
	"notification/internal/idempotency"
	"notification/internal/provider"
	"notification/internal/retry"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn        *amqp.Connection
	ch          *amqp.Channel
	store       *idempotency.RedisStore
	emailSender provider.EmailSender
	maxRetries  int
	baseBackoff time.Duration
}

func NewConsumer(
	amqpURL string,
	store *idempotency.RedisStore,
	emailSender provider.EmailSender,
	maxRetries int,
	baseBackoff time.Duration,
) (*Consumer, error) {
	conn, err := dialWithRetry(amqpURL, 10, 2*time.Second)
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
		conn:        conn,
		ch:          ch,
		store:       store,
		emailSender: emailSender,
		maxRetries:  maxRetries,
		baseBackoff: baseBackoff,
	}, nil
}

func dialWithRetry(amqpURL string, attempts int, delay time.Duration) (*amqp.Connection, error) {
	var lastErr error

	for i := 1; i <= attempts; i++ {
		conn, err := amqp.Dial(amqpURL)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		time.Sleep(delay)
	}

	return nil, lastErr
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

			c.handleDelivery(ctx, delivery)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) {
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

	processed, err := c.store.IsProcessed(ctx, event.EventID)
	if err != nil {
		log.Printf("[Notification] Failed to check idempotency, will retry later: %v", err)
		_ = delivery.Nack(false, true)
		return
	}

	if processed {
		log.Printf("[Notification] Duplicate event skipped: %s", event.EventID)
		_ = delivery.Ack(false)
		return
	}

	message := provider.EmailMessage{
		EventID: event.EventID,
		To:      event.CustomerEmail,
		OrderID: event.OrderID,
		Amount:  event.Amount,
		Status:  event.Status,
	}

	err = retry.Do(ctx, c.maxRetries, c.baseBackoff, func() error {
		sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return c.emailSender.Send(sendCtx, message)
	})
	if err != nil {
		log.Printf("[Notification] Provider failed after retries, moved to DLQ: %v", err)
		_ = delivery.Nack(false, false)
		return
	}

	if err := c.store.MarkProcessed(ctx, event.EventID); err != nil {
		log.Printf("[Notification] Email sent but failed to mark processed: %v", err)
		_ = delivery.Nack(false, true)
		return
	}

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
