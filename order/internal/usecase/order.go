package usecase

import (
	"errors"
	"order/internal/domain/entities"
	"order/internal/domain/repository"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PaymentGateway interface {
	Authorize(orderID string, amount int64) (string, error)
}

type OrderUseCase struct {
	repo          repository.OrderRepository
	paymentClient PaymentGateway
	subscribers   map[string][]chan string
	mu            sync.RWMutex
}

func NewOrderUseCase(r repository.OrderRepository, p PaymentGateway) *OrderUseCase {
	return &OrderUseCase{
		repo:          r,
		paymentClient: p,
		subscribers:   make(map[string][]chan string),
	}
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
	if err != nil {
		order.Status = entities.StatusFailed
		if updateErr := u.repo.Update(order); updateErr != nil {
			return nil, updateErr
		}
		u.notifySubscribers(order.ID, order.Status)
		return nil, errors.New("payment service unavailable")
	}

	if status != entities.StatusPaid {
		order.Status = entities.StatusFailed
	} else {
		order.Status = entities.StatusPaid
	}

	if err := u.repo.Update(order); err != nil {
		return nil, err
	}

	u.notifySubscribers(order.ID, order.Status)

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

	if err := u.repo.Update(order); err != nil {
		return err
	}

	u.notifySubscribers(id, order.Status)
	return nil
}

func (u *OrderUseCase) GetRecent(limit int) ([]entities.Order, error) {
	if limit <= 0 || limit > 100 {
		return nil, errors.New("invalid limit")
	}
	return u.repo.GetRecent(limit)
}

func (u *OrderUseCase) Subscribe(orderID string) chan string {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch := make(chan string, 1)
	u.subscribers[orderID] = append(u.subscribers[orderID], ch)
	return ch
}

func (u *OrderUseCase) Unsubscribe(orderID string, ch chan string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	subs := u.subscribers[orderID]

	for i, sub := range subs {
		if sub == ch {
			u.subscribers[orderID] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}

	if len(u.subscribers[orderID]) == 0 {
		delete(u.subscribers, orderID)
	}
}

func (u *OrderUseCase) notifySubscribers(orderID string, status string) {
	u.mu.RLock()
	subs := append([]chan string(nil), u.subscribers[orderID]...)
	u.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- status:
		default:
		}
	}
}
