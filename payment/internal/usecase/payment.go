package usecase

import (
	"context"
	"errors"
	"payment/events"
	"payment/internal/domain/entities"
	"payment/internal/domain/repository"
	"payment/internal/messaging"
	"time"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo      repository.PaymentRepository
	publisher messaging.EventPublisher
}

func NewPaymentUseCase(
	r repository.PaymentRepository,
	publisher messaging.EventPublisher,
) *PaymentUseCase {
	return &PaymentUseCase{
		repo:      r,
		publisher: publisher,
	}
}

func (u *PaymentUseCase) Process(orderID string, amount int64, customerEmail string) (*entities.Payment, error) {
	status := entities.StatusAuthorized
	if amount > 100000 {
		status = entities.StatusDeclined
	}

	payment := &entities.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		CustomerEmail: customerEmail,
		Amount:        amount,
		Status:        status,
	}

	err := u.repo.Create(payment)
	if err != nil {
		return payment, err
	}

	if status == entities.StatusAuthorized {
		event := events.PaymentCompletedEvent{
			EventID:       uuid.New().String(),
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: payment.CustomerEmail,
			Status:        payment.Status,
			OccurredAt:    time.Now().UTC(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := u.publisher.PublishPaymentCompleted(ctx, event); err != nil {
			return payment, err
		}
	}

	return payment, nil
}

func (u *PaymentUseCase) GetByID(orderID string) (*entities.Payment, error) {
	return u.repo.GetByID(orderID)
}

func (u *PaymentUseCase) ListPayments(min, max int64) ([]entities.Payment, error) {
	if min >= 0 && max >= 0 && min > max {
		return nil, errors.New("min cannot be greater than max")
	}

	return u.repo.FindByAmountRange(min, max)
}
