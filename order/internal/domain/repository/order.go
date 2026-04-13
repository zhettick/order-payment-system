package repository

import "order/internal/domain/entities"

type OrderRepository interface {
	Create(order *entities.Order) error
	GetByID(id string) (*entities.Order, error)
	Update(order *entities.Order) error
	GetRecent(limit int) ([]entities.Order, error)
}
