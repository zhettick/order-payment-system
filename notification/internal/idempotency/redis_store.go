package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewRedisStore(client *goredis.Client, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client: client,
		ttl:    ttl,
	}
}

func (s *RedisStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	key := s.key(eventID)

	_, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *RedisStore) MarkProcessed(ctx context.Context, eventID string) error {
	return s.client.Set(ctx, s.key(eventID), "processed", s.ttl).Err()
}

func (s *RedisStore) key(eventID string) string {
	return fmt.Sprintf("notification:processed:%s", eventID)
}
