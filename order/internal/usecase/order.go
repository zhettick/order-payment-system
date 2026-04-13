package usecase

import (
	"errors"
	"order/internal/domain/entities"
	"order/internal/domain/repository"
	"order/internal/transport/http/client"
	"time"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo          repository.OrderRepository
	paymentClient *client.PaymentClient
}

func NewOrderUseCase(r repository.OrderRepository, p *client.PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: r, paymentClient: p}
}

func (u *OrderUseCase) Create(customerID, itemName string, amount int64) (*entities.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	order := &entities.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     entities.StatusPending,
		CreatedAt:  time.Now(),
	}

	if err := u.repo.Create(order); err != nil {
		return nil, err
	}

	status, err := u.paymentClient.Authorize(order.ID, order.Amount)
	if err != nil || status != entities.StatusPaid {
		order.Status = entities.StatusFailed
	} else {
		order.Status = entities.StatusPaid
	}

	u.repo.Update(order)

	if err != nil {
		return nil, errors.New("payment service unavailable")
	}
	return order, nil
}

func (u *OrderUseCase) GetByID(id string) (*entities.Order, error) {
	return u.repo.GetByID(id)
}

func (u *OrderUseCase) Cancel(id string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != entities.StatusPending {
		return errors.New("only pending orders can be cancelled")
	}
	order.Status = entities.StatusCancelled
	return u.repo.Update(order)
}

func (u *OrderUseCase) GetRecent(limit int) ([]entities.Order, error) {
	if limit <= 0 || limit > 100 {
		return nil, errors.New("invalid limit, it should be between 1 and 100")
	}

	orders, err := u.repo.GetRecent(limit)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
