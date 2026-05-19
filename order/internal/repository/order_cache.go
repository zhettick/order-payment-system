package repository

import (
	"context"
	"order/internal/domain/entities"
)

type OrderCache interface {
	Get(ctx context.Context, id string) (*entities.Order, error)
	Set(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id string) error
}
