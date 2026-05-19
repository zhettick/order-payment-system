package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order/internal/domain/entities"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewOrderCache(client *goredis.Client, ttl time.Duration) *OrderCache {
	return &OrderCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *OrderCache) Get(ctx context.Context, id string) (*entities.Order, error) {
	key := c.key(id)

	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var order entities.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (c *OrderCache) Set(ctx context.Context, order *entities.Order) error {
	key := c.key(order.ID)

	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *OrderCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, c.key(id)).Err()
}

func (c *OrderCache) key(id string) string {
	return fmt.Sprintf("order:%s", id)
}
