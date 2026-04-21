package repository

import "payment/internal/domain/entities"

type PaymentRepository interface {
	Create(p *entities.Payment) error
	GetByID(orderID string) (*entities.Payment, error)
	FindByAmountRange(min, max int64) ([]entities.Payment, error)
}
