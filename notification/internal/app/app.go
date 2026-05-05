package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification/internal/idempotency"
	"notification/internal/messaging"
)

func Run() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	store := idempotency.NewStore()

	consumer, err := messaging.NewConsumer(rabbitURL, store)
	if err != nil {
		log.Fatalf("failed to create notification consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("[Notification] Service started. Waiting for payment.completed events...")

	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("notification consumer stopped with error: %v", err)
	}

	log.Println("[Notification] Graceful shutdown completed")
}
