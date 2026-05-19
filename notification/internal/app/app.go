package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"notification/internal/idempotency"
	"notification/internal/messaging"
	"notification/internal/provider"

	goredis "github.com/redis/go-redis/v9"
)

func Run() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	providerMode := os.Getenv("PROVIDER_MODE")
	if providerMode == "" {
		providerMode = "SIMULATED"
	}

	maxRetries := getEnvInt("NOTIFICATION_MAX_RETRIES", 3)
	baseBackoff := getEnvDuration("NOTIFICATION_BASE_BACKOFF", 2*time.Second)
	idempotencyTTL := getEnvDuration("NOTIFICATION_IDEMPOTENCY_TTL", 24*time.Hour)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	store := idempotency.NewRedisStore(redisClient, idempotencyTTL)

	emailSender := buildEmailSender(providerMode)

	consumer, err := messaging.NewConsumer(
		rabbitURL,
		store,
		emailSender,
		maxRetries,
		baseBackoff,
	)
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

func buildEmailSender(mode string) provider.EmailSender {
	switch mode {
	case "SIMULATED":
		return provider.NewSimulatedEmailSender()
	default:
		log.Printf("unknown PROVIDER_MODE=%s, using SIMULATED", mode)
		return provider.NewSimulatedEmailSender()
	}
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
