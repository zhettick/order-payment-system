package usecase

import (
	"payment/internal/domain/entities"
	"payment/internal/domain/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo repository.PaymentRepository
}

func NewPaymentUseCase(r repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
}

func (u *PaymentUseCase) Process(orderID string, amount int64) (*entities.Payment, error) {
	status := entities.StatusAuthorized
	if amount > 100000 {
		status = entities.StatusDeclined
	}

	payment := &entities.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	err := u.repo.Create(payment)
	return payment, err
}

func (u *PaymentUseCase) GetByID(orderID string) (*entities.Payment, error) {
	return u.repo.GetByID(orderID)
}
