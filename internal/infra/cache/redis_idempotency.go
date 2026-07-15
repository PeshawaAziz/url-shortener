package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisIdempotencyStore(client *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: client,
		ttl:    24 * time.Hour,
	}
}

func (r *RedisIdempotencyStore) Get(ctx context.Context, key string) (uuid.UUID, error) {
	if key == "" {
		return uuid.Nil, errors.New("empty key")
	}

	redisKey := fmt.Sprintf("idem:%s", key)
	val, err := r.client.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.Nil, errors.New("not found")
	}
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(val)
}

func (r *RedisIdempotencyStore) Save(ctx context.Context, key string, urlID uuid.UUID) error {
	if key == "" {
		return nil
	}

	redisKey := fmt.Sprintf("idem:%s", key)
	return r.client.SetNX(ctx, redisKey, urlID.String(), r.ttl).Err()
}
