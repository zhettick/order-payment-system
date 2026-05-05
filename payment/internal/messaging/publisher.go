package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"payment/events"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisher interface {
	PublishPaymentCompleted(ctx context.Context, event events.PaymentCompletedEvent) error
	Close() error
}

type RabbitPublisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms <-chan amqp.Confirmation
}

func NewRabbitPublisher(amqpURL string) (*RabbitPublisher, error) {
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

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &RabbitPublisher{
		conn:     conn,
		ch:       ch,
		confirms: confirms,
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

func (p *RabbitPublisher) PublishPaymentCompleted(
	ctx context.Context,
	event events.PaymentCompletedEvent,
) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment completed event: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		PaymentExchange,
		PaymentCompletedRoutingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			MessageId:    event.EventID,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish payment completed event: %w", err)
	}

	select {
	case confirm := <-p.confirms:
		if !confirm.Ack {
			return fmt.Errorf("RabbitMQ did not confirm published message")
		}
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RabbitPublisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}

	if p.conn != nil {
		return p.conn.Close()
	}

	return nil
}
