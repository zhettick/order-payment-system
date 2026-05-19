package provider

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

type SimulatedEmailSender struct {
	failureRate float64
	minDelay    time.Duration
	maxDelay    time.Duration
}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{
		failureRate: 0.3,
		minDelay:    500 * time.Millisecond,
		maxDelay:    1500 * time.Millisecond,
	}
}

func (s *SimulatedEmailSender) Send(ctx context.Context, message EmailMessage) error {
	delay := s.randomDelay()

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	if rand.Float64() < s.failureRate {
		return errors.New("simulated provider temporary failure")
	}

	log.Printf(
		"[Notification] Sent email to %s for Order #%s. Amount: $%d",
		message.To,
		message.OrderID,
		message.Amount,
	)

	return nil
}

func (s *SimulatedEmailSender) randomDelay() time.Duration {
	if s.maxDelay <= s.minDelay {
		return s.minDelay
	}

	diff := s.maxDelay - s.minDelay
	return s.minDelay + time.Duration(rand.Int63n(int64(diff)))
}
